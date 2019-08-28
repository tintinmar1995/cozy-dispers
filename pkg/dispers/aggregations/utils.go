package aggregations

import "errors"

func asFloat64(in interface{}) (float64, error) {
	switch in.(type) {
	case string:
		return 0.0, errors.New("Unable to operate on strings")
	case int:
		return float64(in.(int)), nil
	case bool:
		if in.(bool) {
			return 1.0, nil
		} else {
			return 0.0, nil
		}
	case float64:
		return in.(float64), nil
	}
	return 0.0, nil
}

func needArgs(args map[string]interface{}, keys ...string) error {
	for _, key := range keys {
		if _, ok := args[key]; !ok {
			return errors.New("Arg " + key + " not found")
		}
	}
	return nil
}

func retrieveKeys(argsKeys interface{}) ([]string, error) {
	var keys []string
	if argsKeys != nil {
		switch argsKeys.(interface{}).(type) {
		case []string:
			keys = argsKeys.([]string)
			return keys, nil
		case []interface{}:
			keys = make([]string, len(argsKeys.([]interface{})))
			for index, key := range argsKeys.([]interface{}) {
				keys[index] = key.(string)
			}
			return keys, nil
		default:
			return nil, errors.New("Cannot convert args[\"keys\"]")
		}
	}
	return nil, errors.New("Cannot find args[\"keys\"]")
}
