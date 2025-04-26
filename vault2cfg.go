package vault2cfg

import (
	"reflect"
)

const tagName = "vault"

// Bind vault data values to configuration struct
// Tag defining name of variable in data map is defined by `tagName`
func Bind(cfg interface{}, data map[string]interface{}) {
	bind(cfg, data)
}

func bind(cfg interface{}, data map[string]interface{}) {
	var (
		rt = reflect.TypeOf(cfg)
		rv = reflect.ValueOf(cfg)
	)

	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
		rv = rv.Elem()
	}

	for i := 0; i < rt.NumField(); i++ {
		var (
			field = rt.Field(i)
			value = rv.FieldByName(field.Name)
		)

		// skip unexported
		if len(field.PkgPath) != 0 {
			continue
		}

		switch field.Type.Kind() {
		case reflect.Struct:
			bind(value.Addr().Interface(), data)
		default:
			if secretPath := field.Tag.Get(tagName); len(secretPath) > 0 {
				secretValue, exists := data[secretPath]
				if !exists {
					continue
				}
				processValue(value, secretValue)
			}
		}
	}
}

func processValue(v reflect.Value, new interface{}) {
	if new == nil {
		return
	}

	//nolint:gocritic
	switch v.Kind() {
	case reflect.String:
		strValue, ok := new.(string)
		if !ok {
			return
		}
		v.SetString(strValue)
	}
}
