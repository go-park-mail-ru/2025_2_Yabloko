package handlers

import (
	"apple_backend/auth"
	"apple_backend/custom_errors"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
)

func requireJWTCookie(t *testing.T, cookies []*http.Cookie) *http.Cookie {
	t.Helper()
	for _, c := range cookies {
		if c.Name == "jwt_token" {
			require.NotEmpty(t, c.Value, "jwt_token пустой")
			require.True(t, c.Expires.After(time.Now()), "cookie истекло")
			require.True(t, c.HttpOnly)
			require.Equal(t, "/", c.Path)
			return c
		}
	}
	t.Fatalf("jwt_token нет в куки")
	return nil
}

func requireNoJWTCookie(t *testing.T, cookies []*http.Cookie) {
	t.Helper()
	for _, c := range cookies {
		require.NotEqual(t, "jwt_token", c.Name, "jwt_token не должно быть")
	}
}

func TestLogin(t *testing.T) {
	uuid1 := uuid.MustParse("00000000-0000-0000-0000-000000000001").String()
	hash1, _ := auth.HashPassword("fwy!&Daj812")

	cases := []struct {
		name        string
		requestBody auth.LoginRequest
		setupPool   func(pool pgxmock.PgxPoolIface)
		wantStatus  int
		wantCookie  bool
		wantBody    interface{}
	}{
		{
			name:        "Success",
			requestBody: auth.LoginRequest{Email: "user1@vk.com", Password: "fwy!&Daj812"},
			setupPool: func(pool pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "hash"}).AddRow(uuid1, hash1)
				pool.ExpectQuery(regexp.QuoteMeta(`
					select id, hash
					from account
					where email = $1
				`)).
					WithArgs("user1@vk.com").
					WillReturnRows(rows)
			},
			wantStatus: http.StatusOK,
			wantCookie: true,
			wantBody: AuthResponse{
				Message: "OK",
				Email:   "user1@vk.com",
				UserID:  uuid1,
			},
		},
		{
			name:        "Unregistered user",
			requestBody: auth.LoginRequest{Email: "user1@vk.com", Password: "fwy!&Daj812"},
			setupPool: func(pool pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "hash"})
				pool.ExpectQuery(regexp.QuoteMeta(`
					select id, hash
					from account
					where email = $1
				`)).
					WithArgs("user1@vk.com").
					WillReturnRows(rows)
			},
			wantStatus: http.StatusUnauthorized,
			wantCookie: false,
			wantBody:   ErrResponse{Err: custom_errors.IncorrectLoginOrPassword.Error()},
		},
		{
			name:        "Wrong password",
			requestBody: auth.LoginRequest{Email: "user1@vk.com", Password: "wrong-password"},
			setupPool: func(pool pgxmock.PgxPoolIface) {
				// возвращаем пользователя, но хеш пароля не совпадёт с введённым
				rows := pgxmock.NewRows([]string{"id", "hash"}).AddRow(uuid1, hash1)
				pool.ExpectQuery(regexp.QuoteMeta(`
					select id, hash
					from account
					where email = $1
				`)).
					WithArgs("user1@vk.com").
					WillReturnRows(rows)
			},
			wantStatus: http.StatusUnauthorized,
			wantCookie: false,
			wantBody:   ErrResponse{Err: custom_errors.IncorrectLoginOrPassword.Error()},
		},
		{
			name:        "DB error",
			requestBody: auth.LoginRequest{Email: "user1@vk.com", Password: "fwy!&Daj812"},
			setupPool: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(regexp.QuoteMeta(`
					select id, hash
					from account
					where email = $1
				`)).
					WithArgs("user1@vk.com").
					WillReturnError(pgx.ErrTxClosed)
			},
			wantStatus: http.StatusInternalServerError,
			wantCookie: false,
			wantBody:   nil,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {

			pool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer pool.Close()

			tt.setupPool(pool)

			h := NewHandler(pool, "", "", 0)
			router := NewAuthRouter(h)

			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(parseJSON(tt.requestBody)))
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			require.Equal(t, tt.wantStatus, w.Code)

			if tt.wantCookie {
				reqCookie := requireJWTCookie(t, w.Result().Cookies())
				require.NotEmpty(t, reqCookie.Value)
			} else {
				requireNoJWTCookie(t, w.Result().Cookies())
			}

			if tt.wantBody != nil {
				expectedJSON, err := json.Marshal(tt.wantBody)
				require.NoError(t, err)
				require.JSONEq(t, string(expectedJSON), w.Body.String())
			}
		})
	}
}

func TestSignup(t *testing.T) {

	validReq := auth.RegisterRequest{
		Email:    "newuser@vk.com",
		Password: "fwy!&Daj812",
	}

	cases := []struct {
		name       string
		reqBody    interface{}
		setupPool  func(pool pgxmock.PgxPoolIface)
		wantStatus int
		wantCookie bool
		wantBody   interface{}
	}{
		// fixme
		//{
		//	name:    "Success",
		//	reqBody: validReq,
		//	setupPool: func(pool pgxmock.PgxPoolIface) {
		//		pool.ExpectExec(regexp.QuoteMeta(`
		//			insert into account (id, email, hash)
		//			values ($1, $2, $3)
		//		`)).
		//			WithArgs(pgxmock.AnyArg(), validReq.Email, pgxmock.AnyArg()).
		//			WillReturnResult(pgxmock.NewResult("", 1))
		//	},
		//	wantStatus: http.StatusOK,
		//	wantCookie: true,
		//	wantBody: AuthResponse{
		//		Message: "OK",
		//		Email:   validReq.Email,
		//	},
		//},
		{
			name:       "Invalid JSON",
			reqBody:    "not-a-json",
			setupPool:  func(pool pgxmock.PgxPoolIface) {},
			wantStatus: http.StatusBadRequest,
			wantCookie: false,
			wantBody:   nil,
		},
		{
			name:       "Validation error",
			reqBody:    auth.RegisterRequest{Email: "bad-email", Password: "short"},
			setupPool:  func(pool pgxmock.PgxPoolIface) {},
			wantStatus: http.StatusBadRequest,
			wantCookie: false,
			wantBody:   nil,
		},
		{
			name:    "Email already exists",
			reqBody: validReq,
			setupPool: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectExec(regexp.QuoteMeta(`
					insert into account (id, email, hash)
					values ($1, $2, $3)
				`)).
					WithArgs(pgxmock.AnyArg(), validReq.Email, pgxmock.AnyArg()).
					WillReturnError(errors.New("(SQLSTATE 23505)"))
			},
			wantStatus: http.StatusConflict,
			wantCookie: false,
			wantBody:   ErrResponse{Err: custom_errors.EmailAlreadyExist.Error()},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {

			pool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer pool.Close()

			tt.setupPool(pool)

			h := NewHandler(pool, "", "", 0)
			router := NewAuthRouter(h)

			var body []byte
			switch v := tt.reqBody.(type) {
			case string:
				body = []byte(v)
			default:
				body = parseJSON(v)
			}

			req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewBuffer(body))
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			require.Equal(t, tt.wantStatus, w.Code)

			if tt.wantCookie {
				requireJWTCookie(t, w.Result().Cookies())
			} else {
				requireNoJWTCookie(t, w.Result().Cookies())
			}

			if tt.wantBody != nil {
				expectedJSON, _ := json.Marshal(tt.wantBody)
				require.JSONEq(t, string(expectedJSON), w.Body.String())
			}
		})
	}
}
