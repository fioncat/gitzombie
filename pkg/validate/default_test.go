package validate

import (
	"os"
	"reflect"
	"testing"
)

func TestDefault(t *testing.T) {
	type testStruct struct {
		Name string `default:"john"`
		Desc string `default:"hello"`

		Limit int `default:"32"`

		Path string `default:"$HOME/dev" env:"true"`

		NoneDefault    string `default:"none-default"`
		NoneDefaultNum int    `default:"404"`
	}

	s := &testStruct{
		NoneDefault:    "user-specifed",
		NoneDefaultNum: 200,
	}
	ExpandDefault(s)

	expected := &testStruct{
		Name:  "john",
		Desc:  "hello",
		Limit: 32,

		Path: os.ExpandEnv("$HOME/dev"),

		NoneDefault:    "user-specifed",
		NoneDefaultNum: 200,
	}
	if !reflect.DeepEqual(s, expected) {
		t.Fatalf("unexpected value: %+v", s)
	}
}
