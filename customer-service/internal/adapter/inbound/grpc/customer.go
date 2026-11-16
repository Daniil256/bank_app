package grpc

import (
	"context"
	"errors"
	"fmt"
	"google.golang.org/protobuf/types/known/emptypb"

	"bank_app/shared/roles"
	"customer-service/internal/app/usecase"
	"customer-service/internal/domain"
	customerv1 "customer-service/proto/customer/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type CustomerServer struct {
	customerv1.UnimplementedCustomerServiceServer
	getCustomer   *usecase.GetCustomerUseCase
	updateProfile *usecase.UpdateProfileUseCase
	getStatus     *usecase.GetStatusUseCase
	listCustomers *usecase.ListCustomersUseCase // ← 4-й
}

func NewCustomerServer(
	getCustomer *usecase.GetCustomerUseCase,
	updateProfile *usecase.UpdateProfileUseCase,
	getStatus *usecase.GetStatusUseCase,
	listCustomers *usecase.ListCustomersUseCase, // ← 4-й
) *CustomerServer {
	return &CustomerServer{
		getCustomer:   getCustomer,
		updateProfile: updateProfile,
		getStatus:     getStatus,
		listCustomers: listCustomers,
	}
}

func (s *CustomerServer) GetCustomer(ctx context.Context, req *customerv1.GetCustomerRequest) (*customerv1.Customer, error) {
	fmt.Println("[GetCustomer] req", req)
	md, _ := metadata.FromIncomingContext(ctx)
	role := getMeta(md, "x-user-role")

	if role != roles.Admin && role != roles.Officer {
		return nil, mapError(domain.ErrAccessDenied)
	}

	customer, err := s.getCustomer.Execute(ctx, req.UserId)
	if err != nil {
		fmt.Println("[GetCustomer] Execute err", err)
		return nil, mapError(err)
	}
	proto, err := toProto(customer)
	if err != nil {
		fmt.Println("[GetCustomer] err toProto", err)
	}
	return proto, nil
}

func (s *CustomerServer) GetClient(ctx context.Context, req *emptypb.Empty) (*customerv1.Client, error) {
	fmt.Println("[GetClient] req", req)
	md, _ := metadata.FromIncomingContext(ctx)
	role := getMeta(md, "x-user-role")

	if role != roles.Client {
		return nil, mapError(domain.ErrAccessDenied)
	}

	requesterUserID := getMeta(md, "x-user-id")
	customer, err := s.getCustomer.Execute(ctx, requesterUserID)
	if err != nil {
		fmt.Println("[GetClient] err", err)
		return nil, mapError(err)
	}
	proto, _ := toProto(customer)
	return toClient(proto), nil
}

func (s *CustomerServer) UpdateProfile(ctx context.Context, req *customerv1.UpdateProfileRequest) (*customerv1.Client, error) {
	fmt.Println("[UpdateProfile] req", req)
	md, _ := metadata.FromIncomingContext(ctx)
	role := getMeta(md, "x-user-role")
	UserID := getMeta(md, "x-user-id")

	// Только CLIENT может редактировать свой профиль
	if role != roles.Client {
		return nil, status.Error(codes.PermissionDenied, "only CLIENT can update profile")
	}

	customer, err := s.updateProfile.Execute(ctx, UserID, usecase.UpdateProfileInput{
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		BirthDate:   req.BirthDate,
		Address:     req.Address,
		Phone:       req.Phone,
		Citizenship: req.Citizenship,
	})
	if err != nil {
		fmt.Println("[UpdateProfile] err", err)
		return nil, mapError(err)
	}
	proto, _ := toProto(customer)
	return toClient(proto), nil
}

func (s *CustomerServer) GetCustomerStatus(ctx context.Context, req *customerv1.GetCustomerStatusRequest) (*customerv1.GetCustomerStatusResponse, error) {
	fmt.Println("[GetCustomerStatus] req", req)
	md, _ := metadata.FromIncomingContext(ctx)
	role := getMeta(md, "x-user-role")

	if role != roles.Admin && role != roles.Officer {
		return nil, mapError(domain.ErrAccessDenied)
	}
	customer, err := s.getStatus.Execute(ctx, req.UserId)
	if err != nil {
		fmt.Println("[GetCustomerStatus] err", err)
		return nil, mapError(err)
	}
	return &customerv1.GetCustomerStatusResponse{
		UserId: customer.UserID,
		Status: customer.Status,
	}, nil
}

