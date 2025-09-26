package main

import (
	"apple_backend/handlers"
	"apple_backend/logger"
	"apple_backend/middlewares"
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	PORT = "8080"

	POSTGRES_USER     = "postgres"
	POSTGRES_PASSWORD = "admin"
	POSTGRES_HOST     = "127.0.0.1"
	POSTGRES_PORT     = 5432
	DB_NAME           = "postgres"
)

func main() {
	dbPath := fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		POSTGRES_USER, POSTGRES_PASSWORD, POSTGRES_HOST, POSTGRES_PORT, DB_NAME)
	port := fmt.Sprintf(":%s", PORT)

	dbPool, err := pgxpool.New(context.Background(), dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer dbPool.Close()

	mux := http.NewServeMux()

	// store
	storeLog := "./log/store.log"
	storeHandler := handlers.NewStoreHandler(dbPool, storeLog, logger.DEBUG)
	storeAPI := "/api/v0/stores"
	mux.HandleFunc(storeAPI, middlewares.LoggerWrapper(storeHandler.GetStores))

	// login
	authLog := "./log/auth.log"
	authHandler := handlers.NewAuthHandler(dbPool, authLog, logger.DEBUG)
	authAPI := "/api/v0/auth"
	mux.HandleFunc(authAPI+"/signup", middlewares.LoggerWrapper(authHandler.SignupHandler))
	mux.HandleFunc(authAPI+"/login", middlewares.LoggerWrapper(authHandler.LoginHandler))
	mux.HandleFunc(authAPI+"/logout", middlewares.LoggerWrapper(authHandler.LogoutHandler))
	// refresh должен быть публичным или с отдельной проверкой
	mux.HandleFunc(authAPI+"/refresh", middlewares.LoggerWrapper(authHandler.SignupHandler))

	cors := middlewares.CorsMiddleware(mux)
	fmt.Println("starting server at " + port)
	log.Fatal(http.ListenAndServe(port, cors))
}
