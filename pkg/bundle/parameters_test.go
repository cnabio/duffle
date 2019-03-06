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
	definitions, err := Unmarshal([]byte(json))
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
	is := assert.New(t)
	dataType := "int"
	defaultValue := "some default"
	allowedValues0 := "foo"
	allowedValues1 := "bar"
	minValue := 100
	maxValue := 200
	exMin := 101
	exMax := 199
	minLength := 300
	maxLength := 400
	description := "some description"
	action0 := "action0"
	action1 := "action1"

	json := fmt.Sprintf(`{
		"parameters": {
			"test": {
				"type": "%s",
				"defaultValue": "%s",
				"enum": [ "%s", "%s" ],
				"minimum": %d,
				"maximum": %d,
				"exclusiveMinimum": %d,
				"exclusiveMaximum": %d,
				"minLength": %d,
				"maxLength": %d,
				"metadata": {
					"description": "%s"
				},
				"apply-to": [ "%s", "%s" ]
			}
		}
	}`,
		dataType, defaultValue, allowedValues0, allowedValues1,
		minValue, maxValue, exMin, exMax, minLength, maxLength, description, action0, action1)

	definitions, err := Unmarshal([]byte(json))
	if err != nil {
		t.Fatal(err)
	}

	p := definitions.Parameters["test"]
	is.Equal(dataType, p.DataType)
	is.Equal(defaultValue, p.DefaultValue)
	is.Len(p.Enum, 2)
	is.Equal(allowedValues0, p.Enum[0])
	is.Equal(allowedValues1, p.Enum[1])
	is.Equal(minValue, *p.Minimum)
	is.Equal(maxValue, *p.Maximum)
	is.Equal(exMin, *p.ExclusiveMinimum)
	is.Equal(exMax, *p.ExclusiveMaximum)
	is.Equal(minLength, *p.MinLength)
	is.Equal(maxLength, *p.MaxLength)
	is.Equal(description, p.Metadata.Description)
	is.Equal(p.ApplyTo[0], action0)
	is.Equal(p.ApplyTo[1], action1)

}

func valueTestJSON(jsonRepresentation string) []byte {
	return []byte(fmt.Sprintf(`{
		"parameters": {
			"test": {
				"defaultValue": %s,
				"Enum": [ %s ]
			}
		}
	}`, jsonRepresentation, jsonRepresentation))
}

func TestCanReadValues(t *testing.T) {
	is := assert.New(t)
	strValue := "\"some string\""
	intValue := "123"
	boolValue := "true"

	strDef, err := Unmarshal(valueTestJSON(strValue))
	if err != nil {
		t.Fatal(err)
	}
	is.Equal("some string", strDef.Parameters["test"].DefaultValue)
	is.Equal("some string", strDef.Parameters["test"].Enum[0])

	intDef, err := Unmarshal(valueTestJSON(intValue))
	if err != nil {
		t.Fatal(err)
	}
	// Numerics are parsed as float64s
	is.Equal(123.0, intDef.Parameters["test"].DefaultValue)
	is.Equal(123.0, intDef.Parameters["test"].Enum[0])

	boolDef, err := Unmarshal(valueTestJSON(boolValue))
	if err != nil {
		t.Fatal(err)
	}
	is.True(boolDef.Parameters["test"].DefaultValue.(bool))
	is.True(boolDef.Parameters["test"].Enum[0].(bool))
}

func TestValidateStringParameterValue_AnyAllowed(t *testing.T) {
	is := assert.New(t)
	pd := ParameterDefinition{
		DataType: "string",
	}

	is.NoError(pd.ValidateParameterValue("foo"))
	is.Error(pd.ValidateParameterValue(17))
}

func TestValidateStringParameterValue_AllowedOnly(t *testing.T) {
	is := assert.New(t)
	pd := ParameterDefinition{
		DataType: "string",
		Enum:     []interface{}{"foo", "bar"},
	}

	is.NoError(pd.ValidateParameterValue("foo"))
	is.Error(pd.ValidateParameterValue("quux"))
}

func TestValidateStringParameterValue_MinLength(t *testing.T) {
	is := assert.New(t)
	pd := ParameterDefinition{
		DataType:  "string",
		MinLength: intPtr(5),
	}
	is.NoError(pd.ValidateParameterValue("foobar"))
	is.Error(pd.ValidateParameterValue("foo"))
}

func TestValidateStringParameterValue_MaxLength(t *testing.T) {
	is := assert.New(t)
	pd := ParameterDefinition{
		DataType:  "string",
		MaxLength: intPtr(5),
	}
	is.NoError(pd.ValidateParameterValue("foo"))
	is.Error(pd.ValidateParameterValue("foobar"))
}

func TestValidateStringParameterValue_Pattern(t *testing.T) {
	is := assert.New(t)
	pd := ParameterDefinition{
		DataType: "string",
		Pattern:  "^he[l]+o$",
	}
	is.NoError(pd.ValidateParameterValue("hello"))
	is.NoError(pd.ValidateParameterValue("helllllllo"))
	is.Error(pd.ValidateParameterValue("yolo hello"))

	// Bad pattern
	pd.Pattern = "hell[o"
	is.Error(pd.ValidateParameterValue("hello"))
}

