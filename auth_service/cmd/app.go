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

func Run(appLog, accessLog logger.Logger) {
	if appLog == nil {
		log.Fatal("app log is nil")
	}
	if accessLog == nil {
		log.Fatal("access log is nil")
	}

	conf := config.LoadConfig()

	dbPool, err := pgxpool.New(context.Background(), conf.DBPath())
	if err != nil {
		log.Fatal(err)
	}
	defer dbPool.Close()

	repo := repository.NewAuthRepoPostgres(dbPool)
	uc := usecase.NewAuthUseCase(repo, conf.SecretKeyStr())

	mainMux := http.NewServeMux()

	authhttp.NewAuthRouter(mainMux, "/api/v0", appLog, uc)

	handler := authmw.CorsMiddleware(
		authmw.AccessLog(accessLog,
			authmw.SessionMiddleware(
				//authmw.SmartCSRFMiddleware(
				mainMux,
				// ),
			),
		),
	)

	addr := fmt.Sprintf("0.0.0.0:%s", conf.AppPortStr())
	log.Println("Auth service running on", addr)
	log.Fatal(http.ListenAndServe(addr, handler))
}
