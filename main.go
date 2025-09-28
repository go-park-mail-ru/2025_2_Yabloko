package main

import (
	"apple_backend/handlers"
	"apple_backend/logger"
	"apple_backend/middlewares"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dbPath := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"),
		"db", os.Getenv("POSTGRES_PORT"), os.Getenv("DB_NAME"))
	port := fmt.Sprintf(":%s", os.Getenv("APP_PORT"))

	dbPool, err := pgxpool.New(context.Background(), dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer dbPool.Close()

	mux := http.NewServeMux()

	// store
	storeLog := "./log/store.log"
	storeHandler := handlers.New(dbPool, "STORE", storeLog, logger.DEBUG)
	storeAPI := "/api/v0/stores"
	mux.HandleFunc(storeAPI, middlewares.AccessLog(storeHandler.GetStores))

	// login
	authLog := "./log/auth.log"
	authHandler := handlers.New(dbPool, "AUTH", authLog, logger.DEBUG)
	authAPI := "/api/v0/auth"
	mux.HandleFunc(authAPI+"/signup", middlewares.AccessLog(authHandler.Signup))
	mux.HandleFunc(authAPI+"/login", middlewares.AccessLog(authHandler.Login))
	mux.HandleFunc(authAPI+"/logout", middlewares.AccessLog(authHandler.Logout))
	// refresh должен быть публичным или с отдельной проверкой
	mux.HandleFunc(authAPI+"/refresh", middlewares.AccessLog(authHandler.Signup))

	// health
	mux.HandleFunc("/health", middlewares.AccessLog(authHandler.HealthCheck))
	// images
	mux.HandleFunc("/api/v0/image/", middlewares.AccessLog(authHandler.GetImage))

	cors := middlewares.CorsMiddleware(mux)
	fmt.Println("starting server at " + port)
	log.Fatal(http.ListenAndServe(port, cors))
}
