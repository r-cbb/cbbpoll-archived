package errors

type Error struct {
	Op   Op
	Kind Code
	Err  error
}

type Op string

type Code uint

const (
	KindUnexpected Code = iota // zero type is purposefully KindUnexpected
	KindNotFound
)


func (e Error) Error() string {
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