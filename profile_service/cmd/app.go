package main

import (
	"apple_backend/pkg/logger"
	"apple_backend/profile_service/internal/config"
	phttp "apple_backend/profile_service/internal/delivery/http"
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	conf := config.LoadConfig()
	dbPool, err := pgxpool.New(context.Background(), conf.DBPath())
	if err != nil {
		log.Fatal(err)
	}
	defer dbPool.Close()

	accessLog := logger.NewLogger("./logs/access.log", slog.LevelDebug)
	appLog := logger.NewLogger("./logs/app.log", slog.LevelDebug)

	profileHandler := phttp.NewProfileRouter(dbPool, "/api/v0", appLog, accessLog)

	log.Println(fmt.Sprintf("Profile service running on %s", conf.AppPort))
	log.Fatal(http.ListenAndServe(fmt.Sprintf("0.0.0.0:%s", conf.AppPort), profileHandler))
}
