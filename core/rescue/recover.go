package rescue

import "gitlab.deepwisdomai.com/infra/go-zero/core/logx"

// Recover is used with defer to do cleanup on panics.
// Use it like:
//  defer Recover(func() {})
func Recover(cleanups ...func()) {
	for _, cleanup := range cleanups {
		cleanup()
	}

	if p := recover(); p != nil {
		logx.ErrorStack(p)
	}
}
