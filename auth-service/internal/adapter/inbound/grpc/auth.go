package grpc

import (
	"context"
	"errors"
	"fmt"
	"google.golang.org/grpc/metadata"
	"log"

	"auth-service/internal/app/usecase"
	"auth-service/internal/domain"
	authv1 "auth-service/proto/auth/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type AuthServer struct {
	authv1.UnimplementedAuthServiceServer
	register      *usecase.RegisterUseCase
	login         *usecase.LoginUseCase
	refresh       *usecase.RefreshUseCase
	logout        *usecase.LogoutUseCase
	createOfficer *usecase.CreateOfficerUseCase
}

func NewAuthServer(
	register *usecase.RegisterUseCase,
	login *usecase.LoginUseCase,
	refresh *usecase.RefreshUseCase,
	logout *usecase.LogoutUseCase,
	createOfficer *usecase.CreateOfficerUseCase,
) *AuthServer {
	return &AuthServer{
		register:      register,
		login:         login,
		refresh:       refresh,
		logout:        logout,
		createOfficer: createOfficer,
	}
}

func mapError(err error) error {
	switch {
	case errors.Is(err, domain.ErrEmailExists):
		return status.Error(codes.AlreadyExists, "email already taken")
	case errors.Is(err, domain.ErrInvalidEmail):
		return status.Error(codes.InvalidArgument, "the email address is incorrect")
	case errors.Is(err, domain.ErrWeakPassword):
		return status.Error(codes.InvalidArgument, "password too weak")
	case errors.Is(err, domain.ErrInvalidCredentials):
		return status.Error(codes.Unauthenticated, "invalid username or password")
	case errors.Is(err, domain.ErrInvalidRefreshToken):
		return status.Error(codes.Unauthenticated, "invalid refresh token")
	case errors.Is(err, domain.ErrInvalidLogout):
		return status.Error(codes.InvalidArgument, "invalid logout")
	default:
		return status.Error(codes.Internal, "internal error")
	}
}

func (s *AuthServer) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	fmt.Println("[auth.Register] RegisterRequest ", req)

	userID, err := s.register.Execute(ctx, req.Email, req.Password)
	if err != nil {
		fmt.Println("[auth.Register] error ", err)
		return nil, mapError(err)
	}
	response := authv1.RegisterResponse{UserId: userID}
	fmt.Println("[auth.Register] RegisterResponse ", req)

	return &response, nil
}

func (s *AuthServer) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	fmt.Println("[auth.Login] LoginRequest ", req)
	tokens, err := s.login.Execute(ctx, req.Email, req.Password)
	if err != nil {
		fmt.Println("[auth.Login] error ", err)
		return nil, mapError(err)
	}

	response := authv1.LoginResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.ExpiresIn,
	}
	fmt.Println("[auth.Login] LoginResponse ", &response)

	return &response, nil
}

func (s *AuthServer) Refresh(ctx context.Context, req *authv1.RefreshRequest) (*authv1.LoginResponse, error) {
	fmt.Println("[auth.Refresh] RefreshRequest ", req)
	tokens, err := s.refresh.Execute(ctx, req.RefreshToken)
	if err != nil {
		fmt.Println("[auth.Refresh] error ", err)
		return nil, mapError(domain.ErrInvalidRefreshToken)
	}
	fmt.Println("[auth.Refresh] TokenPair ", tokens)

	if tokens == nil {
		return nil, mapError(domain.ErrInvalidRefreshToken)
	}
	response := authv1.LoginResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.ExpiresIn,
	}
	fmt.Println("[auth.Refresh] LoginResponse ", &response)
	return &response, nil
}

func (s *AuthServer) Logout(ctx context.Context, req *authv1.LogoutRequest) (*emptypb.Empty, error) {
	fmt.Println("[auth.Logout] LogoutRequest ", req)
	if err := s.logout.Execute(ctx, req.RefreshToken); err != nil {
		fmt.Println("[auth.Logout] error ", err)
		return nil, mapError(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *AuthServer) CreateOfficer(ctx context.Context, req *authv1.CreateOfficerRequest) (*authv1.RegisterResponse, error) {
	fmt.Println("[auth.CreateOfficer] CreateOfficerRequest ", req)
	// 1. Извлекаем role из metadata (прокинут gateway)
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing metadata")
	}
	log.Println("[auth.CreateOfficer] metadata: ", md)

	roles := md.Get("x-user-role")
	if len(roles) == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing role")
	}

	// 2. Проверяем, что role == "ADMIN"
	if roles[0] != domain.RoleAdmin {
		return nil, status.Error(codes.PermissionDenied, "only ADMIN can create officers")
	}

	// 3. Создаём OFFICER
	userID, err := s.createOfficer.Execute(ctx, req.Email, req.Password)
	if err != nil {
		return nil, mapError(err)
	}

	return &authv1.RegisterResponse{UserId: userID}, nil
}
