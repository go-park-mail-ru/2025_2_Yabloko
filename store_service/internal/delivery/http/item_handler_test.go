package http

import (
	"apple_backend/handlers"
	"apple_backend/pkg/logger"
	"apple_backend/store_service/internal/delivery/mock"
	"apple_backend/store_service/internal/delivery/transport"
	"apple_backend/store_service/internal/domain"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

// parseJSON ТОЛЬКО ДЛЯ ТЕСТОВ
func parseJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func TestItemHandler_GetItemTypes(t *testing.T) {
	url := "/stores/%s/item-types"
	type testCase struct {
		name              string
		method            string
		id                string
		mockSetup         func(uc *mock.MockItemUsecaseInterface)
		expectedCode      int
		expectedResult    []*transport.ItemType
		expectedErrResult *handlers.ErrResponse
	}

	uid1 := "00000000-0000-0000-0000-000000000001"
	uid2 := "00000000-0000-0000-0000-000000000002"

	tests := []testCase{
		{
			name:   "GetItemTypes успешный вызов",
			method: http.MethodGet,
			id:     uid1,
			mockSetup: func(uc *mock.MockItemUsecaseInterface) {
				uc.EXPECT().
					GetItemTypes(context.Background(), uid1).
					Return([]*domain.ItemType{
						{ID: uid1, Name: "Type1"},
						{ID: uid2, Name: "Type2"},
					}, nil)
			},
			expectedCode: http.StatusOK,
			expectedResult: []*transport.ItemType{
				{ID: uid1, Name: "Type1"},
				{ID: uid2, Name: "Type2"},
			},
		},
		{
			name:              "GetItemTypes метод не разрешен",
			method:            http.MethodPost,
			id:                uid1,
			mockSetup:         func(uc *mock.MockItemUsecaseInterface) {},
			expectedCode:      http.StatusMethodNotAllowed,
			expectedErrResult: &handlers.ErrResponse{Err: domain.ErrHTTPMethod.Error()},
		},
		{
			name:              "GetItemTypes неверный формат id",
			method:            http.MethodGet,
			id:                "00000000-1",
			mockSetup:         func(uc *mock.MockItemUsecaseInterface) {},
			expectedCode:      http.StatusBadRequest,
			expectedErrResult: &handlers.ErrResponse{Err: domain.ErrRequestParams.Error()},
		},
		{
			name:   "GetItemTypes не найдено данных",
			method: http.MethodGet,
			id:     uid1,
			mockSetup: func(uc *mock.MockItemUsecaseInterface) {
				uc.EXPECT().
					GetItemTypes(context.Background(), uid1).
					Return(nil, domain.ErrRowsNotFound)
			},
			expectedCode:      http.StatusNotFound,
			expectedErrResult: &handlers.ErrResponse{Err: domain.ErrRowsNotFound.Error()},
		},
		{
			name:   "GetItemTypes внутренняя ошибка",
			method: http.MethodGet,
			id:     uid1,
			mockSetup: func(uc *mock.MockItemUsecaseInterface) {
				uc.EXPECT().
					GetItemTypes(context.Background(), uid1).
					Return(nil, domain.ErrInternalServer)
			},
			expectedCode:      http.StatusInternalServerError,
			expectedErrResult: &handlers.ErrResponse{Err: domain.ErrInternalServer.Error()},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uc := mock.NewMockItemUsecaseInterface(ctrl)
	handler := NewItemHandler(uc, logger.NewNilLogger())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup(uc)

			req := httptest.NewRequest(tt.method, fmt.Sprintf(url, tt.id), bytes.NewBuffer(nil))
			req = req.WithContext(context.Background())
			req.SetPathValue("id", tt.id)

			w := httptest.NewRecorder()

			handler.GetItemTypes(w, req)

			require.Equal(t, tt.expectedCode, w.Code)
			if tt.expectedResult != nil {
				require.JSONEq(t, w.Body.String(), parseJSON(tt.expectedResult))
			}
			if tt.expectedErrResult != nil {
				require.JSONEq(t, w.Body.String(), parseJSON(tt.expectedErrResult))
			}
		})
	}
}

