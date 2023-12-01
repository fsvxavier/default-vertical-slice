package main

import (
	"fmt"

	"github.com/invopop/jsonschema"
	"github.com/samber/lo"

	"github.com/fsvxavier/default-vertical-slice/internal/core/domains"
)

type SchemaValidationType string

const (
	INVALID_TYPE              SchemaValidationType = "invalid_type"
	ADDITIONAL_PROPERTY       SchemaValidationType = "additional_property_not_allowed"
	ATTRIBUTE_MUST_BE_PRESENT SchemaValidationType = "required"
	VALIDATION_TAG_KEYWORD    string               = "validate"

	// Use if field must be present (does not check if the value is nil or empty).
	REQUIRED_ATTRIBUTE_KEYWORD string = "is_present"

	DOCK_DECIMAL_TYPE = "Decimal"
)

func main() {
	model := domains.HealthCheck{}

	types := []SchemaValidationType{INVALID_TYPE, ATTRIBUTE_MUST_BE_PRESENT}

	reflector := new(jsonschema.Reflector)
	reflector.AllowAdditionalProperties = !lo.Contains(types, ADDITIONAL_PROPERTY)

	schema := reflector.Reflect(model)

	bs, err := schema.MarshalJSON()
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println(string(bs))
}
