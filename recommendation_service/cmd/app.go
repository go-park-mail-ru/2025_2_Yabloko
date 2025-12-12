package cmd

import (
	"apple_backend/pkg/logger"
	"apple_backend/pkg/metrics"
	"apple_backend/pkg/middlewares"
	"apple_backend/recommendation_service/internal/config"
	shttp "apple_backend/recommendation_service/internal/delivery/http"
	"apple_backend/recommendation_service/internal/repository"
	"apple_backend/recommendation_service/internal/usecase"
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Run() {
	conf := config.MustConfig()
	apiV0Prefix := "/api/v0/"

	dbPool, err := pgxpool.New(context.Background(), conf.DBPath())
	if err != nil {
		log.Fatal(err)
	}
	defer dbPool.Close()

	recRepo := repository.NewRecommendationRepoPostgres(dbPool)
	recUC := usecase.NewRecommendationUsecase(recRepo)
	recHandler := shttp.NewRecommendationHandler(recUC)

	mux := http.NewServeMux()
	shttp.NewRecommendationRouter(mux, recHandler, apiV0Prefix)

	handler := middlewares.AccessLog(
		logger.Global(),
		middlewares.CorsMiddleware(mux),
	)

	metricsHandler := metrics.HTTPMetricsMiddleware("recommendation_service", handler)

	rootMux := http.NewServeMux()
	rootMux.Handle("/metrics", promhttp.Handler())
	rootMux.Handle("/", metricsHandler)

	addr := fmt.Sprintf("0.0.0.0:%s", conf.AppPort)
	log.Printf("Recommendation service running on http://localhost:%s", conf.AppPort)
	log.Fatal(http.ListenAndServe(addr, rootMux))
}
