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

func Run(appLog, accessLog *logger.Logger) {
	conf := config.LoadConfig()

	dbPool, err := pgxpool.New(context.Background(), conf.DBPath())
	if err != nil {
		log.Fatal(err)
	}
	defer dbPool.Close()

	mux := http.NewServeMux()

	phttp.NewProfileRouter(mux, dbPool, "/api/v0", appLog, conf.UploadPath, conf.BaseURL)

	handler := middlewares.CorsMiddleware(middlewares.AccessLog(accessLog, mux))

	log.Println(fmt.Sprintf("Profile service running on %s", conf.AppPort))
	log.Fatal(http.ListenAndServe(fmt.Sprintf("0.0.0.0:%s", conf.AppPort), handler))
}