func (s *CustomerServer) SystemGetCustomerStatus(ctx context.Context, req *customerv1.GetCustomerStatusRequest) (*customerv1.GetCustomerStatusResponse, error) {
	fmt.Println("[SystemGetCustomerStatus] req", req)
	customer, err := s.getStatus.Execute(ctx, req.UserId)
	if err != nil {
		fmt.Println("[SystemGetCustomerStatus] err", err)
		return nil, mapError(err)
	}
	return &customerv1.GetCustomerStatusResponse{
		UserId: customer.UserID,
		Status: customer.Status,
	}, nil
}

func (s *CustomerServer) SystemGetCustomer(ctx context.Context, req *customerv1.GetCustomerRequest) (*customerv1.GetCustomerResponse, error) {
	fmt.Println("[SystemGetCustomer] req", req)
	customer, err := s.getCustomer.Execute(ctx, req.UserId)
	if err != nil {
		fmt.Println("[SystemGetCustomer] err", err)
		return nil, mapError(err)
	}
	return &customerv1.GetCustomerResponse{
		UserId:    customer.UserID,
		FirstName: customer.FirstName,
		LastName:  customer.LastName,
		BirthDate: customer.BirthDate.Format("2006-01-02"),
	}, nil
}

func (s *CustomerServer) GetClientStatus(ctx context.Context, req *emptypb.Empty) (*customerv1.GetClientStatusResponse, error) {
	fmt.Println("[GetCustomerStatus] req", req)
	md, _ := metadata.FromIncomingContext(ctx)
	role := getMeta(md, "x-user-role")
	UserID := getMeta(md, "x-user-id")
	customer, err := s.getStatus.Execute(ctx, UserID)

	if role != roles.Client {
		return nil, status.Error(codes.PermissionDenied, "only CLIENT can update self profile")
	}

	if err != nil {
		fmt.Println("[GetCustomerStatus] err", err)
		return nil, mapError(err)
	}
	return &customerv1.GetClientStatusResponse{
		Status: customer.Status,
	}, nil
}

func (s *CustomerServer) ListCustomers(ctx context.Context, req *customerv1.ListCustomersRequest) (*customerv1.ListCustomersResponse, error) {
	fmt.Println("[ListCustomers] req", req)
	md, _ := metadata.FromIncomingContext(ctx)
	role := getMeta(md, "x-user-role")

	if role != roles.Admin && role != roles.Officer {
		return nil, mapError(domain.ErrAccessDenied)
	}
	res, err := s.listCustomers.Execute(ctx, domain.ListFilter{
		Status: req.Status,
		Search: req.Search,
		Limit:  int(req.PageSize),
		Cursor: req.PageToken,
	})
	if err != nil {
		fmt.Println("[ListCustomers] err", err)
		return nil, mapError(err)
	}

	out := make([]*customerv1.Customer, 0, len(res.Customers))
	for i := range res.Customers {
		proto, _ := toProto(&res.Customers[i])
		out = append(out, proto)
	}
	return &customerv1.ListCustomersResponse{
		Customers:     out,
		NextPageToken: res.NextToken,
	}, nil
}

func getMeta(md metadata.MD, key string) string {
	v := md.Get(key)
	if len(v) == 0 {
		fmt.Println("[customer.getMeta] not found ", key)
		return ""
	}
	return v[0]
}

func toProto(c *domain.Customer) (*customerv1.Customer, error) {
	birthDate := ""
	if c.BirthDate != nil {
		birthDate = c.BirthDate.Format("2006-01-02")
	}
	return &customerv1.Customer{
		Id:          c.ID,
		UserId:      c.UserID,
		FirstName:   c.FirstName,
		LastName:    c.LastName,
		BirthDate:   birthDate,
		Address:     c.Address,
		Phone:       c.Phone,
		Citizenship: c.Citizenship,
		Status:      c.Status,
		CreatedAt:   c.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   c.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}
func toClient(c *customerv1.Customer) *customerv1.Client {
	return &customerv1.Client{
		FirstName:   c.FirstName,
		LastName:    c.LastName,
		BirthDate:   c.BirthDate,
		Address:     c.Address,
		Phone:       c.Phone,
		Citizenship: c.Citizenship,
		Status:      c.Status,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}

func mapError(err error) error {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return status.Error(codes.NotFound, "customer not found")
	case errors.Is(err, domain.ErrAccessDenied):
		return status.Error(codes.PermissionDenied, "access denied")
	case errors.Is(err, domain.ErrCannotEditOnKyc):
		return status.Error(codes.FailedPrecondition, "cannot edit profile after KYC started")
	case errors.Is(err, domain.ErrInvalidProfile):
		return status.Error(codes.InvalidArgument, "invalid profile data")
	case errors.Is(err, domain.ErrUserIDRequired):
		return status.Error(codes.FailedPrecondition, "customer_id required")
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
