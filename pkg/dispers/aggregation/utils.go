package aggregations

import "github.com/cozy/cozy-stack/pkg/dispers/errors"

func AsFloat64(in interface{}) (float64, error) {
	switch in.(type) {
	case string:
		return 0.0, errors.ErrStrToFloat
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

func NeedArgs(args map[string]interface{}, keys ...string) error {
	for _, key := range keys {
		if _, ok := args[key]; !ok {
			return errors.ErrKeyNotFound
		}
	}
	return nil
}
