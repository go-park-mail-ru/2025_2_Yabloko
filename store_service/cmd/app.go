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

func Run(appLog, accessLog logger.Logger) {
	conf := config.MustConfig()
	apiV0Prefix := "/api/v0/"
	dbPool, err := pgxpool.New(context.Background(), conf.DBPath())
	if err != nil {
		log.Fatal(err)
	}
	defer dbPool.Close()

	openMux := http.NewServeMux()
	protectedMux := http.NewServeMux()

	// ПУБЛИЧНЫЕ ручки
	shttp.NewStoreRouter(openMux, dbPool, apiV0Prefix, appLog)
	shttp.NewItemRouter(openMux, dbPool, apiV0Prefix, appLog)
	shttp.NewBaseRouter(openMux, appLog, dbPool, apiV0Prefix, conf.ImageDir)

	// ЗАЩИЩЁННЫЕ ручки
	shttp.NewCartRouter(protectedMux, dbPool, apiV0Prefix, appLog)
	shttp.NewOrderRouter(protectedMux, dbPool, apiV0Prefix, appLog)

	protectedHandler := middlewares.AuthMiddleware(protectedMux, conf.JWTSecret)

	mux := http.NewServeMux()

	mux.Handle(apiV0Prefix+"cart", protectedHandler)
	mux.Handle(apiV0Prefix+"orders", protectedHandler)
	mux.Handle(apiV0Prefix+"orders/", protectedHandler)
	mux.Handle(apiV0Prefix, openMux)

	handler := middlewares.CorsMiddleware(middlewares.AccessLog(accessLog, mux))

	log.Println(fmt.Sprintf("Store service running on %s", conf.AppPort))
	log.Fatal(http.ListenAndServe(fmt.Sprintf("0.0.0.0:%s", conf.AppPort), handler))
}