func TestValidateIntParameterValue_AnyAllowed(t *testing.T) {
	is := assert.New(t)
	pd := ParameterDefinition{
		DataType: "int",
	}

	is.NoError(pd.ValidateParameterValue(17))
	is.NoError(pd.ValidateParameterValue(float64(17)))
	is.Error(pd.ValidateParameterValue(17.5))
	is.Error(pd.ValidateParameterValue("17"))
}

func TestValidateIntParameterValue_AllowedOnly(t *testing.T) {
	is := assert.New(t)
	pd := ParameterDefinition{
		DataType: "int",
		Enum:     []interface{}{17, 23},
	}

	is.NoError(pd.ValidateParameterValue(17))
	is.Error(pd.ValidateParameterValue(58))
}

func TestValidateIntParameterValue_Minimum(t *testing.T) {
	is := assert.New(t)
	pd := ParameterDefinition{
		DataType: "int",
		Minimum:  intPtr(5),
	}
	is.NoError(pd.ValidateParameterValue(17))
	is.NoError(pd.ValidateParameterValue(5))
	is.Error(pd.ValidateParameterValue(3))
	is.Equal("value is lower than 5", pd.ValidateParameterValue(3).Error())
}
func TestValidateIntParameterValue_ExclusiveMinimum(t *testing.T) {
	is := assert.New(t)
	pd := ParameterDefinition{
		DataType:         "int",
		ExclusiveMinimum: intPtr(5),
	}
	is.NoError(pd.ValidateParameterValue(17))
	is.Error(pd.ValidateParameterValue(3))
	is.Equal("value is less than or equal to 5", pd.ValidateParameterValue(5).Error())
}

func TestValidateIntParameterValue_Maximum(t *testing.T) {
	is := assert.New(t)
	pd := ParameterDefinition{
		DataType: "int",
		Maximum:  intPtr(5),
	}

	is.NoError(pd.ValidateParameterValue(3))
	is.NoError(pd.ValidateParameterValue(5))
	is.Error(pd.ValidateParameterValue(17))
	is.Equal("value is higher than 5", pd.ValidateParameterValue(6).Error())
}
func TestValidateIntParameterValue_ExclusiveMaximum(t *testing.T) {
	is := assert.New(t)
	pd := ParameterDefinition{
		DataType:         "int",
		ExclusiveMaximum: intPtr(5),
	}

	is.NoError(pd.ValidateParameterValue(3))
	is.Error(pd.ValidateParameterValue(17))
	is.Error(pd.ValidateParameterValue(5))
	is.Equal("value is higher than or equal to 5", pd.ValidateParameterValue(6).Error())
}

func TestValidateBoolParameterValue(t *testing.T) {
	is := assert.New(t)
	pd := ParameterDefinition{
		DataType: "bool",
	}

	is.NoError(pd.ValidateParameterValue(true))
	is.NoError(pd.ValidateParameterValue(false))
	is.Error(pd.ValidateParameterValue(17))
	is.Error(pd.ValidateParameterValue("true"))
}

func TestValidateIntParameterValue_MinimumMaximum(t *testing.T) {
	is := assert.New(t)
	pd := ParameterDefinition{
		DataType: "int",
		Minimum:  intPtr(1),
		Maximum:  intPtr(5),
	}

	is.NoError(pd.ValidateParameterValue(1))
	is.NoError(pd.ValidateParameterValue(3))
	is.NoError(pd.ValidateParameterValue(5))
	is.Error(pd.ValidateParameterValue(0))
	is.Error(pd.ValidateParameterValue(6))
}

func TestValidateIntParameterValue_ExclusiveMinimumMaximum(t *testing.T) {
	is := assert.New(t)
	pd := ParameterDefinition{
		DataType:         "int",
		ExclusiveMinimum: intPtr(1),
		ExclusiveMaximum: intPtr(5),
	}

	is.Error(pd.ValidateParameterValue(1))
	is.NoError(pd.ValidateParameterValue(3))
	is.Error(pd.ValidateParameterValue(5))
	is.Error(pd.ValidateParameterValue(0))
	is.Error(pd.ValidateParameterValue(6))
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

	pd.DataType = "chimpanzee"
	_, err = pd.ConvertValue("onetwothree")
	is.Error(err)
}

func TestCoerceValue(t *testing.T) {
	is := assert.New(t)
	pd := ParameterDefinition{
		DataType: "int",
	}
	is.Equal(5, pd.CoerceValue(5))
	is.Equal(5, pd.CoerceValue(5.0))
	// So this is definitely what the code is designed to do. But this
	// is not what the docs say it does.
	is.Equal(5.1, pd.CoerceValue(5.1))
}

func intPtr(i int) *int {
	return &i
}
