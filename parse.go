package gofire

import (
	"bytes"
	"context"
	"fmt"
	"go/ast"
	goparser "go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Parse tries to parse the function from provided ast into command type.
func Parse(ctx context.Context, pckg, function string) (*Command, error) {
	fset := token.NewFileSet()
	dir, err := goparser.ParseDir(fset, pckg, func(fi fs.FileInfo) bool {
		return strings.HasSuffix(fi.Name(), "_test.go")
	}, goparser.AllErrors)
	if err != nil {
		return nil, fmt.Errorf("ast package can't be parsed %w", err)
	}
	_, last := filepath.Split(pckg)
	pkg, ok := dir[last]
	if !ok {
		return nil, fmt.Errorf("ast package %s wasn't found in provided path %s", last, pckg)
	}
	p := parser{
		pckg:    pckg,
		fset:    fset,
		buffers: make(map[string]*bytes.Buffer),
		groups:  make(map[string]Group),
	}
	var fdecl *ast.FuncDecl
	for _, file := range pkg.Files {
		// Visit all files inide the package to gather raw file buffers.
		if err := p.regFile(file); err != nil {
			return nil, fmt.Errorf("file %s ast parsing error %w", file.Name.Name, err)
		}
		// Visit all types inide the package to build flag groups.
		for _, decl := range file.Decls {
			gdecl, ok := decl.(*ast.GenDecl)
			if ok {
				for _, spec := range gdecl.Specs {
					tspec, ok := spec.(*ast.TypeSpec)
					if ok {
						if err := p.regType(tspec); err != nil {
							return nil, fmt.Errorf("group type %s ast parsing error %w", tspec.Name.Name, err)
						}
					}
				}
			}
			// In case we found function declaration that we need - save it,
			// we will process it later after the visit loop.
			if lfdecl, ok := decl.(*ast.FuncDecl); ok && lfdecl.Name.Name == function {
				fdecl = lfdecl
			}
		}
	}
	if fdecl == nil {
		return nil, fmt.Errorf("function %s can't be found in ast", function)
	}
	// Than in case the function was found process its declaration.
	var cmd Command
	cmd.Package = pkg.Name
	cmd.Function = function
	cmd.Doc = fdecl.Doc.Text()
	cmd.Results = p.results(fdecl)
	params, context, err := p.parameters(fdecl)
	if err != nil {
		return nil, fmt.Errorf("function %s ast parsing error %w", function, err)
	}
	cmd.Context = context
	cmd.Parameters = params
	return &cmd, nil

}

type parser struct {
	pckg    string
	fset    *token.FileSet
	buffers map[string]*bytes.Buffer
	groups  map[string]Group
}

func (p parser) results(fdecl *ast.FuncDecl) (results []string) {
	for _, result := range fdecl.Type.Results.List {
		rawType := p.rawType(result.Type.Pos(), result.Type.End())
		n := 1
		if l := len(result.Names); l > 0 {
			n = l
		}
		for i := 0; i < n; i++ {
			results = append(results, string(rawType))
		}
	}
	return
}

func (p *parser) parameters(fdecl *ast.FuncDecl) (parameters []Parameter, context bool, err error) {
	var arg uint64
	for i, param := range fdecl.Type.Params.List {
		if i == 0 && p.isContext(param.Type) {
			context = true
			continue
		}
		n := len(param.Names)
		// Try to parse parameter as one of flag groups first.
		g, ok := p.group(param.Type)
		if ok {
			// Check if we need just a type placeholder instead of rich parameter.
			if n == 0 {
				parameters = append(parameters, Placeholder{Type: g.Type})
				continue
			}
			for i := 0; i < n; i++ {
				name := param.Names[i].Name
				// Check if we need just a type placeholder instead of rich parameter.
				if name == "_" {
					parameters = append(parameters, Placeholder{Type: g.Type})
					continue
				}
				group := *g
				group.Name = param.Names[i].Name
				parameters = append(parameters, group)
			}
			continue
		}
		// If parameter is not a group parse its type.
		typ, terr := p.typ(param.Type)
		if terr != nil {
			err = fmt.Errorf(
				"parameter %s type can't be parsed %w",
				p.rawType(param.Pos(), param.End()),
				terr,
			)
			return
		}
		// Check if we need just a type placeholder instead of rich parameter.
		if n == 0 {
			parameters = append(parameters, Placeholder{Type: typ})
			continue
		}
		for i := 0; i < n; i++ {
			name := param.Names[i].Name
			// Check if we need just a type placeholder instead of rich parameter.
			if name == "_" {
				parameters = append(parameters, Placeholder{Type: typ})
				continue
			}
			// In case type of parameter is pointer we define it as autoflag.
			if ptr, ok := typ.(TPtr); ok {
				parameters = append(parameters, Flag{
					Full:     name,
					Optional: true,
					Default:  ptr.ETyp.Kind().Default(),
					Type:     typ,
				})
				continue
			}
			// Otherwise parameter is positional argument.
			parameters = append(parameters, Argument{
				Index: uint64(arg),
				Type:  typ,
			})
			arg++
		}
	}
	return
}

func (p *parser) regFile(f *ast.File) error {
	full := filepath.Join(p.pckg, f.Name.Name)
	fr, err := os.Open(full)
	if err != nil {
		return fmt.Errorf("error happen on file %s open %w", full, err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(fr)
	if err != nil {
		return fmt.Errorf("error happen on file %s read %w", full, err)
	}
	p.buffers[f.Name.Name] = &buf
	return nil
}

func (p *parser) regType(tspec *ast.TypeSpec) error {
	// In case it's not a structure type skip it.
	stype, ok := tspec.Type.(*ast.StructType)
	if !ok {
		return nil
	}
	var g Group
	g.Doc = tspec.Doc.Text()
	g.Type = TStruct{Typ: g.Name}
	for _, f := range stype.Fields.List {
		// In case no flag names provided we can skip the field.
		if len(f.Names) == 0 {
			continue
		}
		typ, err := p.typ(f.Type)
		if err != nil {
			return err
		}
		flag, err := p.tagFlag(f.Tag.Value)
		if err != nil {
			return err
		}
		// Short flag names supported only for single name structure fields.
		if flag.Short != "" && len(f.Names) > 1 {
			return fmt.Errorf(
				"ambiguous short flag name %s for multiple fields %s",
				flag.Short,
				p.rawType(f.Pos(), f.End()),
			)
		}
		flag.Doc = f.Doc.Text()
		flag.Type = typ
		for _, name := range f.Names {
			if name.Name == "-" {
				continue
			}
			flag.Full = name.Name
			// Fix broken default in case it wasn't set.
			if flag.Default == "" {
				flag.Default = flag.Type.Kind().Default()
			}
			g.Flags = append(g.Flags, *flag)
		}
	}
	p.groups[g.Type.Type()] = g
	return nil
}

func (p parser) rawType(pos, end token.Pos) string {
	fpos, fend := p.fset.Position(pos), p.fset.Position(end)
	buf := p.buffers[fpos.Filename]
	return buf.String()[fpos.Offset:fend.Offset]
}

func (p parser) isContext(tp ast.Expr) bool {
	sel, ok := tp.(*ast.SelectorExpr)
	if ok && sel.Sel.Name == "Context" {
		ctx, ok := sel.X.(*ast.Ident)
		return ok && ctx.Name == "context"
	}
	return false
}

func (p parser) group(tp ast.Expr) (*Group, bool) {
	g, ok := tp.(*ast.Ident)
	if ok {
		g, ok := p.groups[g.Name]
		return &g, ok
	}
	return nil, false
}

func (p parser) typ(tp ast.Expr) (Typ, error) {
	switch tt := tp.(type) {
	case *ast.Ident:
		var k Kind
		switch tt.Name {
		case Bool.Type():
			k = Bool
		case Int.Type():
			k = Int
		case Int8.Type():
			k = Int8
		case Int16.Type():
			k = Int16
		case Int32.Type():
			k = Int32
		case Int64.Type():
			k = Int64
		case Uint.Type():
			k = Uint
		case Uint8.Type():
			k = Uint8
		case Uint16.Type():
			k = Uint16
		case Uint32.Type():
			k = Uint32
		case Uint64.Type():
			k = Uint64
		case Float32.Type():
			k = Float32
		case Float64.Type():
			k = Float64
		case Complex64.Type():
			k = Complex64
		case Complex128.Type():
			k = Complex128
		case String.Type():
			k = String
		default:
			return nil, fmt.Errorf("unsupported primitive type")
		}
		return TPrimitive{TKind: k}, nil
	case *ast.ArrayType:
		etyp, err := p.typ(tt.Elt)
		if err != nil {
			return nil, err
		}
		if tt.Len == nil {
			return TSlice{ETyp: etyp}, nil
		}
		lit, ok := tt.Len.(*ast.BasicLit)
		if !ok || lit.Kind != token.INT {
			return nil, fmt.Errorf("unsupported array size literal type")
		}
		size, err := strconv.ParseInt(lit.Value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("unsupported array size literal value %w", err)
		}
		return TArray{ETyp: etyp, Size: size}, nil
	case *ast.MapType:
		ktyp, err := p.typ(tt.Key)
		if err != nil {
			return nil, err
		}
		vtyp, err := p.typ(tt.Value)
		if err != nil {
			return nil, err
		}
		return TMap{KTyp: ktyp, VTyp: vtyp}, nil
	case *ast.StarExpr:
		etyp, err := p.typ(tt.X)
		if err != nil {
			return nil, err
		}
		return TPtr{ETyp: etyp}, nil
	default:
		return nil, fmt.Errorf("unsupported complex type")
	}
}

func (p parser) tagFlag(rawTag string) (*Flag, error) {
	var f Flag
	// Skip empty tags they will be transformed into auto flags.
	if rawTag == "" {
		return &f, nil
	}
	tags := strings.Split(rawTag, " ")
	for _, ftag := range tags {
		parts := strings.Split(ftag, ":")
		if len(parts) != 2 || parts[0] != "gofire" {
			continue
		}
		// Skip omitted tags they will be transformed into auto flags.
		if parts[1] == `"-"` {
			return &f, nil
		}
		tags := strings.Split(strings.Trim(parts[1], `"`), ",")
		for _, tag := range tags {
			tv := strings.Split(tag, "=")
			ltv := len(tv)
			if ltv > 2 {
				return nil, fmt.Errorf("can't parse tag %s as key=value pair in %s", tag, rawTag)
			}
			// Validate key/values and parse the value.
			var val interface{}
			switch tv[0] {
			case "short", "default":
				if ltv != 2 {
					return nil, fmt.Errorf(
						"can't parse tag %s missing %q key value in %s",
						tag,
						tv[0],
						rawTag,
					)

				}
				val = tv[1]
			case "optional", "deprecated", "hidden":
				if ltv == 1 {
					val = true
				} else {
					valb, err := strconv.ParseBool(tv[1])
					if err != nil {
						return nil, fmt.Errorf(
							"can't parse tag %s as boolean for %q key and %s value in %s",
							tag,
							tv[0],
							tv[1],
							rawTag,
						)
					}
					val = valb
				}
			default:
				return nil, fmt.Errorf(
					"can't parse tag %s unknown %q key in %s",
					tag,
					tv[0],
					rawTag,
				)
			}
			// Fill key/values after the validation and parsing.
			switch tv[0] {
			case "short":
				f.Short = val.(string)
			case "default":
				f.Short = val.(string)
			case "optional":
				f.Optional = val.(bool)
			case "deprecated":
				f.Deprecated = val.(bool)
			case "hidden":
				f.Hidden = val.(bool)
			}
		}
	}
	return &f, nil
}
