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
	stack []string
	err   error
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
		eo.stack = append(eo.stack, op)
		return eo
	}
	return &Error{stack: []string{op}, err: err}
}

func (err *Error) Error() string {
	ops := make([]string, 0, len(err.stack))
	for i := len(err.stack) - 1; i >= 0; i-- {
		ops = append(ops, err.stack[i])
	}
	op := strings.Join(ops, ": ")
	return fmt.Sprintf("%s: %v", op, err.err)
}

func (err *Error) Extra() {
	if ext, ok := err.err.(Extra); ok {
		ext.Extra()
	}
}
