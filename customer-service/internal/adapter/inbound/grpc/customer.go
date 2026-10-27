package grpc

import (
	"context"
	"errors"

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
	md, _ := metadata.FromIncomingContext(ctx)
	role := getMeta(md, "x-user-role")
	requesterCustomerID := getMeta(md, "x-customer-id")

	customer, err := s.getCustomer.Execute(ctx, req.CustomerId, role, requesterCustomerID)
	if err != nil {
		return nil, mapError(err)
	}
	return toProto(customer), nil
}

func (s *CustomerServer) UpdateProfile(ctx context.Context, req *customerv1.UpdateProfileRequest) (*customerv1.Customer, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	role := getMeta(md, "x-user-role")
	customerID := getMeta(md, "x-customer-id")

	// Только CLIENT может редактировать свой профиль
	if role != roles.Client {
		return nil, status.Error(codes.PermissionDenied, "only CLIENT can update profile")
	}
	if customerID == "" {
		return nil, status.Error(codes.FailedPrecondition, "customer_id required")
	}

	customer, err := s.updateProfile.Execute(ctx, customerID, usecase.UpdateProfileInput{
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		BirthDate:   req.BirthDate,
		Address:     req.Address,
		Phone:       req.Phone,
		Citizenship: req.Citizenship,
	})
	if err != nil {
		return nil, mapError(err)
	}
	return toProto(customer), nil
}

func (s *CustomerServer) GetCustomerStatus(ctx context.Context, req *customerv1.GetCustomerStatusRequest) (*customerv1.GetCustomerStatusResponse, error) {
	customer, err := s.getStatus.Execute(ctx, req.CustomerId)
	if err != nil {
		return nil, mapError(err)
	}
	return &customerv1.GetCustomerStatusResponse{
		CustomerId: customer.ID,
		Status:     customer.Status,
	}, nil
}

func (s *CustomerServer) ListCustomers(ctx context.Context, req *customerv1.ListCustomersRequest) (*customerv1.ListCustomersResponse, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	role := getMeta(md, "x-user-role")

	res, err := s.listCustomers.Execute(ctx, role, domain.ListFilter{
		Status: req.Status,
		Search: req.Search,
		Limit:  int(req.PageSize),
		Cursor: req.PageToken,
	})
	if err != nil {
		return nil, mapError(err)
	}

	out := make([]*customerv1.Customer, 0, len(res.Customers))
	for i := range res.Customers {
		out = append(out, toProto(&res.Customers[i]))
	}
	return &customerv1.ListCustomersResponse{
		Customers:     out,
		NextPageToken: res.NextToken,
	}, nil
}

func getMeta(md metadata.MD, key string) string {
	v := md.Get(key)
	if len(v) == 0 {
		return ""
	}
	return v[0]
}

func toProto(c *domain.Customer) *customerv1.Customer {
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
	case errors.Is(err, domain.ErrCustomerIDRequired):
		return status.Error(codes.FailedPrecondition, "customer_id required")
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
