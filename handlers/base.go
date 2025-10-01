package handlers

import (
	"apple_backend/custom_errors"
	"apple_backend/db"
	"apple_backend/logger"
	"apple_backend/middlewares"
	"context"
	"encoding/json"
	"net/http"
	"path/filepath"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewMainRouter(dbPool *pgxpool.Pool) http.Handler {
	mux := http.NewServeMux()

	// store
	storeHandler := NewHandler(dbPool, "STORE", "./log/store.log", logger.DEBUG)
	mux.Handle("/api/v0/stores", NewStoresRouter(storeHandler))

	// auth
	authHandler := NewHandler(dbPool, "AUTH", "./log/auth.log", logger.DEBUG)
	mux.Handle("/api/v0/auth/", http.StripPrefix("/api/v0/auth", NewAuthRouter(authHandler)))

	// health
	mux.HandleFunc("/health", middlewares.AccessLog(authHandler.HealthCheck))

	// images
	mux.HandleFunc("/api/v0/image/", middlewares.AccessLog(authHandler.GetImage))

	return middlewares.CorsMiddleware(mux)
}

type Handler struct {
	dbPool db.PoolDB
	log    *logger.Logger
}

func NewHandler(dbPool db.PoolDB, routerName, logPath string, logLevel logger.LogLevel) *Handler {
	var log *logger.Logger
	if routerName != "" && logPath != "" {
		log = logger.NewLogger(routerName, logPath, logLevel)
	} else {
		log = logger.NewNilLogger()
	}

	return &Handler{
		dbPool: dbPool,
		log:    log,
	}
}

// хелпер для отправки ответов
func (h *Handler) handleResponse(w http.ResponseWriter, statusCode int, response interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.log.Error(logger.LogInfo{Err: err, Info: "Ошибка декодирования json ответа", Meta: response})
	}
}

func (h *Handler) handleError(w http.ResponseWriter, statusCode int, userError error, internalErr error) {
	if internalErr != nil {
		h.log.Error(logger.LogInfo{Err: internalErr, Meta: userError})
	} else {
		h.log.Error(logger.LogInfo{Err: userError})
	}

	h.handleResponse(w, statusCode, ErrResponse{Err: userError.Error()})
}

type ErrResponse struct {
	Err string `json:"error"`
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	err := h.dbPool.Ping(context.Background())
	if err != nil {
		h.handleError(w, http.StatusInternalServerError, custom_errors.InnerErr, err)
	}

	h.handleResponse(w, http.StatusOK, nil)
}

func (h *Handler) GetImage(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Path[len("/api/v0/image/"):]
	if filename == "" {
		h.handleError(w, http.StatusBadRequest, custom_errors.InvalidJSONErr, nil)
		return
	}
	// TODO параметризовать путь
	path := filepath.Join("../images", filename)

	w.Header().Set("Content-Type", "image/png")
	http.ServeFile(w, r, path)
}
