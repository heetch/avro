package main

import (
	"bytes"
	"github.com/actgardner/gogen-avro/v10/parser"
	"github.com/actgardner/gogen-avro/v10/schema"
	"github.com/sebdah/goldie/v2"
	"github.com/stretchr/testify/assert"
	"testing"

	avro "github.com/actgardner/gogen-avro/v10/schema"
	qt "github.com/frankban/quicktest"
)

func TestShouldImportAvroTypeGen(t *testing.T) {
	var eventNameQualifiedName = schema.QualifiedName{Namespace: "EventName", Name: "EventName"}
	var eventNameAsRecordDefinition = schema.NewRecordDefinition(eventNameQualifiedName, []avro.QualifiedName{}, []*avro.Field{}, "", map[string]interface{}{})
	var eventNameAsFixedFieldDefinition = schema.NewFixedDefinition(eventNameQualifiedName, []avro.QualifiedName{}, 142, map[string]interface{}{})
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
			testName: "true-definition-only-present-in-namespace-and-is-fixed-type",
			namespace: &parser.Namespace{
				Definitions: map[schema.QualifiedName]schema.Definition{
					eventNameQualifiedName: eventNameAsFixedFieldDefinition,
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
	}

	c := qt.New(t)

	for _, test := range shouldImportAvroTypeGenTests {
		c.Run(test.testName, func(c *qt.C) {
			value := shouldImportAvroTypeGen(test.namespace, test.definitions)
			c.Assert(value, qt.Equals, test.shouldImportAvroTypeGen)
		})
	}
}

func TestGenerate(t *testing.T) {
	fixtureDir := "testdata/schema"
	g := goldie.New(t, goldie.WithFixtureDir(fixtureDir), goldie.WithNameSuffix(".golden.go"))

	var buf bytes.Buffer
	testPackage := "dummy"
	ns, fileDefinitions, err := parseFiles([]string{"testdata/schema/object.avsc"})
	assert.NoError(t, err)

	err = generate(&buf, testPackage, ns, fileDefinitions[0])
	assert.NoError(t, err)
	g.Assert(t, "object", buf.Bytes())
}
