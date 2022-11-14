package validate

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/fioncat/gitzombie/pkg/term"
	"github.com/go-playground/validator/v10"
	"github.com/iancoleman/strcase"
)

// Use for generating enum validators.
var enumMap = map[string][]string{
	"protocol": {"https", "ssh"},
	"provider": {"github", "gitlab"},
}

var (
	// Global validator
	instance     *validator.Validate
	instanceOnce sync.Once
)

func initInstance() {
	instance = validator.New()
	for enumKey, enumValues := range enumMap {
		enumValues := enumValues
		tag := fmt.Sprintf("enum_%s", enumKey)
		instance.RegisterValidation(tag, func(fl validator.FieldLevel) bool {
			str := fl.Field().String()
			for _, val := range enumValues {
				if val == str {
					return true
				}
			}
			return false
		})
	}
}

func Do(v any) error {
	instanceOnce.Do(initInstance)
	err := instance.Struct(v)
	if err == nil {
		return nil
	}
	if _, ok := err.(*validator.InvalidValidationError); ok {
		return err
	}
	var errs Error
	for _, err := range err.(validator.ValidationErrors) {
		name := convertFieldName(err.StructNamespace())
		value := fmt.Sprint(err.Value())
		tag := err.Tag()
		errs.Fields = append(errs.Fields, &ErrorField{
			Name:  name,
			Tag:   tag,
			Value: value,

			val: err.Value(),
		})
	}
	return &errs
}

func convertFieldName(ns string) string {
	parts := strings.Split(ns, ".")
	if len(parts) <= 1 {
		return ns
	}

	// Trim the root struct.
	// For example, "Config.Remote.Name" should be converted to
	// "remote.name", the root "Config" is trimed.
	parts = parts[1:]

	for i, part := range parts {
		parts[i] = strcase.ToSnake(part)
	}

	return strings.Join(parts, ".")
}

type Error struct {
	Fields []*ErrorField
}

func (err *Error) Error() string {
	return "failed to validate configuration:"
}

func (err *Error) Extra() {
	for _, f := range err.Fields {
		term.Print(f.String())
	}
}

type ErrorField struct {
	Name  string
	Tag   string
	Value string

	val any
}

type tagPrinter struct {
	prefix string
	print  func(f *ErrorField) string
}

// Error message printers for different type of tags.
// We use prefix to match tags.
var tagPrinters = []*tagPrinter{
	{
		prefix: "required",
		print: func(f *ErrorField) string {
			return "cannot be empty"
		},
	},
	{
		prefix: "enum_",
		print: func(f *ErrorField) string {
			enumName := strings.TrimPrefix(f.Tag, "enum_")
			enumVals := enumMap[enumName]
			return fmt.Sprintf("invalid %s %q, expect one of %v", enumName, f.Value, enumVals)
		},
	},
	{
		prefix: "unique",
		print: func(f *ErrorField) string {
			val := reflect.ValueOf(f.val)
			sliceLen := val.Len()
			var duplicateName string
			set := make(map[string]int, sliceLen)
			var ia, ib int
			for i := 0; i < sliceLen; i++ {
				nameField := val.Index(i).Elem().FieldByName("Name")
				name := nameField.String()
				if j, ok := set[name]; ok {
					ia = j
					ib = i
					duplicateName = name
					break
				}
				set[name] = i
			}
			return fmt.Sprintf("found duplicate name %q, index: %d and %d", duplicateName, ia, ib)
		},
	},
}

func (f *ErrorField) String() string {
	var msg string
	for _, printer := range tagPrinters {
		if strings.HasPrefix(f.Tag, printer.prefix) {
			msg = printer.print(f)
			break
		}
	}
	if msg == "" {
		msg = fmt.Sprintf("invalid %s %q", f.Tag, f.Value)
	}
	return fmt.Sprintf("%s: %s", f.Name, msg)
}
