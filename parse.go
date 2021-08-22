package gofire

import (
	"bytes"
	"context"
	"fmt"
	"go/ast"
	goparser "go/parser"
	"go/token"
	"io/fs"
	"strconv"
	"strings"
)

// Parse tries to parse the function from provided ast into command type.
func Parse(ctx context.Context, dir fs.FS, pckg, function string) (*Command, error) {
	// Start with parsing actual ast from fs driver.
	fentries, err := fs.ReadDir(dir, ".")
	if err != nil {
		return nil, fmt.Errorf("ast package %s can't be read, %w", pckg, err)
	}
	var files []file
	fset := token.NewFileSet()
	for _, fentry := range fentries {
		fname := fentry.Name()
		if fentry.IsDir() || !strings.HasSuffix(fname, ".go") || strings.HasSuffix(fname, "_test.go") {
			continue
		}
		b, err := fs.ReadFile(dir, fname)
		if err != nil {
			return nil, fmt.Errorf("ast file %s in package %s can't be read, %w", fname, pckg, err)
		}
		buf := bytes.NewBuffer(b)
		f, err := goparser.ParseFile(fset, "", buf, goparser.AllErrors|goparser.ParseComments)
		if err != nil {
			return nil, fmt.Errorf("ast file %s in package %s can't be parsed, %w", fname, pckg, err)
		}
		if f.Name.Name != pckg {
			continue
		}
		files = append(files, file{fset: fset, fname: fname, ast: f, buf: buf})
	}
	// Now as ast is parsed successfully parse it into command.
	p := make(parser)
	var fparse func(context.Context) (*Command, error)
	for _, file := range files {
		// Visit all types inide the package to build flag groups.
		for _, decl := range file.ast.Decls {
			gdecl, ok := decl.(*ast.GenDecl)
			if ok {
				for _, spec := range gdecl.Specs {
					tspec, ok := spec.(*ast.TypeSpec)
					if ok {
						if err := p.register(file, gdecl, tspec); err != nil {
							return nil, fmt.Errorf(
								"ast file %s in package %s group type %s ast parsing error, %w",
								file.fname,
								pckg,
								tspec.Name.Name,
								err,
							)
						}
					}
				}
			}
			// In case we found function declaration that we need - save it,
			// we will process it later after the visit loop.
			file := file
			if fdecl, ok := decl.(*ast.FuncDecl); ok && fdecl.Name.Name == function {
				fparse = func(context.Context) (*Command, error) {
					var cmd Command
					cmd.Package = pckg
					cmd.Function = function
					cmd.Doc = strings.TrimSpace(fdecl.Doc.Text())
					cmd.Results = p.results(file, fdecl)
					params, context, err := p.parameters(file, fdecl)
					if err != nil {
						return nil, fmt.Errorf(
							"ast file %s in package %s function %s ast parsing error, %w",
							file.fname,
							pckg,
							function,
							err,
						)
					}
					cmd.Context = context
					cmd.Parameters = params
					return &cmd, nil
				}
			}
		}
	}
	if fparse == nil {
		return nil, fmt.Errorf("function %s can't be found in ast package %s", function, pckg)
	}
	return fparse(ctx)
}

type file struct {
	fset  *token.FileSet
	fname string
	ast   *ast.File
	buf   *bytes.Buffer
}

func (f file) definition(pos, end token.Pos) string {
	fpos, fend := f.fset.Position(pos), f.fset.Position(end)
	return f.buf.String()[fpos.Offset:fend.Offset]
}

type parser map[string]Group

func (p parser) results(f file, fdecl *ast.FuncDecl) (results []string) {
	var list []*ast.Field
	if fdecl.Type.Results != nil {
		list = fdecl.Type.Results.List
	}
	for _, result := range list {
		def := f.definition(result.Type.Pos(), result.Type.End())
		n := 1
		if l := len(result.Names); l > 0 {
			n = l
		}
		for i := 0; i < n; i++ {
			results = append(results, string(def))
		}
	}
	return
}

