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
	"strings"

	"github.com/codegangsta/cli"
)

func Generate(c *cli.Context) {
	var f interface{}
	dec := json.NewDecoder(os.Stdin)
	if err := dec.Decode(&f); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	var m map[string]interface{}

	t := reflect.TypeOf(f)
	switch t.Kind() {
	case reflect.Map:
		m = f.(map[string]interface{})
	case reflect.Slice:
		m = (f.([]interface{}))[0].(map[string]interface{})
	default:
		log.Fatal(t.Kind)
		os.Exit(1)
	}

	ch := make(chan ast.Spec)
	go func() {
		NewType(ch, c, c.String("name"), m)
		close(ch)
	}()

	var types []ast.Decl
	for spec := range ch {
		types = append(types, &ast.GenDecl{
			Tok: token.TYPE,
			Specs: []ast.Spec{
				spec,
			},
		})
	}

	file := &ast.File{
		Name: ast.NewIdent(c.String("package")),
		Decls: types,
	}

	printer.Fprint(os.Stdout, token.NewFileSet(), file)
}

func NewType(ch chan ast.Spec, c *cli.Context, name string, m map[string]interface{}) {
	var fields []*ast.Field

	mk := make([]string, len(m))
	i := 0
	for k, _ := range m {
		mk[i] = k
		i++
	}
	sort.Strings(mk)

	addTag := ""
	if c.Bool("omitempty") {
		addTag = ",omitempty"
	}

	for _, k := range mk {
		ts := "string"
		v := m[k]

		t := reflect.TypeOf(v)
		if t != nil {
			switch t.Kind() {
			case reflect.Map:
				tName := PascalCase(k)
				ts = strings.Join([]string{"*", tName}, "")
				NewType(ch, c, tName, v.(map[string]interface{}))
			case reflect.Slice:
				log.Print("slice", k, t)
			default:
				ts = t.String()
			}
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
				Value:    fmt.Sprintf("`json:\"%s%s\"`", k, addTag),
			},
		})
	}
	spec := &ast.TypeSpec{
		Name: ast.NewIdent(name),
		Type: &ast.StructType{
			Fields: &ast.FieldList{
				List: fields,
			},
		},
	}
	ch <- spec
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
