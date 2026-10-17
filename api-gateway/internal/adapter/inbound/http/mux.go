package http

import (
	"context"
	"fmt"
	"net/http"

	authv1 "api-gateway/proto/auth/v1"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// NewMux создаёт grpc-gateway ServeMux и регистрирует все сервисы
func NewMux(ctx context.Context, grpcEndpoint string) (http.Handler, error) {
	mux := runtime.NewServeMux(
		runtime.WithErrorHandler(errorHandler),
		runtime.WithIncomingHeaderMatcher(headerMatcher),
	)

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	if err := authv1.RegisterAuthServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts); err != nil {
		return nil, fmt.Errorf("register auth service: %w", err)
	}

	return mux, nil
}

// errorHandler приводит ошибки gRPC к единому формату
func errorHandler(ctx context.Context, mux *runtime.ServeMux, m runtime.Marshaler, w http.ResponseWriter, r *http.Request, err error) {
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
