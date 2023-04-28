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
		ListField: []ListType{
			{
				Name:    "units",
				Headers: []string{"id", "status"},
				Items: []ItemType{
					{"unit1", "started"},
					{"unit2", "stopped"},
				},
			},
			{
				Name:    "service instances",
				Headers: []string{"service", "instance", "plan"},
				Items: []ItemType{
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
