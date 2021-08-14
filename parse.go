package gofire

import (
	"bytes"
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
)

func Parse(ctx context.Context, r io.Reader, name string) (*Command, error) {
	var buf bytes.Buffer
	_, err := buf.ReadFrom(r)
	if err != nil {
		return nil, err
	}
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", buf, parser.AllErrors)
	if err != nil {
		return nil, err
	}
	var cmd Command
	cmd.Package = f.Name.Name
	for _, decl := range f.Decls {
		fdecl, ok := decl.(*ast.FuncDecl)
		if ok && fdecl.Name.Name == name {
			cmd.Doc = fdecl.Doc.Text()
			for _, result := range fdecl.Type.Results.List {
				rawType := buf.Bytes()[int(result.Type.Pos()):int(result.Type.End())]
				n := 1
				if l := len(result.Names); l > 0 {
					n = l
				}
				for i := 0; i < n; i++ {
					cmd.Results = append(cmd.Results, string(rawType))
				}
			}
			var argi uint64
		params:
			for i, param := range fdecl.Type.Params.List {
				var typ Typ
				switch pt := param.Type.(type) {
				case *ast.SelectorExpr:
					pckg, ok := pt.X.(*ast.Ident)
					if i == 0 && ok && pckg.Name == "context" && pt.Sel.Name == "Context" {
						cmd.Context = true
					}
					continue params
				// TODO add parse of supported types.
				default:
					rawType := buf.Bytes()[int(param.Type.Pos()):int(param.Type.End())]
					return nil, fmt.Errorf("provided function %s type %s can't be parsed", name, rawType)
				}
				n := 1
				if l := len(param.Names); l > 0 {
					n = l
				}
				// TODO parse arguments and flags separately
				for i := 0; i < n; i++ {
					cmd.Parameters = append(cmd.Parameters, Argument{
						Index: uint64(argi),
						Type:  typ,
					})
					argi++
				}
			}
			return &cmd, nil
		}
	}
	return nil, fmt.Errorf("provided function %s can't be found", name)
}
