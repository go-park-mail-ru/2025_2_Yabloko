package main

import (
	"apple_backend/handlers"
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	conf := LoadConfig()

	dbPool, err := pgxpool.New(context.Background(), conf.DBPath())
	if err != nil {
		log.Fatal(err)
	}
	defer dbPool.Close()

	router := handlers.NewMainRouter(dbPool)

	fmt.Println("starting server at " + conf.AppPort)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+conf.AppPort, router))
}
