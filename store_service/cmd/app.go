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

	// все роутеры без передачи логгера
	shttp.NewStoreRouter(openMux, dbPool, apiV0Prefix)
	shttp.NewItemRouter(openMux, dbPool, apiV0Prefix)
	shttp.NewCartRouter(protectedMux, dbPool, apiV0Prefix)
	shttp.NewOrderRouter(protectedMux, dbPool, apiV0Prefix)

	paymentHandler := shttp.NewPaymentHandler()
	openMux.HandleFunc(apiV0Prefix+"fake-payment", paymentHandler.FakePayment)

	protectedHandler := middlewares.AuthMiddleware(protectedMux, conf.JWTSecret)

	mux := http.NewServeMux()

	// статика для изображений товаров
	mux.Handle("/images/items/", http.StripPrefix("/images/items/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fullPath := filepath.Join(conf.UploadItemDir, r.URL.Path)
		logger.Global().Debug("Serving item image",
			"url_path", r.URL.Path,
			"fs_path", fullPath,
		)
		http.ServeFile(w, r, fullPath)
	})))

	// статика для изображений магазинов
	mux.Handle("/images/stores/", http.StripPrefix("/images/stores/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fullPath := filepath.Join(conf.UploadStoreDir, r.URL.Path)
		logger.Global().Debug("Serving store image",
			"path", r.URL.Path,
			"file", fullPath,
		)
		http.ServeFile(w, r, fullPath)
	})))

	// маршрутизация API
	mux.Handle(apiV0Prefix+"cart", protectedHandler)
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
