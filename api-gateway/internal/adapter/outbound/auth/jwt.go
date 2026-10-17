package auth

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/metadata"
)

type JWTValidator struct {
	publicKey *rsa.PublicKey
}

func NewJWTValidator(publicKeyPath string) (*JWTValidator, error) {
	data, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("read public key: %w", err)
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("decode PEM: invalid block")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse public key: %w", err)
	}
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not RSA public key")
	}
	return &JWTValidator{publicKey: rsaPub}, nil
}

// Middleware валидирует JWT и прокидывает claims в gRPC metadata
func (v *JWTValidator) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Пропускаем публичные эндпоинты
		if isPublicPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			writeJSONError(w, http.StatusUnauthorized, "missing or invalid authorization header")
			return
		}

		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return v.publicKey, nil
		})
		if err != nil || !token.Valid {
			writeJSONError(w, http.StatusUnauthorized, "invalid token")
			return
		}

		// Пробрасываем JWT и claims в gRPC metadata
		ctx := metadata.AppendToOutgoingContext(r.Context(),
			"authorization", "Bearer "+tokenStr,
			"x-user-id", getString(claims, "sub"),
			"x-user-role", getString(claims, "role"),
			"x-customer-id", getString(claims, "customer_id"),
		)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func isPublicPath(path string) bool {
	public := []string{
		"/v1/auth/register",
		"/v1/auth/login",
		"/v1/auth/refresh",
		"/health",
		"/ready",
	}
	for _, p := range public {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}

func getString(m jwt.MapClaims, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func writeJSONError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	fmt.Fprintf(w, `{"code":%d,"message":%q}`, code, msg)
}
