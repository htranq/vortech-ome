package config

import (
	"fmt"
	"reflect"
	"strings"
)

func uppercaseMapKey(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range data {
		result[strings.ToUpper(k)] = v
	}
	return result
}

func getMapField(data map[string]interface{}, field string) (map[string]interface{}, error) {
	if field == "" {
		return data, nil
	}
	fields := strings.Split(field, ".")
	var val = data
	for i, f := range fields {
		if i == len(fields)-1 {
			result, ok := val[f]
			if !ok {
				break
			}
			return result.(map[string]interface{}), nil
		}
		tmp := val[f]
		if reflect.TypeOf(tmp).Kind() == reflect.Map {
			val = tmp.(map[string]interface{})
		} else {
			return nil, fmt.Errorf("%s is not a map", field)
		}
	}
	return nil, fmt.Errorf("%s is not found", field)
}
