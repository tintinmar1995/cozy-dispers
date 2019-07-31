package aggregations

import (
	"errors"
)

// Sum takes in input the data
// a map speciafying some parameters :
// - keys : Value on which compute the sum. The specified keys should be one of the keys from Data.
// If key isn't specified, sum will be computed on every attribute of data
// - weight : specify a key to compute a weighted sum
func Sum(data []map[string]interface{}, args map[string]interface{}) (map[string]interface{}, error) {

	var keys []string

	if args["keys"] != nil {
		// We retrieve keys from args
		switch args["keys"].(interface{}).(type) {
		case []string:
			keys = args["keys"].([]string)
		case []interface{}:
			keys = make([]string, len(args["keys"].([]interface{})))
			for index, key := range args["keys"].([]interface{}) {
				keys[index] = key.(string)
			}
		default:
			return nil, errors.New("Cannot convert args[\"keys\"]")
		}

	} else {
		// We collect the keys for one row, every variable will be summed
		keys = make([]string, 0, len(data[0]))
		for k := range data[0] {
			keys = append(keys, k)
		}
	}

	sums := make(map[string]interface{})
	isWeightedSum := (args["weight"] != nil)
	keyWeight := ""
	if isWeightedSum {
		keyWeight = args["weight"].(string)
	}
	var value float64
	var weight float64

	for _, key := range keys {
		// Initiate sum
		sumKey := 0.0
		// We iterate over the data and sum the values if possible
		for _, row := range data {

			switch row[key].(type) {
			case string:
				return nil, errors.New("Unable to operate on strings")
			case int:
				value = float64(row[key].(int))
			case float64:
				value = row[key].(float64)
			}

			if isWeightedSum {
				switch row[keyWeight].(type) {
				case string:
					return nil, errors.New("Unable to operate on strings")
				case int:
					weight = float64(row[keyWeight].(int))
				case float64:
					weight = row[keyWeight].(float64)
				}
			}

			if isWeightedSum {
				sumKey = sumKey + value/weight
			} else {
				sumKey = sumKey + value
			}
		}
		// Add sumKey to the returned structure
		sums[key] = sumKey
	}

	sums["length"] = len(data)

	return sums, nil
}
