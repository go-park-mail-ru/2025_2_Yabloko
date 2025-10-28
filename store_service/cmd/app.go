package cmd

import (
	"apple_backend/pkg/logger"
	"apple_backend/store_service/internal/config"
	shttp "apple_backend/store_service/internal/delivery/http"
	"apple_backend/store_service/internal/delivery/middlewares"
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

	shttp.NewStoreRouter(mux, dbPool, "/api/v0", appLog)
	shttp.NewItemRouter(mux, dbPool, "/api/v0", appLog)

	handler := middlewares.CorsMiddleware(middlewares.AccessLog(accessLog, mux))

	log.Println(fmt.Sprintf("Store service running on %s", conf.AppPort))
	log.Fatal(http.ListenAndServe(fmt.Sprintf("0.0.0.0:%s", conf.AppPort), handler))
}
