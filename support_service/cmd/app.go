package cmd

import (
	"apple_backend/support_service/internal/config"
	shttp "apple_backend/support_service/internal/delivery/http"
	"apple_backend/support_service/internal/delivery/middlewares"
	ws "apple_backend/support_service/internal/delivery/ws"
	"apple_backend/support_service/internal/repository"
	"context"
	"fmt"
	"log"
	"net/http"

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

	shttp.NewSupportRouter(openMux, dbPool, apiV0Prefix)
	ticketRepo := repository.NewTicketRepoPostgres(dbPool)
	ws.NewRealtimeRouter(openMux, ticketRepo)

	mux := http.NewServeMux()
	mux.Handle(apiV0Prefix, openMux)

	handler := middlewares.CORS(mux)
	addr := fmt.Sprintf("0.0.0.0:%s", conf.AppPort)
	log.Printf("Support service running on http://localhost:%s", conf.AppPort)
	log.Fatal(http.ListenAndServe(addr, handler))
}
