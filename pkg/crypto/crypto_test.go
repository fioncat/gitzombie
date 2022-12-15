package crypto

import (
	"fmt"
	"testing"
)

func TestCrypto(t *testing.T) {
	var testCases = []struct {
		password string
		value    string
	}{
		{
			password: "test123",
			value:    "hello",
		},
		{
			password: "test666",
			value:    "111",
		},
		{
			password: "a",
			value:    "This is a very new message",
		},
		{
			password: "b",
			value:    "This is a secret message",
		},
	}
	for _, testCase := range testCases {
		encrypted, salt, err := Encrypt(testCase.password, []byte(testCase.value), false)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Printf("encrypted: %s, salt: %s\n", encrypted, salt)
		raw, err := Decrypt(testCase.password, salt, encrypted)
		if err != nil {
			t.Fatal(err)
		}
		if string(raw) != testCase.value {
			t.Fatalf("incorrect decrypt result: %q, expect %q", raw, testCase.value)
		}

		badPassword := "bad password"
		_, err = Decrypt(badPassword, salt, encrypted)
		if err != ErrIncorrectPassword {
			t.Fatal("expect incorrect password")
		}
	}
}
