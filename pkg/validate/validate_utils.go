package validate

import (
	"github.com/go-playground/validator/v10"
	"reflect"
	"strings"
)

func ValidateStruct(v any) map[string]string {
	if V == nil {
		V = validator.New()
	}
	err := V.Struct(v)
	if err == nil {
		return nil
	}

	errors := make(map[string]string)

	for _, e := range err.(validator.ValidationErrors) {
		field := e.Field()
		jsonField := toJSONField(v, field)

		switch e.Tag() {
		case "required":
			errors[jsonField] = jsonField + " is required"
		case "min":
			errors[jsonField] = jsonField + " must not be empty"
		case "email":
			errors[jsonField] = jsonField + " must be a valid email"
		default:
			errors[jsonField] = jsonField + " is invalid"
		}
	}

	return errors
}

func toJSONField(v any, structField string) string {
	t := reflect.TypeOf(v)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if f, ok := t.FieldByName(structField); ok {
		tag := f.Tag.Get("json")
		if tag != "" && tag != "-" {
			return strings.Split(tag, ",")[0]
		}
	}

	return strings.ToLower(structField)
}
