package printer

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"reflect"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func getConstTypeEnumsFromFile(t *testing.T, filename, typeName string) []string {
	fset := token.NewFileSet() // positions are relative to fset
	f, err := parser.ParseFile(fset, filename, nil, 0)
	assert.NoError(t, err)

	typeEnums := []string{}
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.GenDecl:
			if x.Tok == token.CONST {
				for _, s := range x.Specs {
					switch s := s.(type) {
					case *ast.ValueSpec:
						switch dt := s.Type.(type) {
						case *ast.Ident:
							if dt.Name == typeName {
								for _, v := range s.Names {
									typeEnums = append(typeEnums, v.Name)
								}
							}
						}
					}
				}
			}
		}
		return true
	})
	return typeEnums
}

func TestPrintInfo_Switch(t *testing.T) {
	outputTypeEnums := getConstTypeEnumsFromFile(t, "printer.go", reflect.TypeOf(OutputType("")).Name())

	fset := token.NewFileSet() // positions are relative to fset
	f, err := parser.ParseFile(fset, "info.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	switchCaseNames := []string{}
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.CaseClause:
			if len(x.List) > 0 {
				if listElemAsIdent, ok := x.List[0].(*ast.Ident); ok {
					switchCaseNames = append(switchCaseNames, listElemAsIdent.Name)
				}
			}
		}
		return true
	})

	sort.Strings(outputTypeEnums)
	sort.Strings(switchCaseNames)
	assert.Equal(t, outputTypeEnums, switchCaseNames, "not all OutputType enums are covered in switch case at info.go")

	// for coverage
	for _, enum := range outputTypeEnums {
		format := OutputType(enum)
		err := PrintInfo(io.Discard, format, "")
		assert.NoError(t, err)
	}
	err = PrintInfo(io.Discard, "invalid", "")
	assert.NoError(t, err)
}
