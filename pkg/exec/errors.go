package exec

type Error struct {
	Cmd string

	Stdout string
	Stderr string

	Err error
}

func (e *Error) Error() string {
	return e.Err.Error()
}