func TestItemHandler_GetItems(t *testing.T) {
	url := "/stores/%s/items"
	type testCase struct {
		name              string
		method            string
		id                string
		mockSetup         func(uc *mock.MockItemUsecaseInterface)
		expectedCode      int
		expectedResult    []*transport.Item
		expectedErrResult *handlers.ErrResponse
	}

	uid1 := "00000000-0000-0000-0000-000000000001"
	uid2 := "00000000-0000-0000-0000-000000000002"

	item1 := &domain.ItemAgg{
		ID:          uid1,
		Name:        "name1",
		Price:       1,
		Description: "description1",
		CardImg:     "card_img1",
		TypesID:     []string{"type1"},
	}
	item2 := &domain.ItemAgg{
		ID:          uid2,
		Name:        "name2",
		Price:       2,
		Description: "description2",
		CardImg:     "card_img2",
		TypesID:     []string{"type2"},
	}

	itemResp1 := &transport.Item{
		ID:          uid1,
		Name:        "name1",
		Price:       1,
		Description: "description1",
		CardImg:     "card_img1",
		TypesID:     []string{"type1"},
	}
	itemResp2 := &transport.Item{
		ID:          uid2,
		Name:        "name2",
		Price:       2,
		Description: "description2",
		CardImg:     "card_img2",
		TypesID:     []string{"type2"},
	}

	tests := []testCase{
		{
			name:   "GetItems успешный вызов",
			method: http.MethodGet,
			id:     uid1,
			mockSetup: func(uc *mock.MockItemUsecaseInterface) {
				uc.EXPECT().
					GetItems(context.Background(), uid1).
					Return([]*domain.ItemAgg{item1, item2}, nil)
			},
			expectedCode:   http.StatusOK,
			expectedResult: []*transport.Item{itemResp1, itemResp2},
		},
		{
			name:              "GetItems метод не разрешен",
			method:            http.MethodPost,
			id:                uid1,
			mockSetup:         func(uc *mock.MockItemUsecaseInterface) {},
			expectedCode:      http.StatusMethodNotAllowed,
			expectedErrResult: &handlers.ErrResponse{Err: domain.ErrHTTPMethod.Error()},
		},
		{
			name:              "GetItems неверный формат id",
			method:            http.MethodGet,
			id:                "00000000-1",
			mockSetup:         func(uc *mock.MockItemUsecaseInterface) {},
			expectedCode:      http.StatusBadRequest,
			expectedErrResult: &handlers.ErrResponse{Err: domain.ErrRequestParams.Error()},
		},
		{
			name:   "GetItems не найдено данных",
			method: http.MethodGet,
			id:     uid1,
			mockSetup: func(uc *mock.MockItemUsecaseInterface) {
				uc.EXPECT().
					GetItems(context.Background(), uid1).
					Return(nil, domain.ErrRowsNotFound)
			},
			expectedCode:      http.StatusNotFound,
			expectedErrResult: &handlers.ErrResponse{Err: domain.ErrRowsNotFound.Error()},
		},
		{
			name:   "GetItems внутренняя ошибка",
			method: http.MethodGet,
			id:     uid1,
			mockSetup: func(uc *mock.MockItemUsecaseInterface) {
				uc.EXPECT().
					GetItems(context.Background(), uid1).
					Return(nil, domain.ErrInternalServer)
			},
			expectedCode:      http.StatusInternalServerError,
			expectedErrResult: &handlers.ErrResponse{Err: domain.ErrInternalServer.Error()},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uc := mock.NewMockItemUsecaseInterface(ctrl)
	handler := NewItemHandler(uc, logger.NewNilLogger())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup(uc)

			req := httptest.NewRequest(tt.method, fmt.Sprintf(url, tt.id), bytes.NewBuffer(nil))
			req = req.WithContext(context.Background())
			req.SetPathValue("id", tt.id)

			w := httptest.NewRecorder()

			handler.GetItems(w, req)

			require.Equal(t, tt.expectedCode, w.Code)
			if tt.expectedResult != nil {
				require.JSONEq(t, w.Body.String(), parseJSON(tt.expectedResult))
			}
			if tt.expectedErrResult != nil {
				require.JSONEq(t, w.Body.String(), parseJSON(tt.expectedErrResult))
			}
		})
	}
}
