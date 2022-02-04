package main

import (
	"encoding/json"
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
