package aggregations

import (
	"errors"
)

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

func asFloat64(in interface{}) (float64, error) {
	switch in.(type) {
	case string:
		return 0.0, errors.New("Unable to operate on strings")
	case int:
		return float64(in.(int)), nil
	case float64:
		return in.(float64), nil
	}
	return 0.0, nil
}

// Sum takes in input the data
// a map speciafying some parameters :
// - keys : Value on which compute the sum. The specified keys should be one of the keys from Data.
// - weight : specify a key to compute a weighted sum
func Sum(result map[string]interface{}, row map[string]interface{}, args map[string]interface{}) (map[string]interface{}, error) {

	// We retrieve from args the keys on which compute sum
	keys, err := retrieveKeys(args["keys"])
	if err != nil {
		return nil, err
	}

	// We retrieve from args the key that will be used to weight the sum
	isWeightedSum := (args["weight"] != nil)
	keyWeight := ""
	if isWeightedSum {
		keyWeight = args["weight"].(string)
	}

	var value float64
	var weight float64
	for _, key := range keys {

		value, err = asFloat64(row[key])
		if err != nil {
			return nil, err
		}

		if isWeightedSum {
			weight, err = asFloat64(row[keyWeight])
			if err != nil {
				return nil, err
			}
		}

		previousRes, err := asFloat64(result["sum_"+key])
		if err != nil {
			return nil, err
		}

		if isWeightedSum {
			result["sum_"+key] = previousRes + value/weight
		} else {
			result["sum_"+key] = previousRes + value
		}
	}
	return result, nil
}

// SumSquare takes in input the data
// a map speciafying some parameters :
// - keys : Value on which compute the sum. The specified keys should be one of the keys from Data.
// - weight : specify a key to compute a weighted sum
func SumSquare(result map[string]interface{}, row map[string]interface{}, args map[string]interface{}) (map[string]interface{}, error) {

	// We retrieve from args the keys on which compute sum
	keys, err := retrieveKeys(args["keys"])
	if err != nil {
		return nil, err
	}

	// We retrieve from args the key that will be used to weight the sum
	isWeightedSum := (args["weight"] != nil)
	keyWeight := ""
	if isWeightedSum {
		keyWeight = args["weight"].(string)
	}

	var value float64
	var weight float64
	for _, key := range keys {

		value, err = asFloat64(row[key])
		if err != nil {
			return nil, err
		}

		if isWeightedSum {
			weight, err = asFloat64(row[keyWeight])
			if err != nil {
				return nil, err
			}
		}

		previousRes, err := asFloat64(result["sum_square_"+key])
		if err != nil {
			return nil, err
		}

		if isWeightedSum {
			result["sum_square_"+key] = previousRes + value*value/weight
		} else {
			result["sum_square_"+key] = previousRes + value*value
		}
	}
	return result, nil
}

// Min takes in input the data
// a map speciafying some parameters :
// - keys : Value on which compute the Min. The specified keys should be one of the keys from Data.
func Min(indexRow int, result map[string]interface{}, row map[string]interface{}, args map[string]interface{}) (map[string]interface{}, error) {

	// We retrieve from args the keys on which compute sum
	keys, err := retrieveKeys(args["keys"])
	if err != nil {
		return nil, err
	}

	var value float64
	for _, key := range keys {

		value, err = asFloat64(row[key])
		if err != nil {
			return nil, err
		}

		previousRes, err := asFloat64(result["min_"+key])
		if err != nil {
			return nil, err
		}

		if indexRow == 0 || previousRes > value {
			result["min_"+key] = value
		} else {
			result["min_"+key] = previousRes
		}

	}
	return result, nil
}

// Max takes in input the data
// a map speciafying some parameters :
// - keys : Value on which compute the Max. The specified keys should be one of the keys from Data.
func Max(indexRow int, result map[string]interface{}, row map[string]interface{}, args map[string]interface{}) (map[string]interface{}, error) {

	// We retrieve from args the keys on which compute sum
	keys, err := retrieveKeys(args["keys"])
	if err != nil {
		return nil, err
	}

	var value float64
	for _, key := range keys {

		value, err = asFloat64(row[key])
		if err != nil {
			return nil, err
		}

		previousRes, err := asFloat64(result["max_"+key])
		if err != nil {
			return nil, err
		}

		if indexRow == 0 || previousRes < value {
			result["max_"+key] = value
		} else {
			result["max_"+key] = previousRes
		}

	}
	return result, nil
}
