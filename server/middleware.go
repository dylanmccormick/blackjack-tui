package server

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/http"
	"time"
)

type Middleware func(http.Handler) http.Handler

type responseWriterWithStatus struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriterWithStatus) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := rw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("responsewriter does not support hijacking")
	}
	return hijacker.Hijack()
}

func (rw *responseWriterWithStatus) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionId := r.URL.Query().Get("session")
		session, err := s.SessionManager.GetSession(sessionId)
		if err != nil {
			s.log.Error("error with session", "error", err, "request_id", r.Context().Value("requestId"))
			http.Error(w, "Authentication Required", http.StatusUnauthorized)
			return
		}
		if !session.Authenticated {
			http.Error(w, "Authentication Required", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), "session", session)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) panicRecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				s.log.Error("PANIC recovered",
					"error", err,
					"path", r.URL.Path,
					"method", r.Method,
					"request_id", r.Context().Value("requestId"),
				)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestId := generateId()
		s.log.Info("[Request Start]", "method", r.Method, "path", r.URL.Path, "request_id", requestId)
		ctx := context.WithValue(r.Context(), "request_id", requestId)
		reqWithCtx := r.WithContext(ctx)
		next.ServeHTTP(w, reqWithCtx)
		duration := time.Since(start).Seconds()
		s.log.Info("[Request End]", "method", r.Method, "path", r.URL.Path, "duration", duration, "request_id", requestId)
	})
}

func (s *Server) metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		s.Metrics.HTTPRequestsInFlight.Inc()
		defer s.Metrics.HTTPRequestsInFlight.Dec()

		wrappedWriter := &responseWriterWithStatus{ResponseWriter: w, statusCode: 200}
		next.ServeHTTP(wrappedWriter, r)

		duration := time.Since(start).Seconds()
		method := r.Method
		path := r.URL.Path
		status := fmt.Sprintf("%d", wrappedWriter.statusCode)
		s.Metrics.HTTPRequestsDuration.WithLabelValues(method, path).Observe(duration)
		s.Metrics.HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
	})
}

func chainMiddleware(handler http.Handler, middlewares ...Middleware) http.Handler {
	// Apply middleware in REVERSE order so they execute in the order listed
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}
