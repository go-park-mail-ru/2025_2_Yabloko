package http

import (
	"apple_backend/pkg/http_response"
	"apple_backend/pkg/logger"
	"apple_backend/store_service/internal/domain"
	"apple_backend/store_service/internal/repository"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type BaseHandler struct {
	imageDir string
	rs       *http_response.ResponseSender
	db       repository.PgxIface
}

func NewBaseHandler(log logger.Logger, dbPool repository.PgxIface, imageDir string) *BaseHandler {
	return &BaseHandler{
		imageDir: imageDir,
		rs:       http_response.NewResponseSender(log),
		db:       dbPool,
	}
}

func NewBaseRouter(mux *http.ServeMux, log logger.Logger, db repository.PgxIface, apiPrefix, uploadStoreDir string) {
	// Health-check
	mux.HandleFunc(apiPrefix+"health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"store"}`))
	})

	// /images/stores/uuid.jpg → /app/stores/uuid.jpg
	mux.Handle("/images/stores/", http.StripPrefix("/images/stores/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fullPath := filepath.Join(uploadStoreDir, r.URL.Path)
		log.Debug("Serving image", map[string]interface{}{
			"path": r.URL.Path,
			"file": fullPath,
		})
		http.ServeFile(w, r, fullPath)
	})))

	log.Info("BaseRouter: serving store images", map[string]interface{}{
		"dir": uploadStoreDir,
	})
}

func (h *BaseHandler) GetImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.rs.Error(r.Context(), w, http.StatusMethodNotAllowed, "GetImage", domain.ErrHTTPMethod, nil)
		return
	}

	imgPath := r.PathValue("path")
	imgPath = filepath.Clean(imgPath)

	if imgPath == "" || strings.Contains(imgPath, "..") {
		h.rs.Error(r.Context(), w, http.StatusBadRequest, "GetImage", domain.ErrRequestParams, nil)
		return
	}

	fullPath := filepath.Join(h.imageDir, imgPath)
	info, err := os.Stat(fullPath)
	if os.IsNotExist(err) || info.IsDir() {
		h.rs.Error(r.Context(), w, http.StatusNotFound, "GetImage", domain.ErrRowsNotFound, err)
		return
	}
	if err != nil {
		h.rs.Error(r.Context(), w, http.StatusInternalServerError, "GetImage", domain.ErrInternalServer, err)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	http.ServeFile(w, r, fullPath)
}

func (h *BaseHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.rs.Error(r.Context(), w, http.StatusMethodNotAllowed, "HealthCheck", domain.ErrHTTPMethod, nil)
		return
	}
	err := h.db.Ping(r.Context())
	if err != nil {
		h.rs.Error(r.Context(), w, http.StatusInternalServerError, "HealthCheck", domain.ErrInternalServer, err)
		return
	}

	h.rs.Send(r.Context(), w, http.StatusOK, map[string]string{"status": "ok"})
}
