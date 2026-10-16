package httpapp

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jzxq45/golang-todoapp/internal/httpx"
	custommiddleware "github.com/jzxq45/golang-todoapp/internal/middleware"
	"github.com/jzxq45/golang-todoapp/internal/task"
	frontend "github.com/jzxq45/golang-todoapp/internal/web"
)

func New(handler *task.Handler, logger *slog.Logger) *chi.Mux {
	router := chi.NewRouter()
	router.Use(chimiddleware.RequestID)
	router.Use(chimiddleware.RealIP)
	router.Use(custommiddleware.RequestID)
	router.Use(custommiddleware.Recoverer)
	router.Use(custommiddleware.Timeout(30 * time.Second))
	router.Use(custommiddleware.Logger(logger))
	router.Use(custommiddleware.Logging)
	router.Use(custommiddleware.CORS)
	router.Use(contentSecurityPolicy)

	router.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		httpx.OK(w, map[string]string{"status": "ok"})
	})
	handler.RegisterRoutes(router)
	router.Handle("/static/*", frontend.Handler())
	router.NotFound(frontend.Handler().ServeHTTP)
	return router
}

func contentSecurityPolicy(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data:; font-src 'self'; connect-src 'self'; object-src 'none'; base-uri 'none'; frame-ancestors 'none'")
		next.ServeHTTP(w, r)
	})
}
