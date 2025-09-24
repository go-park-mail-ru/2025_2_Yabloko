package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

const (
	PORT = "8080"
)

func logger(fun http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
		fun(w, r)
		log.Printf("%s %s %s duration %s", r.RemoteAddr, r.Method, r.URL, time.Since(start))
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	allowedOrigins := map[string]bool{
		"http://localhost:3000": true,
		"http://127.0.0.1:3000": true,
		// production domains...
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept, Origin, X-Requested-With, X-CSRF-Token")
		} else {
			// origin отсутствует (обычный запрос) или не разрешён - не выставляем заголовок
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	port := fmt.Sprintf(":%s", PORT)
	mux := http.NewServeMux()

	// public routes
	mux.HandleFunc("/login", logger(loginHandler))
	mux.HandleFunc("/register", logger(registerHandler))
	mux.HandleFunc("/logout", logger(logoutHandler))
	// refresh должен быть публичным или с отдельной проверкой
	mux.HandleFunc("/refresh", logger(refreshTokenHandler))

	// protected routes - создаем цепочку middleware

	handler := corsMiddleware(mux)
	fmt.Println("starting server at " + port)
	log.Fatal(http.ListenAndServe(port, handler))
}
