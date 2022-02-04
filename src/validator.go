package main

import (
	"contract-testing/src/serialization/openapi"
	"fmt"
)

func CheckSchema(schema openapi.Schema, object interface{}, canonicalName string) bool {
	valid := false

	switch obj := object.(type) {
	case bool:
		valid = schema.Type == openapi.SchemaTypeBoolean
	case int64:
		valid = schema.Type == openapi.SchemaTypeInteger || schema.Type == openapi.SchemaTypeNumber
	case float32:
	case float64:
		valid = schema.Type == openapi.SchemaTypeNumber
	case string:
		valid = schema.Type == openapi.SchemaTypeString
	case []interface{}:
		valid = schema.Type == openapi.SchemaTypeArray

		for i, val := range obj {
			valid = CheckSchema(*schema.Items, val, fmt.Sprintf("%s[%d]", canonicalName, i)) && valid
		}
	case map[string]interface{}:
		valid = schema.Type == openapi.SchemaTypeObject

		for name, property := range schema.Properties {
			property.Title = name
			if val, ok := obj[name]; ok {
				valid = CheckSchema(property, val, canonicalName+"."+property.Title) && valid
			} else {
				valid = false
				//log.Println(property, "not found")
			}
		}
	case nil:
		valid = schema.Nullable
	}
	//log.Println(canonicalName, aurora.Faint(detectedType), PassFail(valid))
	return valid
}
