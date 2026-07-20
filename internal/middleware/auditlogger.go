package middleware

import (
	"log/slog"
	"net/http"
)

func AuditLogger(next http.HandlerFunc) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		slog.Info("Audit Report", "method", r.Method, "path", r.URL.Path)

		next(w, r)
	}

}
