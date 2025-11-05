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
	conf := config.MustConfig()
	apiV0Prefix := "/api/v0/"
	dbPool, err := pgxpool.New(context.Background(), conf.DBPath())
	if err != nil {
		log.Fatal(err)
	}
	defer dbPool.Close()

	openMux := http.NewServeMux()

	// ПУБЛИЧНЫЕ ручки
	shttp.NewStoreRouter(openMux, dbPool, apiV0Prefix, appLog)
	shttp.NewItemRouter(openMux, dbPool, apiV0Prefix, appLog) // ВАЖНО: регистрация items

	protectedMux := http.NewServeMux()

	shttp.NewCartRouter(protectedMux, dbPool, apiV0Prefix, appLog)
	shttp.NewOrderRouter(protectedMux, dbPool, apiV0Prefix, appLog)

	protectedHandler := middlewares.AuthMiddleware(protectedMux, conf.JWTSecret)

	mux := http.NewServeMux()

	mux.Handle(apiV0Prefix+"cart", protectedHandler)    // /api/v0/cart
	mux.Handle(apiV0Prefix+"orders", protectedHandler)  // /api/v0/orders
	mux.Handle(apiV0Prefix+"orders/", protectedHandler) // ВАЖНО: поддерево /api/v0/orders/*
	mux.Handle(apiV0Prefix, openMux)                    // /api/v0/* (stores, items, cities, tags)

	handler := middlewares.CorsMiddleware(middlewares.AccessLog(accessLog, mux))

	log.Println(fmt.Sprintf("Store service running on %s", conf.AppPort))
	log.Fatal(http.ListenAndServe(fmt.Sprintf("0.0.0.0:%s", conf.AppPort), handler))
}
