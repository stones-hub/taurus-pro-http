package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/stones-hub/taurus-pro-http/pkg/httpx"
)

// ErrorLoggerHandler 错误处理函数
type ErrorLoggerHandler func(err any, stack string)

func RecoveryMiddleware(fn ErrorLoggerHandler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					stack := debug.Stack()
					fn(err, string(stack))
					httpx.SendResponse(w, http.StatusInternalServerError, "Internal Server Error", nil)
					return
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
