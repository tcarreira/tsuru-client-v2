package printer

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testingPrintableType() PrintableType {
	return PrintableType{
		SimpleFields: []FieldType{
			{Name: "name", Value: "myapp"},
			{Name: "description", Value: "This is my app"},
			{Name: "teamOwner", Value: "myteam"},
		},
		DetailedFields: []DetailedFieldType{
			{
				Name:   "units",
				Fields: []string{"id", "status"},
				Items: []ArrayItemType{
					{"unit1", "started"},
					{"unit2", "stopped"},
				},
			},
			{
				Name:   "service instances",
				Fields: []string{"service", "instance", "plan"},
				Items: []ArrayItemType{
					{"mysql", "mydb", "small"},
					{"redis", "mycache-instance", "medium"},
				},
			},
		},
	}
}

func TestPrintableType_Print(t *testing.T) {
	p := testingPrintableType()
	out := bytes.Buffer{}
	p.PrintTable(&out)

	expected := `
name:         myapp
description:  This is my app
teamOwner:    myteam

units:
  ID     STATUS
  unit1  started
  unit2  stopped

service instances:
  SERVICE  INSTANCE          PLAN
  mysql    mydb              small
  redis    mycache-instance  medium
`
	assert.Equal(t, expected, "\n"+out.String())
}
