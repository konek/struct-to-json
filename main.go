package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
)

func getType(node ast.Expr) string {
	if n, ok := node.(*ast.Ident); ok == true {
		return n.String()
	}
	if n, ok := node.(*ast.SelectorExpr); ok == true {
		return n.Sel.String()
	}
	if _, ok := node.(*ast.StructType); ok == true {
		return ""
	}
	if n, ok := node.(*ast.ArrayType); ok == true {
		return fmt.Sprintf("[]%s", n.Elt)
	}
	if n, ok := node.(*ast.StarExpr); ok == true {
		return fmt.Sprintf("*%s", n.X)
	}
	_ = node.(*ast.BinaryExpr)
	return ""
}

func LookupType(pkgs map[string]*ast.Package, name string) *ast.Object {
	for k := range pkgs {
		fset := token.NewFileSet()
		pkg, err := ast.NewPackage(fset, pkgs[k].Files, nil, nil)
		if err != nil {
			// panic(err) // too much errors due to the abscence of an importer (incomplete types)
		}
		return pkg.Scope.Lookup(name)
	}
	return nil
}

func export(_type *ast.StructType, name string) map[string]interface{} {
	_export := make(map[string]interface{})
	_export["$name"] = name
	_export["$type"] = "struct"
	fields := make(map[string]interface{})
	for i := range _type.Fields.List {
		field := _type.Fields.List[i]
		if __type, ok := field.Type.(*ast.StructType); ok == true {
			fields[field.Names[0].String()] = export(__type, "")
		} else if len(field.Names) != 0 {
			fields[field.Names[0].String()] = map[string]interface{}{
				"$name": getType(field.Type),
				"$type": "literal",
			}
		}
	}
	_export["$fields"] = fields
	return _export
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage:", os.Args[0], "<TypeName>")
		os.Exit(2)
	}
	symbol := os.Args[1]
	_export := make(map[string]interface{})
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, ".", nil, parser.DeclarationErrors|parser.AllErrors)
	if err != nil {
		panic(err)
	}
	_type := LookupType(pkgs, symbol)
	if _type != nil {
		_export = export(_type.Decl.(*ast.TypeSpec).Type.(*ast.StructType), symbol)
	}
	b, err := json.MarshalIndent(_export, "", "\t")
	if err == nil {
		fmt.Println(string(b))
	}
}
