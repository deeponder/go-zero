package httpx

import (
	"encoding/json"
	"github.com/pkg/errors"
	"net/http"
	"sync"

	"gitlab.deepwisdomai.com/infra/go-zero/core/logx"
)

var (
	errorHandler func(error) (int, interface{})
	lock         sync.RWMutex
)

// default render
type DefaultRender struct {
	ErrorCode int         `json:"error_code"`
	ErrorMsg  string      `json:"error_msg"`
	Result    interface{} `json:"result"`
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

	code, body := errorHandler(err)
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
	var statusCode int

	switch errors.Cause(err).(type) {
	case ErrorJson:
		renderJson.ErrorCode = errors.Cause(err).(ErrorJson).ErrNo
		renderJson.ErrorMsg = errors.Cause(err).(ErrorJson).ErrMsg
		renderJson.Result = nil

		statusCode = errors.Cause(err).(ErrorJson).StatusCode
	default:
		renderJson.ErrorCode = -1
		renderJson.ErrorMsg = errors.Cause(err).Error()
		renderJson.Result = nil

		statusCode = http.StatusOK
	}
	WriteJson(w, statusCode, renderJson)

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
