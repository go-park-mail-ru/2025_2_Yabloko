package cmd

import (
	"apple_backend/auth_service/internal/config"
	authhttp "apple_backend/auth_service/internal/delivery/http"
	authmw "apple_backend/auth_service/internal/delivery/middlewares"
	"apple_backend/auth_service/internal/repository"
	"apple_backend/auth_service/internal/usecase"
	"apple_backend/pkg/logger"
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

func csrfHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Run запускает auth_service без передачи логгеров — всё использует logger.Global()
func Run() {
	conf := config.LoadConfig()

	dbPool, err := pgxpool.New(context.Background(), conf.DBPath())
	if err != nil {
		log.Fatal(err)
	}
	defer dbPool.Close()

	repo := repository.NewAuthRepoPostgres(dbPool)
	uc := usecase.NewAuthUseCase(repo, conf.SecretKeyStr())

	authMux := http.NewServeMux()
	authMux.Handle("/csrf", http.HandlerFunc(csrfHandler))
	authhttp.NewAuthRouter(authMux, "/auth", uc)

	authHandler := authmw.CSRFTokenMiddleware(
		authmw.CSRFMiddleware(authMux),
	)

	mainMux := http.NewServeMux()
	mainMux.Handle("/api/v0/", http.StripPrefix("/api/v0", authHandler))

	handler := authmw.CorsMiddleware(
		authmw.AccessLog(logger.Global(), mainMux),
	)

	addr := fmt.Sprintf("0.0.0.0:%s", conf.AppPortStr())
	log.Println("Auth service running on", addr)
	log.Fatal(http.ListenAndServe(addr, handler))
}
