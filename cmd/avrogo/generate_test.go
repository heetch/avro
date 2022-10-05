package main

import (
	"github.com/actgardner/gogen-avro/v10/parser"
	"github.com/actgardner/gogen-avro/v10/schema"
	"testing"

	avro "github.com/actgardner/gogen-avro/v10/schema"
	qt "github.com/frankban/quicktest"
)

func TestShouldImportAvroTypeGen(t *testing.T) {
	var eventNameQualifiedName = schema.QualifiedName{Namespace: "EventName", Name: "EventName"}
	var eventNameAsRecordDefinition = schema.NewRecordDefinition(eventNameQualifiedName, []avro.QualifiedName{}, []*avro.Field{}, "", map[string]interface{}{})
	var modelDefinitionQualifiedName = schema.QualifiedName{Namespace: "ModelDefinition", Name: "ModelDefinition"}
	var modelAsEnumDefinition = schema.NewEnumDefinition(modelDefinitionQualifiedName, []avro.QualifiedName{}, []string{"", ""}, "", "defaultValue", map[string]interface{}{})

	var shouldImportAvroTypeGenTests = []struct {
		testName                string
		namespace               *parser.Namespace
		definitions             []schema.QualifiedName
		shouldImportAvroTypeGen bool
	}{
		{
			testName: "true-definition-only-present-in-namespace-and-is-record-type",
			namespace: &parser.Namespace{
				Definitions: map[schema.QualifiedName]schema.Definition{
					eventNameQualifiedName: eventNameAsRecordDefinition,
				},
			},
			definitions:             []schema.QualifiedName{eventNameQualifiedName},
			shouldImportAvroTypeGen: true,
		},
		{
			testName: "true-definition-present-in-namespace-and-is-record-type",
			namespace: &parser.Namespace{
				Definitions: map[schema.QualifiedName]schema.Definition{
					eventNameQualifiedName:       eventNameAsRecordDefinition,
					modelDefinitionQualifiedName: modelAsEnumDefinition,
				},
			},
			definitions:             []schema.QualifiedName{eventNameQualifiedName},
			shouldImportAvroTypeGen: true,
		},
		{
			testName: "false-definition-present-in-namespace-and-not-record-type",
			namespace: &parser.Namespace{
				Definitions: map[schema.QualifiedName]schema.Definition{
					eventNameQualifiedName:       eventNameAsRecordDefinition,
					modelDefinitionQualifiedName: modelAsEnumDefinition,
				},
			},
			definitions:             []schema.QualifiedName{modelDefinitionQualifiedName},
			shouldImportAvroTypeGen: false,
		},
		{
			testName: "false-definition-not-present-in-namespace-and-not-record-type",
			namespace: &parser.Namespace{
				Definitions: map[schema.QualifiedName]schema.Definition{},
			},
			definitions:             []schema.QualifiedName{modelDefinitionQualifiedName},
			shouldImportAvroTypeGen: false,
		},
	}

	c := qt.New(t)

	for _, test := range shouldImportAvroTypeGenTests {
		c.Run(test.testName, func(c *qt.C) {
			value := shouldImportAvroTypeGen(test.namespace, test.definitions)
			c.Assert(value, qt.Equals, test.shouldImportAvroTypeGen)
		})
	}
}
