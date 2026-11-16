package grpc

import (
	"bank_app/shared/roles"
	"context"
	"errors"
	"fmt"
	"google.golang.org/protobuf/types/known/emptypb"
	"time"

	"kyc-service/internal/app/usecase"
	"kyc-service/internal/domain"
	kycv1 "kyc-service/proto/kyc/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type KycServer struct {
	kycv1.UnimplementedKycServiceServer
	submit    *usecase.SubmitUseCase
	get       *usecase.GetUseCase
	getClient *usecase.GetClientUseCase
	list      *usecase.ListUseCase
	claimUC   *usecase.ClaimUseCase
	approve   *usecase.ApproveUseCase
	reject    *usecase.RejectUseCase // ← добавить

}

func NewKycServer(submit *usecase.SubmitUseCase,
	get *usecase.GetUseCase,
	list *usecase.ListUseCase,
	getClient *usecase.GetClientUseCase,
	claimUC *usecase.ClaimUseCase,
	approve *usecase.ApproveUseCase,
	reject *usecase.RejectUseCase, // ← добавить
) *KycServer {
	return &KycServer{submit: submit, get: get, list: list, getClient: getClient, claimUC: claimUC, approve: approve, reject: reject}
}

func (s *KycServer) SubmitApplication(ctx context.Context, _ *emptypb.Empty) (*kycv1.Application, error) {
	fmt.Println("[kyc.SubmitApplication]")

	md, _ := metadata.FromIncomingContext(ctx)
	role := getMeta(md, "x-user-role")
	userID := getMeta(md, "x-user-id")

	if role != roles.Client {
		fmt.Println("[kyc.SubmitApplication] err: only CLIENT")
		return nil, status.Error(codes.PermissionDenied, "only CLIENT can submit")
	}
	if userID == "" {
		fmt.Println("[kyc.SubmitApplication] err: userID missing")
		return nil, status.Error(codes.FailedPrecondition, "userID required")
	}

	app, err := s.submit.Execute(ctx, userID)
	if err != nil {
		fmt.Println("[kyc.SubmitApplication] Execute err", err)
		return nil, mapError(err)
	}

	return toProto(app), nil
}

func (s *KycServer) GetApplication(ctx context.Context, req *kycv1.GetApplicationRequest) (*kycv1.Application, error) {
	fmt.Println("[kyc.GetApplication] req", req)

	md, _ := metadata.FromIncomingContext(ctx)
	role := getMeta(md, "x-user-role")
	if role != roles.Admin && role != roles.Officer {
		fmt.Println("[kyc.GetApplication] err: only ADMIN or OFFICER can get")
		return nil, status.Error(codes.PermissionDenied, "Permission denied")
	}

	app, err := s.get.Execute(ctx, req.ApplicationId)
	if err != nil {
		fmt.Println("[kyc.GetApplication] Execute err", err)
		return nil, mapError(err)
	}

	return toProto(app), nil
}

func (s *KycServer) GetClientApplication(ctx context.Context, _ *emptypb.Empty) (*kycv1.Application, error) {
	fmt.Println("[kyc.GetClientApplication]")

	md, _ := metadata.FromIncomingContext(ctx)
	role := getMeta(md, "x-user-role")
	if role != roles.Client {
		fmt.Println("[kyc.GetClientApplication] err: only CLIENT can get self application")
		return nil, status.Error(codes.PermissionDenied, "only CLIENT can get self application")
	}
	userId := getMeta(md, "x-user-id")

	app, err := s.getClient.Execute(ctx, userId)
	if err != nil {
		fmt.Println("[kyc.GetClientApplication] Execute err", err)
		return nil, mapError(err)
	}

	return toProto(app), nil
}

func (s *KycServer) ListApplications(ctx context.Context, req *kycv1.ListApplicationsRequest) (*kycv1.ListApplicationsResponse, error) {
	fmt.Println("[kyc.ListApplications] req", req)

	md, _ := metadata.FromIncomingContext(ctx)
	role := getMeta(md, "x-user-role")

	apps, nextToken, err := s.list.Execute(ctx, role, req.Status, int(req.PageSize), req.PageToken)
	if err != nil {
		fmt.Println("[kyc.ListApplications] Execute err", err)
		return nil, mapError(err)
	}

	out := make([]*kycv1.Application, 0, len(apps))
	for i := range apps {
		out = append(out, toProto(&apps[i]))
	}
	return &kycv1.ListApplicationsResponse{
		Applications:  out,
		NextPageToken: nextToken,
	}, nil
}

