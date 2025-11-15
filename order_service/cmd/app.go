package cmd

import (
	"apple_backend/order_service/internal/config"
	shttp "apple_backend/order_service/internal/delivery/http"
	"apple_backend/order_service/internal/delivery/middlewares"
	"apple_backend/pkg/logger"
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Run() {
	conf := config.MustConfig()
	apiV0Prefix := "/api/v0/"

	dbPool, err := pgxpool.New(context.Background(), conf.DBPath())
	if err != nil {
		log.Fatal(err)
	}
	defer dbPool.Close()

	openMux := http.NewServeMux()
	// TODO Cut fake handler
	fakeHandler := shttp.NewFakePaymentHandler()
	openMux.HandleFunc(apiV0Prefix+"fake-payment", fakeHandler.FakePayment)

	protectedMux := http.NewServeMux()
	shttp.NewOrderRouter(protectedMux, dbPool, apiV0Prefix)
	shttp.NewPaymentRouter(protectedMux, dbPool, conf, apiV0Prefix)

	protectedHandler := middlewares.AuthMiddleware(protectedMux, conf.JWTSecret)

	mux := http.NewServeMux()
	mux.Handle(apiV0Prefix+"orders", protectedHandler)
	mux.Handle(apiV0Prefix+"orders/", protectedHandler)
	mux.Handle(apiV0Prefix+"payments", protectedHandler)
	mux.Handle(apiV0Prefix+"payments/", protectedHandler)
	mux.Handle(apiV0Prefix, openMux)

	handler := middlewares.AccessLog(
		logger.Global(),
		middlewares.CorsMiddleware(mux),
	)

	addr := fmt.Sprintf("0.0.0.0:%s", conf.AppPort)
	log.Printf("Order service running on http://localhost:%s", conf.AppPort)
	log.Fatal(http.ListenAndServe(addr, handler))
}
