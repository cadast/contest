package main

import (
	"encoding/json"
	"errors"
	"strconv"
)

func JsonUnmarshal(data []byte) (interface{}, error) {
	// Unmarshal into an interface
	var unmarshalled interface{}
	if err := json.Unmarshal(data, &unmarshalled); err != nil {
		return nil, err
	}

	// Check interface type
	switch unmarshalled.(type) {
	case []interface{}:
		// If the value is an array, recursively unmarshal every item
		var raw []json.RawMessage
		if err := json.Unmarshal(data, &raw); err != nil {
			return nil, err
		}
		parsed := make([]interface{}, len(raw))
		for i, message := range raw {
			res, err := JsonUnmarshal(message)
			if err != nil {
				return nil, err
			}
			parsed[i] = res
		}
		return parsed, nil
	case map[string]interface{}:
		// If the value is a map, recursively unmarshal every entry of the map
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(data, &raw); err != nil {
			return nil, err
		}
		parsed := make(map[string]interface{}, len(raw))
		for key, message := range raw {
			res, err := JsonUnmarshal(message)
			if err != nil {
				return nil, err
			}
			parsed[key] = res
		}
		return parsed, nil
	case float64:
		// If the value is a float, check whether it is an int
		s := string(data)

		parsedInt, err := strconv.ParseInt(s, 10, 64)
		if err == nil {
			return parsedInt, nil
		}
	}

	// Return the default unmarshalled value, if it doesn't need to be manually unmarshalled
	return unmarshalled, nil
}

func retypeKeysToStrings(m interface{}) (interface{}, error) {
	switch m.(type) {
	case map[string]interface{}:
		retypedMap := m.(map[string]interface{})
		for k, v := range retypedMap {
			retyped, err := retypeKeysToStrings(v)
			if err != nil {
				return nil, err
			}
			retypedMap[k] = retyped
		}
		return retypedMap, nil
	case map[interface{}]interface{}:
		val, err := retypeMapToStringKeys(m.(map[interface{}]interface{}))
		if err != nil {
			return nil, err
		}
		for k, v := range val {
			retyped, err := retypeKeysToStrings(v)
			if err != nil {
				return nil, err
			}
			val[k] = retyped
		}
		return val, nil
	case []interface{}:
		arr := m.([]interface{})
		for i, v := range arr {
			retyped, err := retypeKeysToStrings(v)
			if err != nil {
				return nil, err
			}
			arr[i] = retyped
		}
		return arr, nil
	default:
		return m, nil
	}
}

func retypeMapToStringKeys(m map[interface{}]interface{}) (map[string]interface{}, error) {
	retyped := make(map[string]interface{})

	for k, v := range m {
		switch k.(type) {
		case string:
			break
		default:
			return nil, errors.New("only strings are allowed as keys in JSON objects")
		}

		retyped[k.(string)] = v
	}

	return retyped, nil
}

func JsonMarshal(data interface{}) ([]byte, error) {
	retyped, err := retypeKeysToStrings(data)
	if err != nil {
		return nil, err
	}

	return json.Marshal(retyped)
}