func (s *KycServer) ClaimApplication(ctx context.Context, req *kycv1.ClaimApplicationRequest) (*kycv1.Application, error) {
	fmt.Println("[kyc.ClaimApplication] req", req)

	md, _ := metadata.FromIncomingContext(ctx)
	role := getMeta(md, "x-user-role")
	userID := getMeta(md, "x-user-id")

	if role != "OFFICER" && role != "ADMIN" {
		fmt.Println("[kyc.ClaimApplication] err: only OFFICER/ADMIN")
		return nil, status.Error(codes.PermissionDenied, "only OFFICER/ADMIN")
	}
	if userID == "" {
		fmt.Println("[kyc.ClaimApplication] err: userID missing")
		return nil, status.Error(codes.Unauthenticated, "userID required")
	}

	app, err := s.claimUC.Execute(ctx, req.ApplicationId, userID)
	if err != nil {
		fmt.Println("[kyc.ClaimApplication] Execute err", err)
		return nil, mapError(err)
	}

	fmt.Println("[kyc.ClaimApplication] success application_id=", app.ID)
	return toProto(app), nil
}

func (s *KycServer) ApproveApplication(ctx context.Context, req *kycv1.ApproveApplicationRequest) (*kycv1.Application, error) {
	fmt.Println("[kyc.ApproveApplication] req", req)

	md, _ := metadata.FromIncomingContext(ctx)
	role := getMeta(md, "x-user-role")
	officerID := getMeta(md, "x-user-id")

	if role != "OFFICER" && role != "ADMIN" {
		fmt.Println("[kyc.ApproveApplication] err: only OFFICER/ADMIN")
		return nil, status.Error(codes.PermissionDenied, "only OFFICER/ADMIN")
	}
	if officerID == "" {
		fmt.Println("[kyc.ApproveApplication] err: officer_id missing")
		return nil, status.Error(codes.Unauthenticated, "officer_id required")
	}

	app, err := s.approve.Execute(ctx, req.ApplicationId, officerID)
	if err != nil {
		fmt.Println("[kyc.ApproveApplication] Execute err", err)
		return nil, mapError(err)
	}

	fmt.Println("[kyc.ApproveApplication] success application_id=", app.ID)
	return toProto(app), nil
}

func getMeta(md metadata.MD, key string) string {
	v := md.Get(key)
	if len(v) == 0 {
		return ""
	}
	return v[0]
}

func toProto(a *domain.Application) *kycv1.Application {
	p := &kycv1.Application{
		Id:        a.ID,
		UserId:    a.UserID,
		Status:    string(a.Status),
		CreatedAt: a.CreatedAt.Format(time.RFC3339),
		UpdatedAt: a.UpdatedAt.Format(time.RFC3339),
	}
	if a.ClaimedBy != nil {
		p.ClaimedBy = *a.ClaimedBy
	}
	if a.RejectionReason != nil {
		p.RejectionReason = *a.RejectionReason
	}
	return p
}

func (s *KycServer) RejectApplication(ctx context.Context, req *kycv1.RejectApplicationRequest) (*kycv1.Application, error) {
	fmt.Println("[kyc.RejectApplication] req", req)

	md, _ := metadata.FromIncomingContext(ctx)
	role := getMeta(md, "x-user-role")
	officerID := getMeta(md, "x-user-id")

	if role != "OFFICER" && role != "ADMIN" {
		fmt.Println("[kyc.RejectApplication] err: only OFFICER/ADMIN")
		return nil, status.Error(codes.PermissionDenied, "only OFFICER/ADMIN")
	}
	if officerID == "" {
		fmt.Println("[kyc.RejectApplication] err: officer_id missing")
		return nil, status.Error(codes.Unauthenticated, "officer_id required")
	}

	app, err := s.reject.Execute(ctx, req.ApplicationId, officerID, req.Reason)
	if err != nil {
		fmt.Println("[kyc.RejectApplication] Execute err", err)
		return nil, mapError(err)
	}

	fmt.Println("[kyc.RejectApplication] success application_id=", app.ID, "reason=", req.Reason)
	return toProto(app), nil
}

func mapError(err error) error {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return status.Error(codes.NotFound, "application not found")
	case errors.Is(err, domain.ErrActiveExists):
		return status.Error(codes.AlreadyExists, "active application already exists")
	case errors.Is(err, domain.ErrProfileNotFilled):
		return status.Error(codes.FailedPrecondition, "profile must be filled")
	case errors.Is(err, domain.ErrDocumentsMissing):
		return status.Error(codes.FailedPrecondition, "required documents missing")
	case errors.Is(err, domain.ErrAccessDenied):
		return status.Error(codes.PermissionDenied, "access denied")
	case errors.Is(err, domain.ErrCustomerIDRequired):
		return status.Error(codes.FailedPrecondition, "customer_id required")
	case errors.Is(err, domain.ErrNotInReview):
		return status.Error(codes.FailedPrecondition, "application in not review")
	case errors.Is(err, domain.ErrNotClaimedByYou):
		return status.Error(codes.FailedPrecondition, "application claimed by another officer")
	case errors.Is(err, domain.ErrReasonRequired):
		return status.Error(codes.InvalidArgument, "rejection reason required")
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
