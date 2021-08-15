package gofire

import (
	"bytes"
	"context"
	"fmt"
	"go/ast"
	goparser "go/parser"
	"go/token"
	"io"
	"strconv"
)

// Parse tries to parse the function from provided ast into command type.
func Parse(ctx context.Context, r io.Reader, function string) (*Command, error) {
	var buf bytes.Buffer
	_, err := buf.ReadFrom(r)
	if err != nil {
		return nil, fmt.Errorf("ast file can't be read %w", err)
	}
	p := parser{buf: buf}
	fset := token.NewFileSet()
	f, err := goparser.ParseFile(fset, "", buf, goparser.AllErrors)
	if err != nil {
		return nil, fmt.Errorf("ast file can't be parsed %w", err)
	}
	var cmd Command
	cmd.Package = f.Name.Name
	cmd.Function = function
	for _, decl := range f.Decls {
		fdecl, ok := decl.(*ast.FuncDecl)
		if ok && fdecl.Name.Name == function {
			cmd.Doc = fdecl.Doc.Text()
			cmd.Results = p.results(fdecl)
			cmd.Parameters, cmd.Context = p.parameters(fdecl)
			if p.err != nil {
				return nil, fmt.Errorf("function %s ast parsing error %w", function, p.err)
			}
			return &cmd, nil
		}
	}
	return nil, fmt.Errorf("function %s can't be found in ast", function)
}

type parser struct {
	buf    bytes.Buffer
	groups map[string]Group
	err    error
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

func (p *parser) parameters(fdecl *ast.FuncDecl) (parameters []Parameter, context bool) {
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
		typ, err := p.typ(param.Type)
		if err != nil {
			p.err = fmt.Errorf(
				"parameter %s type can't be parsed %w",
				p.rawType(param.Pos(), param.End()),
				err,
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

func (p parser) rawType(pos, end token.Pos) string {
	return p.buf.String()[int(pos):int(end)]
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
