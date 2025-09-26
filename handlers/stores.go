package handlers

import (
	"apple_backend/custom_errors"
	"apple_backend/db/store"
	"apple_backend/logger"
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	dbPool *pgxpool.Pool
	log    logger.Logger
}

func NewStoreHandler(dbPool *pgxpool.Pool, logPath string, logLevel logger.LogLevel) *Handler {
	return &Handler{
		dbPool: dbPool,
		log:    *logger.NewLogger("STORE HANDLER", logPath, logLevel),
	}
}

type errResponse struct {
	Err error `json:"error"`
}

func (h *Handler) GetStores(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req store.GetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errResponse{Err: custom_errors.InvalidJSONErr})
		return
	}
	//todo validate params
	if req.Limit <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errResponse{Err: custom_errors.InvalidJSONErr})
		h.log.Warn(logger.LogInfo{Info: "GetStores invalid json", Meta: req})
		return
	}

	stores, err := store.GetStores(h.dbPool, req)
	if err != nil {
		if len(stores) > 0 {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(stores)
			h.log.Warn(logger.LogInfo{Info: "GetStores error with answer", Err: err, Meta: req})
			return
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(errResponse{Err: custom_errors.InnerErr})
			h.log.Warn(logger.LogInfo{Info: "GetStores invalid json", Err: err, Meta: req})
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(stores)
}
