package httpx

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/pkg/errors"
	"gitlab.deepwisdomai.com/infra/go-zero/core/errorx"

	"gitlab.deepwisdomai.com/infra/go-zero/core/logx"
)

var (
	errorHandler func(error) (int, interface{})
	lock         sync.RWMutex
)

// default render
type DefaultRender struct {
	ErrNo  int         `json:"errNo"`
	ErrMsg string      `json:"errMsg"`
	Data   interface{} `json:"data"`
}

// Error writes err into w.
func Error(w http.ResponseWriter, err error) {
	lock.RLock()
	handler := errorHandler
	lock.RUnlock()

	if handler == nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	code, body := handler(err)
	e, ok := body.(error)
	if ok {
		http.Error(w, e.Error(), code)
	} else {
		WriteJson(w, code, body)
	}
}

// Ok writes HTTP 200 OK into w.
func Ok(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
}

// OkJson writes v into w with 200 OK.
func OkJson(w http.ResponseWriter, v interface{}) {
	renderJson := DefaultRender{0, "succ", v}
	WriteJson(w, http.StatusOK, renderJson)
}

func FailJson(w http.ResponseWriter, err error) {
	var renderJson DefaultRender

	switch errors.Cause(err).(type) {
	case ErrorJson:
		renderJson.ErrNo = errors.Cause(err).(ErrorJson).ErrNo
		renderJson.ErrMsg = errors.Cause(err).(ErrorJson).ErrMsg
		renderJson.Data = nil
	default:
		renderJson.ErrNo = -1
		renderJson.ErrMsg = errors.Cause(err).Error()
		renderJson.Data = nil
	}
	WriteJson(w, http.StatusOK, renderJson)

	// 打印错误栈
	//StackLogger(ctx, err)
	return
}

// SetErrorHandler sets the error handler, which is called on calling Error.
func SetErrorHandler(handler func(error) (int, interface{})) {
	lock.Lock()
	defer lock.Unlock()
	errorHandler = handler
}

// WriteJson writes v as json string into w with code.
func WriteJson(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set(ContentType, ApplicationJson)
	w.WriteHeader(code)

	if bs, err := json.Marshal(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else if n, err := w.Write(bs); err != nil {
		// http.ErrHandlerTimeout has been handled by http.TimeoutHandler,
		// so it's ignored here.
		if err != http.ErrHandlerTimeout {
			logx.Errorf("write response failed, error: %s", err)
		}
	} else if n < len(bs) {
		logx.Errorf("actual bytes: %d, written bytes: %d", len(bs), n)
	}
}

func FailJsonWithCode(w http.ResponseWriter, err error) {
	code := 0
	if cErr, ok := err.(errorx.IError); ok {
		code = cErr.GetCode()
	}
	WriteJson(w, http.StatusOK, map[string]interface{}{
		"code":    code,
		"message": err.Error(),
	})
}
