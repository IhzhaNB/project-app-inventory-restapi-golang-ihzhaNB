package middleware

import (
	"inventory-system/utils"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// Logger middleware untuk log setiap HTTP request
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Catat waktu mulai
		start := time.Now()

		// Eksekusi handler
		next.ServeHTTP(w, r)

		// Hitung durasi & log
		duration := time.Since(start)

		utils.Logger.Info("HTTP Request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Duration("duration", duration),
			zap.String("ip", r.RemoteAddr),
			zap.String("user_agent", r.UserAgent()),
		)
	})
}
