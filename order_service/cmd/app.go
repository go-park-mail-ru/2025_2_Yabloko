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
	protectedMux := http.NewServeMux()

	shttp.NewOrderRouter(protectedMux, dbPool, apiV0Prefix)

	paymentHandler := shttp.NewPaymentHandler()
	openMux.HandleFunc(apiV0Prefix+"fake-payment", paymentHandler.FakePayment)

	protectedHandler := middlewares.AuthMiddleware(protectedMux, conf.JWTSecret)

	mux := http.NewServeMux()
	mux.Handle(apiV0Prefix+"orders", protectedHandler)
	mux.Handle(apiV0Prefix+"orders/", protectedHandler)
	mux.Handle(apiV0Prefix, openMux)

	// middleware цепочка
	handler := middlewares.AccessLog(
		logger.Global(),
		middlewares.CorsMiddleware(mux),
	)

	addr := fmt.Sprintf("0.0.0.0:%s", conf.AppPort)
	log.Printf("Store service running on http://localhost:%s", conf.AppPort)
	log.Fatal(http.ListenAndServe(addr, handler))
}
