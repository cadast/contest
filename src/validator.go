package main

import (
	"contract-testing/src/serialization/openapi"
	"fmt"
)

func CheckSchema(schema openapi.Schema, object interface{}, canonicalName string, messages *[]string) bool {
	typeValid := false
	childrenValid := false
	detectedType := "unknown"

	switch obj := object.(type) {
	case bool:
		detectedType = string(openapi.SchemaTypeBoolean)
		typeValid = schema.Type == openapi.SchemaTypeBoolean
	case int64:
		detectedType = string(openapi.SchemaTypeInteger)
		typeValid = schema.Type == openapi.SchemaTypeInteger || schema.Type == openapi.SchemaTypeNumber
	case float32:
	case float64:
		detectedType = string(openapi.SchemaTypeNumber)
		typeValid = schema.Type == openapi.SchemaTypeNumber
	case string:
		detectedType = string(openapi.SchemaTypeString)
		typeValid = schema.Type == openapi.SchemaTypeString
	case []interface{}:
		detectedType = string(openapi.SchemaTypeArray)
		typeValid = schema.Type == openapi.SchemaTypeArray
		if typeValid {
			for i, val := range obj {
				check := CheckSchema(*schema.Items, val, fmt.Sprintf("%s[%d]", canonicalName, i), messages)
				childrenValid = check && childrenValid
			}
		}
	case map[string]interface{}:
		detectedType = string(openapi.SchemaTypeObject)
		typeValid = schema.Type == openapi.SchemaTypeObject

		for name, property := range schema.Properties {
			property.Title = name
			if val, ok := obj[name]; ok {
				check := CheckSchema(*property, val, canonicalName+"."+property.Title, messages)
				childrenValid = check && childrenValid
			} else if schema.Requires(property.Title) {
				childrenValid = false
				*messages = append(*messages, "missing property "+canonicalName+"."+property.Title)
			}
		}
	case nil:
		detectedType = "null"
		typeValid = schema.Nullable
	}
	if !typeValid {
		*messages = append(*messages, fmt.Sprintf("%s is %s not %s", canonicalName, detectedType, schema.Type))
	}
	return typeValid && childrenValid
}
