package vault2cfg

import (
	"fmt"
	"reflect"
)

const defaultTagName = "vault"

type settings struct {
	tagName string
}

// Bind vault data values to configuration struct
func Bind(cfg interface{}, data map[string]interface{}, opts ...Option) error {
	settings := new(settings)
	for _, opt := range opts {
		opt(settings)
	}
	if settings.tagName == "" {
		settings.tagName = defaultTagName
	}
	return bind(cfg, data, settings)
}

func bind(cfg interface{}, data map[string]interface{}, settings *settings) error {
	var (
		rt = reflect.TypeOf(cfg)
		rv = reflect.ValueOf(cfg)
	)

	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("cfg must be a pointer to a struct")
	}

	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
		rv = rv.Elem()
	}

	for i := 0; i < rt.NumField(); i++ {
		var (
			field = rt.Field(i)
			value = rv.Field(i)
		)

		// skip unexported
		if len(field.PkgPath) != 0 {
			continue
		}

		switch field.Type.Kind() {
		case reflect.Struct:
			if value.CanAddr() {
				// Process nested struct but don't return immediately to continue with other fields
				if err := bind(value.Addr().Interface(), data, settings); err != nil {
					return err
				}
			}
		case reflect.Ptr:
			if !value.IsNil() && field.Type.Elem().Kind() == reflect.Struct {
				// Process pointer to struct but don't return immediately
				if err := bind(value.Interface(), data, settings); err != nil {
					return err
				}
			}
		default:
			if secretPath := field.Tag.Get(settings.tagName); len(secretPath) > 0 {
				secretValue, exists := data[secretPath]
				if !exists {
					continue
				}
				if value.CanSet() {
					err := processValue(field, value, secretValue)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func processValue(f reflect.StructField, v reflect.Value, new interface{}) error {
	if new == nil {
		return nil
	}

	switch v.Kind() {
	case reflect.String:
		switch val := new.(type) {
		case string:
			v.SetString(val)
		default:
			v.SetString(fmt.Sprintf("%v", val))
		}
	default:
		return fmt.Errorf("field %s have unsupported type %s", f.Name, v.Type())
	}

	return nil
}
