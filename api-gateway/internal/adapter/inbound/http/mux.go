package http

import (
	"api-gateway/internal/adapter/outbound/auth"
	customerv1 "api-gateway/proto/bank/v1"
	"context"
	"fmt"
	"google.golang.org/grpc/metadata"
	"net/http"

	authv1 "api-gateway/proto/auth/v1"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// NewMux создаёт grpc-gateway ServeMux и регистрирует все сервисы
func NewMux(ctx context.Context, GRPCAuthEndpoint, GRPCCustomerEndpoint string) (http.Handler, error) {
	mux := runtime.NewServeMux(
		runtime.WithErrorHandler(errorHandler),
		runtime.WithIncomingHeaderMatcher(headerMatcher),
		runtime.WithMetadata(func(ctx context.Context, r *http.Request) metadata.MD {
			md := metadata.MD{}

			// Пробрасываем Authorization
			if authorization := r.Header.Get("Authorization"); authorization != "" {
				md.Set("authorization", authorization)
			}

			// Пробрасываем claims из HTTP-контекста
			if claims, ok := auth.ClaimsFromContext(ctx); ok {
				if sub, ok := claims["sub"].(string); ok {
					md.Set("x-user-id", sub)
				}
				if role, ok := claims["role"].(string); ok {
					md.Set("x-user-role", role)
				}
				if cid, ok := claims["customer_id"].(string); ok && cid != "" {
					md.Set("x-customer-id", cid)
				}
				if jti, ok := claims["jti"].(string); ok {
					md.Set("x-jti", jti)
				}
			}

			return md
		}),
	)

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	if err := authv1.RegisterAuthServiceHandlerFromEndpoint(ctx, mux, GRPCAuthEndpoint, opts); err != nil {
		return nil, fmt.Errorf("register auth service: %w", err)
	}

	if err := customerv1.RegisterCustomerServiceHandlerFromEndpoint(ctx, mux, GRPCCustomerEndpoint, opts); err != nil {
		return nil, fmt.Errorf("register customer service: %w", err)
	}

	return mux, nil
}

// errorHandler приводит ошибки gRPC к единому формату
func errorHandler(ctx context.Context, mux *runtime.ServeMux, m runtime.Marshaler, w http.ResponseWriter, r *http.Request, err error) {
	fmt.Println("[mux.errorHandler] detect error ", err)
	st := statusFromErr(err)
	code := runtime.HTTPStatusFromCode(st.Code())
	msg := st.Message()
	if code == http.StatusInternalServerError {
		msg = "internal server error!!!"
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	fmt.Fprintf(w, `{"code":%d,"message":%q}`, code, msg)
}

// headerMatcher пробрасывает X-Request-ID в gRPC metadata
func headerMatcher(key string) (string, bool) {
	switch key {
	case "X-Request-Id", "X-Request-ID":
		return "x-request-id", true
	default:
		return runtime.DefaultHeaderMatcher(key)
	}
}
