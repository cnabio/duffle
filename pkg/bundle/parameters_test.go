package bundle

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCanReadParameterNames(t *testing.T) {
	json := `{
		"parameters": {
			"foo": { },
			"bar": { }
		}
	}`
	definitions, err := Parse(json)
	if err != nil {
		t.Fatal(err)
	}
	if len(definitions.Parameters) != 2 {
		t.Fatalf("Expected 2 parameter definitons, got %d", len(definitions.Parameters))
	}
	if _, ok := definitions.Parameters["foo"]; !ok {
		t.Errorf("Expected an entry with name 'foo' but didn't get one")
	}
	if _, ok := definitions.Parameters["bar"]; !ok {
		t.Errorf("Expected an entry with name 'bar' but didn't get one")
	}
}

func TestCanReadParameterDefinition(t *testing.T) {
	dataType := "int"
	defaultValue := "some default"
	allowedValues0 := "foo"
	allowedValues1 := "bar"
	minValue := 100
	maxValue := 200
	minLength := 300
	maxLength := 400
	description := "some description"
	json := fmt.Sprintf(`{
		"parameters": {
			"test": {
				"type": "%s",
				"defaultValue": "%s",
				"allowedValues": [ "%s", "%s" ],
				"minValue": %d,
				"maxValue": %d,
				"minLength": %d,
				"maxLength": %d,
				"metadata": {
					"description": "%s"
				}
			}
		}
	}`,
		dataType, defaultValue, allowedValues0, allowedValues1,
		minValue, maxValue, minLength, maxLength, description)

	definitions, err := Parse(json)
	if err != nil {
		t.Fatal(err)
	}

	p := definitions.Parameters["test"]
	if p.DataType != dataType {
		t.Errorf("Expected data type '%s' but got '%s'", dataType, p.DataType)
	}
	if p.DefaultValue != defaultValue {
		t.Errorf("Expected default value '%s' but got '%s'", defaultValue, p.DefaultValue)
	}
	if len(p.AllowedValues) != 2 {
		t.Errorf("Expected 2 allowed values but got %d", len(p.AllowedValues))
	}
	if p.AllowedValues[0] != allowedValues0 {
		t.Errorf("Expected allowed value '%s' but got '%s'", allowedValues0, p.AllowedValues[0])
	}
	if p.AllowedValues[1] != allowedValues1 {
		t.Errorf("Expected allowed value '%s' but got '%s'", allowedValues1, p.AllowedValues[1])
	}
	if *p.MinValue != minValue {
		t.Errorf("Expected min value '%d' but got '%d'", minValue, p.MinValue)
	}
	if *p.MinLength != minLength {
		t.Errorf("Expected min length '%d' but got '%d'", minLength, p.MinLength)
	}
	if *p.MaxValue != maxValue {
		t.Errorf("Expected max value '%d' but got '%d'", maxValue, p.MaxValue)
	}
	if *p.MaxLength != maxLength {
		t.Errorf("Expected max length '%d' but got '%d'", maxLength, p.MaxLength)
	}
	if p.Metadata.Description != description {
		t.Errorf("Expected description '%s' but got '%s'", description, p.Metadata.Description)
	}
}

func valueTestJSON(jsonRepresentation string) string {
	return fmt.Sprintf(`{
		"parameters": {
			"test": {
				"defaultValue": %s,
				"allowedValues": [ %s ]
			}
		}
	}`, jsonRepresentation, jsonRepresentation)
}

func expectString(desc string, expected string, o interface{}, t *testing.T) {
	s, ok := o.(string)
	if !ok {
		t.Errorf("Expected string %s but didn't get one", desc)
	}
	if s != expected {
		t.Errorf("Expected %s '%s' but got '%s'", desc, expected, s)
	}
}

func expectInt(desc string, expected int, o interface{}, t *testing.T) {
	i, ok := o.(float64)
	if !ok {
		t.Errorf("Expected int %s but didn't get one", desc)
	}
	if int(i) != expected {
		t.Errorf("Expected %s '%d' but got '%d'", desc, expected, int(i))
	}
}

func expectBool(desc string, expected bool, o interface{}, t *testing.T) {
	b, ok := o.(bool)
	if !ok {
		t.Errorf("Expected bool %s but didn't get one", desc)
	}
	if b != expected {
		t.Errorf("Expected %s '%t' but got '%t'", desc, expected, b)
	}
}

func TestCanReadValues(t *testing.T) {
	strValue := "\"some string\""
	intValue := "123"
	boolValue := "true"

	strDef, err := Parse(valueTestJSON(strValue))
	if err != nil {
		t.Fatal(err)
	}
	expectString("default value", "some string", strDef.Parameters["test"].DefaultValue, t)
	expectString("allowed value", "some string", strDef.Parameters["test"].AllowedValues[0], t)

	intDef, err := Parse(valueTestJSON(intValue))
	if err != nil {
		t.Fatal(err)
	}
	expectInt("default value", 123, intDef.Parameters["test"].DefaultValue, t)
	expectInt("allowed value", 123, intDef.Parameters["test"].AllowedValues[0], t)

	boolDef, err := Parse(valueTestJSON(boolValue))
	if err != nil {
		t.Fatal(err)
	}
	expectBool("default value", true, boolDef.Parameters["test"].DefaultValue, t)
	expectBool("allowed value", true, boolDef.Parameters["test"].AllowedValues[0], t)
}

