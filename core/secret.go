package core

import (
	"fmt"
	"os"

	"github.com/fioncat/gitzombie/config"
	"github.com/fioncat/gitzombie/pkg/crypto"
	"github.com/fioncat/gitzombie/pkg/errors"
	"github.com/fioncat/gitzombie/pkg/osutil"
	"github.com/fioncat/gitzombie/pkg/term"
	"gopkg.in/yaml.v3"
)

type Secrets map[string]*Secret

type Secret struct {
	Value string `yaml:"value"`
	Salt  string `yaml:"salt"`
}

func readSecrets() (Secrets, error) {
	path := config.GetDir("secrets")
	exists, err := osutil.FileExists(path)
	if err != nil {
		return nil, errors.Trace(err, "check secrets file exists")
	}
	if !exists {
		return Secrets{}, nil
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, errors.Trace(err, "open secrets file")
	}
	defer file.Close()

	s := Secrets{}
	err = yaml.NewDecoder(file).Decode(&s)
	if err != nil {
		return nil, errors.Trace(err, "parse secrets yaml")
	}

	return s, nil
}

func writeSecrets(s Secrets) error {
	path := config.GetDir("secrets")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return errors.Trace(err, "open secrets file")
	}
	defer file.Close()

	encoder := yaml.NewEncoder(file)
	encoder.SetIndent(2)
	err = encoder.Encode(s)
	return errors.Trace(err, "encode secrets yaml")
}

func SetSecret(key string) (string, error) {
	secrets, err := readSecrets()
	if err != nil {
		return "", err
	}

	password, err := term.InputNewPassword("Please intput new password for %s", key)
	if err != nil {
		return "", err
	}
	value := term.InputErase("Please input %s", key)

	encrypted, salt, err := crypto.Encrypt(password, []byte(value), false)
	if err != nil {
		return "", errors.Trace(err, "encrypt")
	}

	secrets[key] = &Secret{
		Value: encrypted,
		Salt:  salt,
	}

	return value, writeSecrets(secrets)
}

func GetSecret(key string, allowCreate bool) (string, error) {
	secrets, err := readSecrets()
	if err != nil {
		return "", err
	}

	s, ok := secrets[key]
	if !ok {
		if allowCreate {
			ok = term.Confirm("The secret %q does not exists, do you want to create it", key)
			if ok {
				return SetSecret(key)
			}
		}
		return "", fmt.Errorf("cannot find secret %q", key)
	}
	password, err := term.InputPassword("Please input password for %s", key)
	if err != nil {
		return "", err
	}

	data, err := crypto.Decrypt(password, s.Salt, s.Value)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func DeleteSecret(key string) error {
	secrets, err := readSecrets()
	if err != nil {
		return err
	}
	delete(secrets, key)
	return writeSecrets(secrets)
}
