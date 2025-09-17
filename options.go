package odoorpc

import (
	"reflect"
	"unicode"
)

type Options struct {
	Limit   int
	Fields  []string
	Order   string
	Context map[string]any
}

func toSnakeCase(in string) (out string) {
	var result []rune
	for i, r := range in {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '_', r+32)
		} else {
			result = append(result, unicode.ToLower(r))
		}
	}
	return string(result)
}

func (opt Options) Kwargs() map[string]any {
	kwargs := make(map[string]any)
	v := reflect.ValueOf(opt)
	t := reflect.TypeOf(opt)
	for i := 0; i < t.NumField(); i++ {
		fieldValue := v.Field(i).Interface()
		fieldName := toSnakeCase(t.Field(i).Name)
		switch fieldValue := fieldValue.(type) {
		case int:
			if fieldValue != 0 {
				kwargs[fieldName] = fieldValue
			}
		case []string:
			if len(fieldValue) > 0 {
				kwargs[fieldName] = fieldValue
			}
		case string:
			if fieldValue != "" {
				kwargs[fieldName] = fieldValue
			}
		case map[string]any:
			if len(fieldValue) > 0 {
				kwargs[fieldName] = fieldValue
			}
		}
	}
	return kwargs
}
