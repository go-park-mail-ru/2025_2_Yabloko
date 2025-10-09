package cmd

import (
	"apple_backend/pkg/logger"
	"apple_backend/store_service/internal/config"
	shttp "apple_backend/store_service/internal/delivery/http"
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Run() {
	conf := config.LoadConfig()
	dbPool, err := pgxpool.New(context.Background(), conf.DBPath())
	if err != nil {
		log.Fatal(err)
	}
	defer dbPool.Close()

	accessLog := logger.NewLogger("./logs/access.log", slog.LevelDebug)
	appLog := logger.NewLogger("./logs/app.log", slog.LevelDebug)

	storeHandler := shttp.NewStoreRouter(dbPool, "/api/v0", appLog, accessLog)

	log.Println(fmt.Sprintf("Store service running on %s", conf.AppPort))
	log.Fatal(http.ListenAndServe(fmt.Sprintf("0.0.0.0:%s", conf.AppPort), storeHandler))
}
