package main

import (
	"github.com/actgardner/gogen-avro/v10/parser"
	"github.com/actgardner/gogen-avro/v10/schema"
	"testing"

	avro "github.com/actgardner/gogen-avro/v10/schema"
	qt "github.com/frankban/quicktest"
)

var eventNameQualifiedName = schema.QualifiedName{Namespace: "EventName", Name: "EventName"}
var eventNameAsRecordDefinition = schema.NewRecordDefinition(eventNameQualifiedName, make([]avro.QualifiedName, 0, 0), make([]*avro.Field, 0), "", map[string]interface{}{})
var modelDefinitionQualifiedName = schema.QualifiedName{Namespace: "ModelDefinition", Name: "ModelDefinition"}
var modelAsEnumDefinition = schema.NewEnumDefinition(modelDefinitionQualifiedName, make([]avro.QualifiedName, 0, 0), []string{"", ""}, "", "defaultValue", map[string]interface{}{})

var shouldImportAvroTypeGenTests = []struct {
	testName    string
	namespace   *parser.Namespace
	definitions []schema.QualifiedName
	expect      bool
}{
	{
		testName: "true-definition-only-present-in-namespace-and-is-record-type",
		namespace: &parser.Namespace{
			Definitions: map[schema.QualifiedName]schema.Definition{
				eventNameQualifiedName: eventNameAsRecordDefinition,
			},
			Roots:       nil,
			ShortUnions: false,
		},
		definitions: []schema.QualifiedName{eventNameQualifiedName},
		expect:      true,
	},
	{
		testName: "true-definition-present-in-namespace-and-is-record-type",
		namespace: &parser.Namespace{
			Definitions: map[schema.QualifiedName]schema.Definition{
				eventNameQualifiedName:       eventNameAsRecordDefinition,
				modelDefinitionQualifiedName: modelAsEnumDefinition,
			},
			Roots:       nil,
			ShortUnions: false,
		},
		definitions: []schema.QualifiedName{eventNameQualifiedName},
		expect:      true,
	},
	{
		testName: "false-definition-present-in-namespace-and-not-record-type",
		namespace: &parser.Namespace{
			Definitions: map[schema.QualifiedName]schema.Definition{
				eventNameQualifiedName:       eventNameAsRecordDefinition,
				modelDefinitionQualifiedName: modelAsEnumDefinition,
			},
			Roots:       nil,
			ShortUnions: false,
		},
		definitions: []schema.QualifiedName{modelDefinitionQualifiedName},
		expect:      false,
	},
	{
		testName: "false-definition-not-present-in-namespace-and-not-record-type",
		namespace: &parser.Namespace{
			Definitions: map[schema.QualifiedName]schema.Definition{},
			Roots:       nil,
			ShortUnions: false,
		},
		definitions: []schema.QualifiedName{modelDefinitionQualifiedName},
		expect:      false,
	},
}

func TestShouldImportAvroTypeGen(t *testing.T) {
	c := qt.New(t)

	for _, test := range shouldImportAvroTypeGenTests {
		c.Run(test.testName, func(c *qt.C) {
			value := shouldImportAvroTypeGen(test.namespace, test.definitions)
			c.Assert(value, qt.Equals, test.expect)
		})
	}
}
