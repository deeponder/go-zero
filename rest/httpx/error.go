package httpx

import (
	"fmt"

	"github.com/pkg/errors"
)

type ErrorJson struct {
	ErrNo  int
	ErrMsg string
}

func NewBaseError(code int, message string) *ErrorJson {
	return &ErrorJson{
		ErrNo:  code,
		ErrMsg: message,
	}
}

func NewError(code int, message, userMsg string) ErrorJson {
	return ErrorJson{
		ErrNo:  code,
		ErrMsg: message,
	}
}

func (err ErrorJson) Error() string {
	return err.ErrMsg
}

func (err ErrorJson) Sprintf(v ...interface{}) ErrorJson {
	err.ErrMsg = fmt.Sprintf(err.ErrMsg, v...)
	return err
}

func (err ErrorJson) Equal(e error) bool {
	switch errors.Cause(err).(type) {
	case ErrorJson:
		return err.ErrNo == errors.Cause(err).(ErrorJson).ErrNo
	default:
		return false
	}
}

func (err ErrorJson) WrapPrint(core error, message string, user ...interface{}) error {
	if core == nil {
		return nil
	}
	ret := err
	SetErrPrintfMsg(&ret, core)
	return errors.Wrap(ret, message)
}

func (err ErrorJson) WrapPrintf(core error, format string, message ...interface{}) error {
	if core == nil {
		return nil
	}
	ret := err
	SetErrPrintfMsg(&ret, core)
	return errors.Wrap(ret, fmt.Sprintf(format, message...))
}

func (err ErrorJson) Wrap(core error) error {
	if core == nil {
		return nil
	}

	msg := err.ErrMsg
	err.ErrMsg = core.Error()
	return errors.Wrap(err, msg)
}

func SetErrPrintfMsg(err *ErrorJson, v ...interface{}) {
	err.ErrMsg = fmt.Sprintf(err.ErrMsg, v...)
}
