package cmd

import (
	"apple_backend/auth_service/internal/config"
	authhttp "apple_backend/auth_service/internal/delivery/http"
	"apple_backend/auth_service/internal/repository"
	"apple_backend/auth_service/internal/usecase"
	"apple_backend/pkg/blacklist"
	"apple_backend/pkg/logger"
	authmw "apple_backend/pkg/middlewares"
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func csrfHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func Run() {
	conf := config.LoadConfig()

	dbPool, err := pgxpool.New(context.Background(), conf.DBPath())
	if err != nil {
		log.Fatal(err)
	}
	defer dbPool.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: conf.RedisURL,
	})
	defer redisClient.Close()

	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.Fatal("Redis connection failed:", err)
	}

	authRepo := repository.NewAuthRepoPostgres(dbPool)
	tokenBlacklist := blacklist.NewRedisTokenBlacklist(redisClient)

	uc := usecase.NewAuthUseCase(authRepo, tokenBlacklist, conf.SecretKeyStr(), logger.Global())

	authMux := http.NewServeMux()
	authMux.Handle("/csrf", http.HandlerFunc(csrfHandler))
	authhttp.NewAuthRouter(authMux, "/auth", uc)

	authHandler := authmw.CSRFTokenMiddleware(
		authmw.CSRFMiddleware(authMux),
	)

	mainMux := http.NewServeMux()
	mainMux.Handle("/api/v0/", http.StripPrefix("/api/v0", authHandler))

	handler := authmw.CorsMiddleware(
		authmw.AccessLog(logger.Global(), mainMux),
	)

	addr := fmt.Sprintf("0.0.0.0:%s", conf.AppPortStr())
	log.Println("Auth service running on", addr)
	log.Fatal(http.ListenAndServe(addr, handler))
}
