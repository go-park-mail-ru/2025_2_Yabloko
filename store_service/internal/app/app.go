package app

import (
	"apple_backend/store_service/internal/config"
	shttp "apple_backend/store_service/internal/delivery/http"
	"context"
	"fmt"
	"log"
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

	storeHandler := shttp.NewStoreRouter(dbPool, "/api/v0")

	log.Println(fmt.Sprintf("Store service running on %s", conf.AppPort))
	log.Fatal(http.ListenAndServe(fmt.Sprintf("0.0.0.0:%s", conf.AppPort), storeHandler))
}
