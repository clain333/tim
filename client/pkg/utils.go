package pkg

import (
	"fmt"
	"reflect"
	"strconv"
)

func StrTurnUint(s string) uint64 {
	num, _ := strconv.ParseUint(s, 10, 64)
	return num
}

func UintTurnStr(s uint64) string {
	return strconv.FormatUint(s, 10)
}

func Cpnum(one, two uint64) (min, mix uint64) {
	if one < two {
		return one, two
	} else {
		return two, one
	}
}
func RemoveFields(input interface{}, excludeFields ...string) (map[string]interface{}, error) {
	v := reflect.ValueOf(input)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("input must be a struct or pointer to struct")
	}

	out := make(map[string]interface{})
	exclude := make(map[string]struct{})
	for _, f := range excludeFields {
		exclude[f] = struct{}{}
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		name := field.Tag.Get("json")
		if name == "" {
			name = field.Name
		}
		if _, ok := exclude[field.Name]; ok {
			continue
		}
		out[name] = v.Field(i).Interface()
	}
	return out, nil
}