func (p *parser) parameters(f file, fdecl *ast.FuncDecl) (parameters []Parameter, context bool, err error) {
	var arg uint64
	var list []*ast.Field
	if fdecl.Type.Params != nil {
		list = fdecl.Type.Params.List
	}
	for i, param := range list {
		if i == 0 && p.context(param.Type) {
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
				"parameter %s type can't be parsed, %w",
				f.definition(param.Pos(), param.End()),
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

func (p *parser) register(f file, gendecl *ast.GenDecl, tspec *ast.TypeSpec) error {
	// In case it's not a structure type skip it.
	stype, ok := tspec.Type.(*ast.StructType)
	if !ok {
		return nil
	}
	var g Group
	g.Name = tspec.Name.Name
	if tspec.Doc != nil {
		g.Doc = strings.TrimSpace(tspec.Doc.Text())
	} else {
		g.Doc = strings.TrimSpace(gendecl.Doc.Text())
	}
	g.Type = TStruct{Typ: g.Name}
	for _, field := range stype.Fields.List {
		// In case no flag names provided we can skip the embedded field.
		if len(field.Names) == 0 {
			continue
		}
		typ, err := p.typ(field.Type)
		if err != nil {
			return fmt.Errorf(
				"field %s type can't be parsed, %w",
				f.definition(field.Pos(), field.End()),
				err,
			)
		}
		var tag string
		if field.Tag != nil {
			tag = field.Tag.Value
		}
		flag, set, err := p.tagflag(tag)
		if err != nil {
			return fmt.Errorf(
				"field %s tag can't be parsed, %w",
				f.definition(field.Pos(), field.End()),
				err,
			)
		}
		// Short flag names supported only for single name structure fields.
		if flag.Short != "" && len(field.Names) > 1 {
			return fmt.Errorf(
				"ambiguous short flag name %s for multiple fields %s",
				flag.Short,
				f.definition(field.Pos(), field.End()),
			)
		}
		flag.Doc = strings.TrimSpace(field.Doc.Text())
		flag.Type = typ
		for _, name := range field.Names {
			if name.Name == "_" {
				continue
			}
			flag.Full = name.Name
			// Fix broken default in case the flag wasn't set.
			if !set {
				flag.Optional = true
				flag.Default = flag.Type.Kind().Default()
			}
			g.Flags = append(g.Flags, *flag)
		}
	}
	(*p)[g.Type.Type()] = g
	return nil
}

func (p parser) context(tp ast.Expr) bool {
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
		g, ok := p[g.Name]
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
			return nil, fmt.Errorf("unsupported primitive type %s", k.Type())
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
			return nil, fmt.Errorf("unsupported array size literal type %s", lit.Kind.String())
		}
		size, err := strconv.ParseInt(lit.Value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("unsupported array size literal value %s", lit.Value)
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

func (p parser) tagflag(rawTag string) (*Flag, bool, error) {
	var f Flag
	// Skip empty tags they will be transformed into auto flags.
	if rawTag == "" {
		return &f, false, nil
	}
	rawTag = strings.Trim(rawTag, "`")
	tags := strings.Split(rawTag, " ")
	for _, ftag := range tags {
		parts := strings.Split(ftag, ":")
		if len(parts) != 2 || parts[0] != "gofire" {
			continue
		}
		// Skip omitted tags they will be transformed into auto flags.
		if parts[1] == `"-"` {
			return &f, false, nil
		}
		tags := strings.Split(strings.Trim(parts[1], `"`), ",")
		for _, tag := range tags {
			tv := strings.Split(tag, "=")
			ltv := len(tv)
			if ltv > 2 {
				return nil, false, fmt.Errorf("can't parse tag %s as key=value pair in %s", tag, rawTag)
			}
			// Validate key/values and parse the value.
			var val interface{}
			switch tv[0] {
			case "short", "default":
				if ltv != 2 {
					return nil, false, fmt.Errorf(
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
						return nil, false, fmt.Errorf(
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
				return nil, false, fmt.Errorf(
					"can't parse tag %s unsupported %q key in %s",
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
				f.Default = val.(string)
			case "optional":
				f.Optional = val.(bool)
			case "deprecated":
				f.Deprecated = val.(bool)
			case "hidden":
				f.Hidden = val.(bool)
			}
		}
		return &f, true, nil
	}
	return &f, false, nil
}
