package odoorpc

import "reflect"

type Options struct {
	Limit  int
	Fields []string
}

func toSnakeCase(in string) (out string) {
	var result []rune
	for i, r := range in {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '_', r+32)
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

func (opt Options) SetLimit(i int) Options {
	opt.Limit = i
	return opt
}

func (opt Options) SetFields(args ...string) Options {
	opt.Fields = args
	return opt
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
		}
	}
	return kwargs
}