func TestValidateStringParameterValue_AnyAllowed(t *testing.T) {
	pd := ParameterDefinition{
		DataType: "string",
	}

	err := pd.ValidateParameterValue("foo")
	if err != nil {
		t.Errorf("Expected valid but got error %s", err)
	}

	err = pd.ValidateParameterValue(17)
	if err == nil {
		t.Errorf("Expected invalid type but got no error")
	}
}

func TestValidateStringParameterValue_AllowedOnly(t *testing.T) {
	pd := ParameterDefinition{
		DataType:      "string",
		AllowedValues: []interface{}{"foo", "bar"},
	}

	err := pd.ValidateParameterValue("foo")
	if err != nil {
		t.Errorf("Expected valid but got error %s", err)
	}

	err = pd.ValidateParameterValue("quux")
	if err == nil {
		t.Errorf("Expected disallowed value but got no error")
	}
}

func TestValidateStringParameterValue_MinLength(t *testing.T) {
	pd := ParameterDefinition{
		DataType:  "string",
		MinLength: intPtr(5),
	}

	err := pd.ValidateParameterValue("foobar")
	if err != nil {
		t.Errorf("Expected valid but got error %s", err)
	}

	err = pd.ValidateParameterValue("foo")
	if err == nil {
		t.Errorf("Expected too-short value but got no error")
	}
}

func TestValidateStringParameterValue_MaxLength(t *testing.T) {
	pd := ParameterDefinition{
		DataType:  "string",
		MaxLength: intPtr(5),
	}

	err := pd.ValidateParameterValue("foo")
	if err != nil {
		t.Errorf("Expected valid but got error %s", err)
	}

	err = pd.ValidateParameterValue("foobar")
	if err == nil {
		t.Errorf("Expected too-long value but got no error")
	}
}

func TestValidateIntParameterValue_AnyAllowed(t *testing.T) {
	pd := ParameterDefinition{
		DataType: "int",
	}

	err := pd.ValidateParameterValue(17)
	if err != nil {
		t.Errorf("Expected valid but got error %s", err)
	}

	err = pd.ValidateParameterValue(float64(17))
	if err != nil {
		t.Errorf("Expected valid but got error %s", err)
	}

	err = pd.ValidateParameterValue(17.5)
	if err == nil {
		t.Errorf("Expected not an integer but got no error")
	}

	err = pd.ValidateParameterValue("17")
	if err == nil {
		t.Errorf("Expected invalid type but got no error")
	}
}

func TestValidateIntParameterValue_AllowedOnly(t *testing.T) {
	pd := ParameterDefinition{
		DataType:      "int",
		AllowedValues: []interface{}{17, 23},
	}

	err := pd.ValidateParameterValue(17)
	if err != nil {
		t.Errorf("Expected valid but got error %s", err)
	}

	err = pd.ValidateParameterValue(58)
	if err == nil {
		t.Errorf("Expected disallowed value but got no error")
	}
}

func TestValidateIntParameterValue_MinValue(t *testing.T) {
	pd := ParameterDefinition{
		DataType: "int",
		MinValue: intPtr(5),
	}

	err := pd.ValidateParameterValue(17)
	if err != nil {
		t.Errorf("Expected valid but got error %s", err)
	}

	err = pd.ValidateParameterValue(3)
	if err == nil {
		t.Errorf("Expected too-small value but got no error")
	}
}

func TestValidateIntParameterValue_MaxValue(t *testing.T) {
	pd := ParameterDefinition{
		DataType: "int",
		MaxValue: intPtr(5),
	}

	err := pd.ValidateParameterValue(3)
	if err != nil {
		t.Errorf("Expected valid but got error %s", err)
	}

	err = pd.ValidateParameterValue(17)
	if err == nil {
		t.Errorf("Expected too-large value but got no error")
	}
}

func TestValidateBoolParameterValue(t *testing.T) {
	pd := ParameterDefinition{
		DataType: "bool",
	}

	err := pd.ValidateParameterValue(true)
	if err != nil {
		t.Errorf("Expected valid but got error %s", err)
	}

	err = pd.ValidateParameterValue(17)
	if err == nil {
		t.Errorf("Expected invalid type but got no error")
	}

	err = pd.ValidateParameterValue("17")
	if err == nil {
		t.Errorf("Expected invalid type but got no error")
	}
}

func TestConvertValue(t *testing.T) {
	pd := ParameterDefinition{
		DataType: "bool",
	}
	is := assert.New(t)

	out, _ := pd.ConvertValue("true")
	is.True(out.(bool))
	out, _ = pd.ConvertValue("false")
	is.False(out.(bool))
	out, _ = pd.ConvertValue("barbeque")
	is.False(out.(bool))

	pd.DataType = "string"
	out, err := pd.ConvertValue("hello")
	is.NoError(err)
	is.Equal("hello", out.(string))

	pd.DataType = "int"
	out, err = pd.ConvertValue("123")
	is.NoError(err)
	is.Equal(123, out.(int))

	_, err = pd.ConvertValue("onetwothree")
	is.Error(err)
}

func intPtr(i int) *int {
	return &i
}
