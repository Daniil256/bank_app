package auth

import "slices"

import "net/http"

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// 1. Разрешаем только доверенные origin
		if isAllowedOrigin(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin") // важно для кэширования
		}

		// 2. Разрешённые методы
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")

		// 3. Разрешённые заголовки
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-ID")

		// 4. Разрешаем credentials (cookies, Authorization)
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// 5. Кэшируем preflight на 1 час
		w.Header().Set("Access-Control-Max-Age", "3600")

		// 6. Preflight — отвечаем 204 и не идём дальше
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func isAllowedOrigin(origin string) bool {
	allowed := []string{
		"https://app.bank.com",
		"https://admin.bank.com",
		"http://localhost:3000", // для разработки
	}
	return slices.Contains(allowed, origin)
}
