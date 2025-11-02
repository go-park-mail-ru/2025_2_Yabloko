package cmd

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	authhttp "apple_backend/auth_service/internal/delivery/http"
	authmw "apple_backend/auth_service/internal/delivery/middlewares"
	"apple_backend/auth_service/internal/repository"
	"apple_backend/auth_service/internal/usecase"
	"apple_backend/pkg/logger"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Run() {
	appLog := logger.NewLogger("auth_service", slog.LevelInfo)

	dbURL := getenv("DB_URL", "")
	if dbURL == "" {
		log.Fatal("DB_URL is not set")
	}
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	secret := os.Getenv("SECRET_KEY")
	if secret == "" {
		log.Fatal("SECRET_KEY is not set")
	}

	repo := repository.NewAuthRepoPostgres(pool)
	uc := usecase.NewAuthUseCase(repo, secret)

	mux := http.NewServeMux()

	// вариант 1: использовать конструктор роутера
	authhttp.NewAuthRouter(mux, "/api/v0", appLog, uc)

	// вариант 2: оставить ручную регистрацию (если хотите — закомментированно)
	// h := authhttp.NewAuthHandler(uc, appLog) // исправлено: передаём логгер
	// mux.Handle("/api/v0/auth/signup", authmw.RateLimit(5, time.Minute)(http.HandlerFunc(h.Register)))
	// mux.Handle("/api/v0/auth/login", authmw.RateLimit(10, time.Minute)(http.HandlerFunc(h.Login)))
	// mux.Handle("/api/v0/auth/refresh", authmw.CSRFMiddleware(http.HandlerFunc(h.RefreshToken)))
	// mux.Handle("/api/v0/auth/logout", authmw.CSRFMiddleware(http.HandlerFunc(h.Logout)))

	root := authmw.AccessLog(appLog, // исправлено: не nil
		authmw.CorsMiddleware(
			authmw.CSRFTokenMiddleware(mux),
		),
	)

	addr := ":" + getenv("PORT", "8080")
	srv := &http.Server{
		Addr:         addr,
		Handler:      root,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	log.Println("auth_service listening on", addr)
	log.Fatal(srv.ListenAndServe())
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
