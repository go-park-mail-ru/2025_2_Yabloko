package handlers

import (
	"apple_backend/custom_errors"
	"apple_backend/db"
	"apple_backend/logger"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"
)

func (h *Handler) GetStores(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.handleError(w, http.StatusMethodNotAllowed, custom_errors.HTTPMethodErr, nil)
		return
	}

	var req db.GetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.handleError(w, http.StatusBadRequest, custom_errors.InvalidJSONErr, err)
		return
	}
	//todo validate params
	if req.Limit <= 0 {
		h.handleError(w, http.StatusBadRequest, custom_errors.InvalidJSONErr, nil)
		return
	}

	stores, err := db.GetStores(h.dbPool, req)
	if err != nil {
		if len(stores) > 0 {
			h.log.Warn(logger.LogInfo{Info: "GetStores ответ с ошибкой", Err: err, Meta: req})
		} else if errors.Is(err, pgx.ErrNoRows) {
			h.log.Info(logger.LogInfo{Info: "GetStores пустой ответ", Err: err, Meta: req})

		} else {
			h.handleError(w, http.StatusInternalServerError, custom_errors.InnerErr, err)
			return
		}
	}

	h.handleResponse(w, http.StatusOK, stores)
}
