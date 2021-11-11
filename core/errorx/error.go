package errorx

import (
	"fmt"
)

type IError interface {
	error

	GetCode() int
}

type CustomError struct {
	error
	code int
	data interface{}
}

func (c CustomError) Error() string {
	return c.error.Error()
}

func (c CustomError) GetCode() int {
	return c.code
}

func NewError(code int, err error) IError {
	return &CustomError{
		code:  code,
		error: err,
	}
}

func NewErrorF(code int, format string, args ...interface{}) IError {
	return &CustomError{
		error: fmt.Errorf(format, args...),
		code:  code,
	}
}
