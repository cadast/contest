package main

import (
	"contract-testing/src/serialization/openapi"
	"fmt"
	"github.com/logrusorgru/aurora/v3"
	"log"
)

func CheckSchema(schema openapi.Schema, object interface{}, canonicalName string) bool {
	valid := false
	detectedType := "unknown"

	switch obj := object.(type) {
	case bool:
		detectedType = "bool"
		valid = schema.Type == openapi.SchemaTypeBoolean
	case int64:
		detectedType = "int64"
		valid = schema.Type == openapi.SchemaTypeInteger || schema.Type == openapi.SchemaTypeNumber
	case float32:
	case float64:
		detectedType = "float64"
		valid = schema.Type == openapi.SchemaTypeNumber
	case string:
		detectedType = "string"
		valid = schema.Type == openapi.SchemaTypeString
	case []interface{}:
		detectedType = "[]interface{}"
		valid = schema.Type == openapi.SchemaTypeArray
		if valid {
			for i, val := range obj {
				check := CheckSchema(*schema.Items, val, fmt.Sprintf("%s[%d]", canonicalName, i))
				valid = check && valid
			}
		}
	case map[string]interface{}:
		detectedType = "map[string]interface{}"
		valid = schema.Type == openapi.SchemaTypeObject

		for name, property := range schema.Properties {
			property.Title = name
			if val, ok := obj[name]; ok {
				check := CheckSchema(property, val, canonicalName+"."+property.Title)
				valid = check && valid
			} else {
				valid = false
				log.Println(property, "not found")
			}
		}
	case nil:
		detectedType = "nil"
		valid = schema.Nullable
	}
	if !valid {
		log.Println(canonicalName, aurora.Faint(detectedType), PassFail(valid))
	}
	_ = detectedType
	return valid
}
