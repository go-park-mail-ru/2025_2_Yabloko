package cmd

import (
	"apple_backend/pkg/logger"
	"apple_backend/profile_service/internal/config"
	phttp "apple_backend/profile_service/internal/delivery/http"
	"apple_backend/profile_service/internal/delivery/middlewares"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Run(appLog, accessLog logger.Logger) {
	conf := config.LoadConfig()

	dbPool, err := pgxpool.New(context.Background(), conf.DBPath())
	if err != nil {
		log.Fatal(err)
	}
	defer dbPool.Close()

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		name := filepath.Base(r.URL.Path)
		if name == "." || name == ".." || strings.Contains(name, "/") {
			http.NotFound(w, r)
			return
		}

		if !strings.Contains(name, "_") {
			http.NotFound(w, r)
			return
		}

		ext := strings.ToLower(filepath.Ext(name))
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
			http.NotFound(w, r)
			return
		}

		fullPath := filepath.Join(conf.UploadPath, name)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Cache-Control", "public, max-age=86400")
		http.ServeFile(w, r, fullPath)
	})

	protectedMux := http.NewServeMux()
	phttp.NewProfileRouter(protectedMux, dbPool, "/api/v0", appLog, conf.UploadPath, conf.BaseURL)

	jwtSecret := conf.JWTSecret
	protectedHandler := middlewares.AuthMiddleware(protectedMux, jwtSecret)
	mux.Handle("/api/v0/", protectedHandler)

	handler := middlewares.AccessLog(accessLog,
		middlewares.CorsMiddleware(
			middlewares.CSRFTokenMiddleware(
				middlewares.CSRFMiddleware(mux),
			),
		),
	)

	addr := fmt.Sprintf("0.0.0.0:%s", conf.AppPort)
	log.Printf("✅ Profile service running on http://%s:%s", conf.AppHost, conf.AppPort)
	log.Printf("✅ Avatar URL base: %s", conf.BaseURL)
	log.Fatal(http.ListenAndServe(addr, handler))
}
