package grpc

import (
	"context"
	"errors"

	"auth-service/internal/app/usecase"
	"auth-service/internal/domain"
	authv1 "auth-service/proto/auth/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthServer struct {
	authv1.UnimplementedAuthServiceServer
	register *usecase.RegisterUseCase
	login    *usecase.LoginUseCase
}

func NewAuthServer(register *usecase.RegisterUseCase, login *usecase.LoginUseCase) *AuthServer {
	return &AuthServer{register: register, login: login}
}

func mapError(err error) error {
	switch {
	case errors.Is(err, domain.ErrEmailTaken):
		return status.Error(codes.AlreadyExists, "email already taken")
	case errors.Is(err, domain.ErrInvalidEmail):
		return status.Error(codes.InvalidArgument, "invalid email")
	case errors.Is(err, domain.ErrWeakPassword):
		return status.Error(codes.InvalidArgument, "password too weak")
	case errors.Is(err, domain.ErrInvalidCredentials):
		return status.Error(codes.Unauthenticated, "invalid credentials")
	default:
		return status.Error(codes.Internal, "internal error")
	}
}

func (s *AuthServer) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	userID, err := s.register.Execute(ctx, req.Email, req.Password)
	if err != nil {
		return nil, mapError(err)
	}
	return &authv1.RegisterResponse{UserId: userID}, nil
}

func (s *AuthServer) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	user, err := s.login.Execute(ctx, req.Email, req.Password)
	if err != nil {
		return nil, mapError(err)
	}

	// TODO: подписать JWT. Пока — заглушка.
	return &authv1.LoginResponse{
		AccessToken:  "mock-access-token-for-" + user.ID,
		RefreshToken: "mock-refresh-token-for-" + user.ID,
		ExpiresIn:    900,
	}, nil
}
