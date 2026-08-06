package middleware

import (
	"net/http"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// SafeLog records request metadata without reading request bodies.
func SafeLog(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next(w, r)
		logx.WithContext(r.Context()).Infof(
			"[HTTP] %s - %s - %s - %s",
			r.Method,
			r.URL.RequestURI(),
			httpx.GetRemoteAddr(r),
			time.Since(start),
		)
	}
}
