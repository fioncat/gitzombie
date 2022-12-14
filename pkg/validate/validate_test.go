package validate

import (
	"testing"

	"github.com/fioncat/gitzombie/pkg/term"
)

func TestValidate(t *testing.T) {
	type TestField struct {
		Name  string `validate:"required"`
		User  string
		Email string `validate:"omitempty,email"`
	}

	type TestStruct struct {
		Name     string       `validate:"required"`
		Protocol string       `validate:"enum_protocol"`
		Email    string       `validate:"email"`
		Provider string       `validate:"enum_provider"`
		Fields   []*TestField `validate:"unique=Name,dive"`
	}

	testCases := []struct {
		val *TestStruct
		ok  bool
	}{
		{
			val: &TestStruct{
				Name:     "test0",
				Protocol: "https",
				Email:    "lazycat7706@gmail.com",
				Provider: "github",
			},
			ok: true,
		},
		{
			val: &TestStruct{
				Name:     "",
				Protocol: "ssh",
				Email:    "lazycat7706@gmail.com",
				Provider: "github",
			},
			ok: false,
		},
		{
			val: &TestStruct{
				Name:     "",
				Protocol: "zzz",
				Email:    "lazycat7706@gmail.com",
				Provider: "github",
			},
			ok: false,
		},
		{
			val: &TestStruct{
				Name:     "",
				Protocol: "ssh",
				Email:    "zzzz",
				Provider: "fake",
			},
			ok: false,
		},
		{
			val: &TestStruct{
				Name:     "github",
				Protocol: "https",
				Email:    "lazycat7706@gmail.com",
				Provider: "github",

				Fields: []*TestField{
					{
						Name: "name0",
					},
					{
						Name: "name0",
					},
				},
			},
			ok: false,
		},
		{
			val: &TestStruct{
				Name:     "github",
				Protocol: "https",
				Email:    "lazycat7706@gmail.com",
				Provider: "github",

				Fields: []*TestField{
					{
						Name: "",
						User: "hello",
					},
				},
			},
			ok: false,
		},
	}

	for _, testCase := range testCases {
		val := testCase.val
		ok := testCase.ok
		err := Do(val)
		if ok && err != nil {
			term.PrintError(err)
			t.Fatalf("unexpected error")
		}
		if !ok && err == nil {
			t.Fatalf("expected error, found nil: %+v", val)
		}
		if err != nil {
			term.Println("==> validate error output")
			term.PrintError(err)
		}
	}
}
