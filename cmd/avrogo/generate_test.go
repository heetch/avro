package main

import (
	"testing"

	"github.com/actgardner/gogen-avro/v10/parser"
	"github.com/actgardner/gogen-avro/v10/schema"

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

func TestGoName(t *testing.T) {
	var testcases = []struct {
		testName         string
		goInitialisms    bool
		extraInitialisms string
		avroName         string
		goName           string
	}{
		{
			testName:      "default naming",
			goInitialisms: false,
			avroName:      "user.first_name",
			goName:        "First_name",
		},
		{
			testName:      "Go initialisms",
			goInitialisms: true,
			avroName:      "user.first_name",
			goName:        "FirstName",
		},
		{
			testName:      "default naming with ID",
			goInitialisms: false,
			avroName:      "user.user_id",
			goName:        "User_id",
		},
		{
			testName:      "Go initialisms with ID",
			goInitialisms: true,
			avroName:      "user.user_id",
			goName:        "UserID",
		},
		{
			testName:      "Go initialisms without extra initialisms",
			goInitialisms: true,
			avroName:      "power.power_mw",
			goName:        "PowerMw",
		},
		{
			testName:         "Go initialisms with extra initialisms",
			goInitialisms:    true,
			extraInitialisms: "KW,MW,GW",
			avroName:         "power.power_mw",
			goName:           "PowerMW",
		},
	}

	c := qt.New(t)

	for _, test := range testcases {
		c.Run(test.testName, func(c *qt.C) {
			goInitalismFlag = &test.goInitialisms
			extraInitialismsFlag = &test.extraInitialisms
			gc := generateContext{caser: getCaser()}

			gotName, err := gc.goName(test.avroName)
			c.Assert(err, qt.IsNil)
			c.Assert(gotName, qt.Equals, test.goName)
		})
	}
}
