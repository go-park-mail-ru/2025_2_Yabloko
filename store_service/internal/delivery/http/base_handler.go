package http

import (
	"apple_backend/pkg/http_response"
	"apple_backend/pkg/logger"
	"apple_backend/store_service/internal/domain"
	"apple_backend/store_service/internal/repository"
	"net/http"
	"os"
	"path/filepath"
)

type BaseHandler struct {
	imageDir string
	rs       *http_response.ResponseSender
	db       repository.PgxIface
}

func NewBaseHandler(log *logger.Logger, dbPool repository.PgxIface, imageDir string) *BaseHandler {
	return &BaseHandler{
		imageDir: imageDir,
		rs:       http_response.NewResponseSender(log),
		db:       dbPool,
	}
}

func NewBaseRouter(mux *http.ServeMux, appLog *logger.Logger, dbPool repository.PgxIface, apiPrefix, imageDir string) {
	baseHandler := NewBaseHandler(appLog, dbPool, imageDir)

	mux.HandleFunc(apiPrefix+"images/{path}", baseHandler.GetImage)
	mux.HandleFunc(apiPrefix+"health", baseHandler.HealthCheck)
}

// GetImage godoc
// @Summary      Получить изображение по пути
// @Description  Возвращает изображение по указанному пути в файловой системе
// @Tags         images
// @Accept       */*
// @Produce      image/png
// @Param        img_path  path  string  true  "Путь к изображению (например, 'products/item1.png')"
// @Success      200  {file}  binary  "Изображение успешно найдено"
// @Failure      400  {object}  http_response.ErrResponse  "Некорректный путь или параметры"
// @Failure      404  {object}  http_response.ErrResponse  "Изображение не найдено"
// @Failure      405  {object}  http_response.ErrResponse  "Неверный HTTP-метод"
// @Failure      500  {object}  http_response.ErrResponse  "Внутренняя ошибка сервера"
// @Router       /images/{img_path} [get]
func (h *BaseHandler) GetImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.rs.Error(r.Context(), w, http.StatusMethodNotAllowed, "GetImage", domain.ErrHTTPMethod, nil)
		return
	}

	imgPath := r.PathValue("path")
	imgPath = filepath.Clean(imgPath)

	if imgPath == "" {
		h.rs.Error(r.Context(), w, http.StatusBadRequest, "GetImage", domain.ErrRequestParams, nil)
		return
	}

	fullPath := filepath.Join(h.imageDir, imgPath)
	info, err := os.Stat(fullPath)
	if err != nil || info.IsDir() {
		h.rs.Error(r.Context(), w, http.StatusBadRequest, "GetImage", domain.ErrRequestParams, err)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	http.ServeFile(w, r, fullPath)
}

// HealthCheck godoc
// @Summary      Проверка сервера
// @Description  Эндпоинт для проверки доступности приложения и базы данных
// @Tags         health
// @Accept       */*
// @Produce      json
// @Success      200  {object}  map[string]string  "Сервер работает исправно"
// @Failure      405  {object}  http_response.ErrResponse  "Неверный HTTP-метод"
// @Failure      500  {object}  http_response.ErrResponse  "Внутренняя ошибка сервера"
// @Router       /health [get]
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

	h.rs.Send(r.Context(), w, http.StatusOK, nil)
}
