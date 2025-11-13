package cmd

import (
	"apple_backend/pkg/logger"
	"apple_backend/profile_service/internal/config"
	phttp "apple_backend/profile_service/internal/delivery/http"
	"apple_backend/profile_service/internal/delivery/middlewares"
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Run() {
	conf := config.LoadConfig()

	dbPool, err := pgxpool.New(context.Background(), conf.DBPath())
	if err != nil {
		log.Fatal(err)
	}
	defer dbPool.Close()

	mux := http.NewServeMux()

	protectedMux := http.NewServeMux()
	phttp.NewProfileRouter(protectedMux, dbPool, "/api/v0", conf.UploadPath, conf.BaseURL)

	jwtSecret := conf.JWTSecret
	protectedHandler := middlewares.AuthMiddleware(protectedMux, jwtSecret)
	mux.Handle("/api/v0/", protectedHandler)

	handler := middlewares.AccessLog(
		logger.Global(),
		middlewares.CorsMiddleware(
			middlewares.CSRFTokenMiddleware(
				middlewares.CSRFMiddleware(mux),
			),
		),
	)

	addr := fmt.Sprintf("0.0.0.0:%s", conf.AppPort)
	log.Printf("Profile service running on http://%s:%s", conf.AppHost, conf.AppPort)
	log.Printf("Avatar URL base: %s", conf.BaseURL)
	log.Fatal(http.ListenAndServe(addr, handler))
}
