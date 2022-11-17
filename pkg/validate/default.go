package validate

import (
	"os"
	"reflect"
	"strconv"
)

func ExpandDefault(v any) {
	ptr := reflect.ValueOf(v)
	if ptr.Kind() != reflect.Pointer {
		panic("v MUST be a pointer")
	}
	val := ptr.Elem()
	valType := val.Type()

	numField := valType.NumField()
	for i := 0; i < numField; i++ {
		fieldType := valType.Field(i)
		fieldVal := val.Field(i)
		defaultVal := fieldType.Tag.Get("default")
		if defaultVal != "" {
			switch val := fieldVal.Interface().(type) {
			case string:
				if val == "" {
					fieldVal.SetString(defaultVal)
				}

			case int:
				if val <= 0 {
					intVal, err := strconv.Atoi(defaultVal)
					if err != nil {
						panic(err)
					}
					fieldVal.SetInt(int64(intVal))
				}
			}
		}

		env := fieldType.Tag.Get("env")
		if env != "" {
			if str, ok := fieldVal.Interface().(string); ok {
				str = os.ExpandEnv(str)
				fieldVal.SetString(str)
			}
		}
	}
}
