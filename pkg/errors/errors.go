package errors

import (
	"errors"
	"fmt"
	"strings"
)

func New(msg string) error {
	return errors.New(msg)
}

type Error struct {
	ops []string
	err error
}

type Extra interface {
	Extra()
}

func Trace(err error, op string, args ...any) error {
	if err == nil {
		return nil
	}
	op = fmt.Sprintf(op, args...)
	if eo, ok := err.(*Error); ok {
		eo.ops = append(eo.ops, op)
		return eo
	}
	return &Error{ops: []string{op}, err: err}
}

func (err *Error) Error() string {
	op := strings.Join(err.ops, ": ")
	return fmt.Sprintf("%s: %v", op, err.err)
}

func (err *Error) Extra() {
	if ext, ok := err.err.(Extra); ok {
		ext.Extra()
	}
}
