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
	"path/filepath"

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

	shttp.NewStoreRouter(openMux, dbPool, apiV0Prefix, appLog)
	shttp.NewItemRouter(openMux, dbPool, apiV0Prefix, appLog)

	// ← ДОБАВЛЕНО: PaymentHandler для заглушки
	paymentHandler := shttp.NewPaymentHandler(appLog)
	openMux.HandleFunc(apiV0Prefix+"fake-payment", paymentHandler.FakePayment)

	// ЗАЩИЩЁННЫЕ ручки
	shttp.NewCartRouter(protectedMux, dbPool, apiV0Prefix, appLog)
	shttp.NewOrderRouter(protectedMux, dbPool, apiV0Prefix, appLog)

	protectedHandler := middlewares.AuthMiddleware(protectedMux, conf.JWTSecret)

	mux := http.NewServeMux()

	mux.Handle("/images/items/", http.StripPrefix("/images/items/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fullPath := filepath.Join(conf.UploadItemDir, r.URL.Path)
		appLog.Debug("Serving item image",
			"url_path", r.URL.Path,
			"fs_path", fullPath,
		)
		http.ServeFile(w, r, fullPath)
	})))

	mux.Handle("/images/stores/", http.StripPrefix("/images/stores/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fullPath := filepath.Join(conf.UploadStoreDir, r.URL.Path)
		appLog.Debug("Serving store image",
			"path", r.URL.Path,
			"file", fullPath,
		)
		http.ServeFile(w, r, fullPath)
	})))

	mux.Handle(apiV0Prefix+"cart", protectedHandler)
	mux.Handle(apiV0Prefix+"orders", protectedHandler)
	mux.Handle(apiV0Prefix+"orders/", protectedHandler)
	mux.Handle(apiV0Prefix, openMux)

	handler := middlewares.CorsMiddleware(middlewares.AccessLog(accessLog, mux))

	log.Println(fmt.Sprintf("Store service running on %s", conf.AppPort))
	log.Fatal(http.ListenAndServe(fmt.Sprintf("0.0.0.0:%s", conf.AppPort), handler))
}
