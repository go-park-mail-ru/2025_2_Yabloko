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

	//embeddingClient, err := client.NewGRPCEmbeddingClient(conf.EmbeddingServiceAddr, logger.Global())
	//if err != nil {
	//	logger.Global().Warn("embedding service unavailable, using noop",
	//		"addr", conf.EmbeddingServiceAddr,
	//		"err", err,
	//	)
	//	os.Exit(1)
	//}

	//logger.Global().Info("starting embedding synchronization")
	//if err := SyncAllEmbeddings(context.Background(), dbPool, embeddingClient, logger.Global()); err != nil {
	//	logger.Global().Error("embedding synchronization failed", "err", err)
	//	os.Exit(1)
	//}
	logger.Global().Info("embedding synchronization completed")

	openMux := http.NewServeMux()
	protectedMux := http.NewServeMux()

	shttp.NewStoreRouter(openMux, dbPool, nil, apiV0Prefix)
	shttp.NewItemRouter(openMux, dbPool, apiV0Prefix)
	shttp.NewCartRouter(protectedMux, dbPool, apiV0Prefix)

	protectedHandler := middlewares.AuthMiddleware(protectedMux, conf.JWTSecret)

	mux := http.NewServeMux()

	mux.Handle("/images/items/", http.StripPrefix("/images/items/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fullPath := filepath.Join(conf.UploadItemDir, r.URL.Path)
		logger.Global().DebugContext(r.Context(), "serving item image",
			"url_path", r.URL.Path,
			"fs_path", fullPath,
		)
		http.ServeFile(w, r, fullPath)
	})))

	mux.Handle("/images/stores/", http.StripPrefix("/images/stores/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fullPath := filepath.Join(conf.UploadStoreDir, r.URL.Path)
		logger.Global().DebugContext(r.Context(), "serving store image",
			"url_path", r.URL.Path,
			"fs_path", fullPath,
		)
		http.ServeFile(w, r, fullPath)
	})))

	mux.Handle(apiV0Prefix+"cart", protectedHandler)
	mux.Handle(apiV0Prefix, openMux)

	handler := middlewares.AccessLog(
		logger.Global(),
		middlewares.CorsMiddleware(mux),
	)

	addr := fmt.Sprintf("0.0.0.0:%s", conf.AppPort)
	logger.Global().Info("store service running", "addr", addr)
	log.Fatal(http.ListenAndServe(addr, handler))
}
