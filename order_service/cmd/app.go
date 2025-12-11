package cmd

import (
	"apple_backend/order_service/internal/config"
	shttp "apple_backend/order_service/internal/delivery/http"
	"apple_backend/order_service/internal/infrastructure/yookassa"
	"apple_backend/order_service/internal/repository"
	"apple_backend/order_service/internal/usecase"
	"apple_backend/pkg/blacklist"
	"apple_backend/pkg/logger"
	"apple_backend/pkg/metrics"
	"apple_backend/pkg/middlewares"
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
)

func Run() {
	conf := config.MustConfig()
	apiV0Prefix := "/api/v0/"

	dbPool, err := pgxpool.New(context.Background(), conf.DBPath())
	if err != nil {
		log.Fatal(err)
	}
	defer dbPool.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: conf.RedisURL,
	})
	defer redisClient.Close()

	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.Fatal("Redis connection failed:", err)
	}

	tokenBlacklist := blacklist.NewRedisTokenBlacklist(redisClient)

	openMux := http.NewServeMux()
	// TODO Cut fake handler
	fakeHandler := shttp.NewFakePaymentHandler()
	openMux.HandleFunc(apiV0Prefix+"fake-payment", fakeHandler.FakePayment)

	paymentRepo := repository.NewPaymentRepoPostgres(dbPool)
	orderRepo := repository.NewOrderRepoPostgres(dbPool)
	yookassaClient := yookassa.NewClient(conf.YookassaBaseURL, conf.YookassaShopID, conf.YookassaSecret)
	paymentUC := usecase.NewPaymentUsecase(paymentRepo, orderRepo, yookassaClient)
	paymentHandler := shttp.NewPaymentHandler(paymentUC, conf.YookassaSecret)

	protectedMux := http.NewServeMux()
	shttp.NewOrderRouter(protectedMux, dbPool, apiV0Prefix)
	shttp.NewPaymentRouter(protectedMux, tokenBlacklist, paymentHandler, conf.JWTSecret, apiV0Prefix)

	protectedHandler := middlewares.AuthMiddleware(tokenBlacklist, conf.JWTSecret, logger.Global())

	mux := http.NewServeMux()

	mux.HandleFunc(apiV0Prefix+"payments/webhook", paymentHandler.HandleWebhook)

	mux.Handle(apiV0Prefix+"orders", protectedHandler(protectedMux))
	mux.Handle(apiV0Prefix+"orders/", protectedHandler(protectedMux))
	mux.Handle(apiV0Prefix+"payments", protectedHandler(protectedMux))
	mux.Handle(apiV0Prefix+"payments/", protectedHandler(protectedMux))

	mux.Handle(apiV0Prefix, openMux)

	handler := middlewares.AccessLog(
		logger.Global(),
		middlewares.CorsMiddleware(mux),
	)

	metricsHandler := metrics.HTTPMetricsMiddleware("order_service", handler)

	rootMux := http.NewServeMux()
	rootMux.Handle("/metrics", promhttp.Handler())
	rootMux.Handle("/", metricsHandler)

	addr := fmt.Sprintf("0.0.0.0:%s", conf.AppPort)
	log.Printf("Order service running on http://localhost:%s", conf.AppPort)
	log.Fatal(http.ListenAndServe(addr, rootMux))
}
