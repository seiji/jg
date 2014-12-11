package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"log"
	"os"
	"reflect"
	"regexp"
	"sort"

	"github.com/codegangsta/cli"
)

func Generate(c *cli.Context) {
	var f interface{}

	dec := json.NewDecoder(os.Stdin)
	if err := dec.Decode(&f); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	m := f.(map[string]interface{})
	mk := make([]string, len(m))
	i := 0
	for k, _ := range m {
		mk[i] = k
		i++
	}
	sort.Strings(mk)

	var fields []*ast.Field
	for _, k := range mk {
		ts := "string"
		v := m[k]
		if v != nil {
			ts = reflect.TypeOf(v).String()
		}

		fields = append(fields, &ast.Field{
			Names: []*ast.Ident{
				&ast.Ident{
					Name:    PascalCase(k),
					NamePos: token.NoPos,
					Obj:     ast.NewObj(ast.Var, k),
				},
			},
			Type: ast.NewIdent(ts),
			Tag: &ast.BasicLit{
				ValuePos: token.NoPos,
				Kind:     token.STRING,
				Value:    fmt.Sprintf("`json:\"%s,omitempty\"`", k),
			},
		})
	}

	types := []ast.Spec{
		&ast.TypeSpec{
			Name: ast.NewIdent(c.String("name")),
			Type: &ast.StructType{
				Fields: &ast.FieldList{
					List: fields,
				},
			},
		},
	}

	file := &ast.File{
		Name: ast.NewIdent("main"),
		Decls: []ast.Decl{
			&ast.GenDecl{
				Tok:   token.TYPE,
				Specs: types,
			},
		},
	}

	printer.Fprint(os.Stdout, token.NewFileSet(), file)
}

var re = regexp.MustCompile("[0-9A-Za-z]+")

func PascalCase(s string) string {
	b := []byte(s)
	values := re.FindAll(b, -1)
	for i, v := range values {
		values[i] = bytes.Title(v)
	}
	return string(bytes.Join(values, nil))
}
