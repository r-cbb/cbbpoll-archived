package errors

import "fmt"

type Op string

type Code uint

type Error struct {
	Op   Op
	Kind Code
	Err  error
	Msg  string
}

const (
	KindUnexpected Code = iota // zero type is purposefully KindUnexpected
	KindNotImplemented
	KindNotFound
	KindConcurrencyProblem
	KindDatabaseError
	KindJWTError
	KindAuthError
	KindServiceUnavailable
)

func (e Error) Error() string {
	if e.Msg != "" {
		return fmt.Sprintf("%s: %s", e.Msg, e.Err.Error())
	}

	return e.Err.Error()
}

func Kind(err error) Code {
	e, ok := err.(*Error)
	if !ok {
		return KindUnexpected
	}

	if e.Kind != 0 {
		return e.Kind
	}

	return Kind(e.Err)
}

func E(args ...interface{}) error {
	e := Error{}
	for _, arg := range args {
		switch arg := arg.(type) {
		case Op:
			e.Op = arg
		case Code:
			e.Kind = arg
		case error:
			e.Err = arg
		default:
			panic("bad call to E")
		}
	}

	return &e
}

func Ops(e *Error) []Op {
	res := []Op{e.Op}

	subErr, ok := e.Err.(*Error)
	if !ok {
		return res
	}

	res = append(res, Ops(subErr)...)

	return res
}
