package middleware

import (
	"net/http"

	"github.com/SameerJadav/syncstream/internal/logger"
)

func LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info.Printf("%s %s %s %s\n", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI())

		next.ServeHTTP(w, r)
	})
}
