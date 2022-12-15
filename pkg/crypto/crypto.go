package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	crand "crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"math/rand"
	"time"
	"unsafe"

	"github.com/fioncat/gitzombie/pkg/errors"
)

var (
	ErrIncorrectPassword = errors.New("incorrect password")
	ErrPasswordEmpty     = errors.New("password cannot be empty")
)

func initGCM(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, errors.Trace(err, "init aes cipher")
	}

	// gcm or Galois/Counter Mode, is a mode of operation for symmetric key
	// cryptographic block ciphers.
	// See: https://en.wikipedia.org/wiki/Galois/Counter_Mode
	gcm, err := cipher.NewGCM(block)
	return gcm, errors.Trace(err, "init gcm")
}

func encodePassword(password, salt string) []byte {
	password += salt

	sum := sha256.Sum256([]byte(password))
	return sum[:32]
}

func Encrypt(password, value string) (string, string, error) {
	if password == "" {
		return "", "", ErrPasswordEmpty
	}

	salt := genSalt()
	key := encodePassword(password, salt)

	gcm, err := initGCM(key)
	if err != nil {
		return "", "", err
	}

	// creates a new byte array the size of the nonce which must be passed to Seal
	nonce := make([]byte, gcm.NonceSize())

	// populates our nonce with a cryptographically secure random sequence
	_, err = io.ReadFull(crand.Reader, nonce)
	if err != nil {
		return "", "", errors.Trace(err, "generate random sequence")
	}

	result := gcm.Seal(nonce, nonce, []byte(value), nil)
	return hex.EncodeToString(result), salt, nil
}

func Decrypt(password, salt, value string) (string, error) {
	if password == "" {
		return "", ErrPasswordEmpty
	}

	key := encodePassword(password, salt)

	gcm, err := initGCM(key)
	if err != nil {
		return "", err
	}

	data, err := hex.DecodeString(value)
	if err != nil {
		return "", errors.New("value is not a hex string")
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("value is bad format")
	}

	var nonce []byte
	nonce, data = data[:nonceSize], data[nonceSize:]
	result, err := gcm.Open(nil, nonce, data, nil)
	if err != nil {
		return "", ErrIncorrectPassword
	}

	return string(result), nil
}

var (
	saltSrc = rand.NewSource(time.Now().UnixNano())
)

const (
	letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	saltLen = 16

	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func genSalt() string {
	b := make([]byte, saltLen)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := saltLen-1, saltSrc.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = saltSrc.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return *(*string)(unsafe.Pointer(&b))
}
