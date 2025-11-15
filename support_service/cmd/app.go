package cmd

import (
	"apple_backend/support_service/internal/config"
	"apple_backend/support_service/internal/delivery/http" // ← твои HTTP-хендлеры (пока не реализованы)
	"apple_backend/support_service/internal/delivery/middlewares"
	"apple_backend/support_service/internal/delivery/ws"
	"apple_backend/support_service/internal/repository"
	"apple_backend/support_service/pkg/logger"

	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Run() {
	conf := config.MustConfig()
	logger.Global().Info("Starting support service", "port", conf.AppPort)

	// Подключение к БД
	dbPool, err := pgxpool.New(context.Background(), conf.DBPath())
	if err != nil {
		log.Fatal("failed to connect to database", err)
	}
	defer dbPool.Close()

	// Основной мультиплексор
	mux := http.NewServeMux()

	// ======================
	// 1. WebSocket (реалтайм)
	// ======================
	ticketRepo := repository.NewTicketRepoPostgres(dbPool)
	ws.NewRealtimeRouter(mux, ticketRepo) // ← регистрирует /ws/ticket/{id}

	// ======================
	// 2. Публичные HTTP-роуты (без аутентификации)
	//    — создание тикетов, получение своих тикетов, сообщения, рейтинг
	// ======================
	openMux := http.NewServeMux()
	// TODO: создать http.NewSupportPublicRouter(openMux, dbPool, "/api/v1")
	// Пока оставим заглушку или реализуем ниже

	// ======================
	// 3. Защищённые HTTP-роуты (требуют JWT)
	//    — админка: все тикеты, изменение статуса, статистика
	// ======================
	protectedMux := http.NewServeMux()
	// TODO: http.NewSupportAdminRouter(protectedMux, dbPool, "/api/v1")

	// Middleware для защищённых роутов
	protectedHandler := middlewares.AuthMiddleware(
		protectedMux,
		conf.JWTSecret,
		conf.AdminUserIDs, // ← передаём список админов
	)

	// ======================
	// 4. Монтируем всё в главный mux
	// ======================
	// Публичные эндпоинты
	// mux.Handle("/api/v1/", openMux)        // ← раскомментировать после реализации

	// Защищённые (админка)
	mux.Handle("/api/v1/admin/", protectedHandler)

	// WebSocket (уже добавлен через ws.NewRealtimeRouter)

	// ======================
	// 5. Middleware: CORS → AccessLog
	// ======================
	handler := middlewares.AccessLog(
		logger.Global(),
		middlewares.CorsMiddleware(mux),
	)

	// ======================
	// 6. Запуск сервера
	// ======================
	addr := fmt.Sprintf("0.0.0.0:%s", conf.AppPort)
	log.Printf("Support service running on http://localhost:%s", conf.AppPort)
	log.Fatal(http.ListenAndServe(addr, handler))
}
