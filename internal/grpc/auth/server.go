package server

import (
	"SSO/internal/domain/models"
	verfic "SSO/internal/lib/verifications"
	"SSO/internal/storage"
	"context"
	"errors"
	"github.com/google/uuid"

	ssov2 "github.com/AlexseyBrashka/protos/gen/go/sso"

	"SSO/internal/services/auth"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type serverAPI struct {
	auth Auth
	ssov2.UnimplementedAuthServer
}
type Auth interface {
	Login(
		ctx context.Context,
		email string,
		password string,
		appUUID uuid.UUID,
	) (accessToken string, refreshToken string, err error)

	RegisterNewUser(
		ctx context.Context,
		email string,
		password string,
	) (userUUID uuid.UUID, err error)

	Logout(
		ctx context.Context,
		email string,
		AppUUID uuid.UUID,
	) error

	AddPermission(
		ctx context.Context,
		appUUID uuid.UUID,
		permission string,
	) (uuid.UUID, error)

	RemovePermission(
		ctx context.Context,
		appUUID uuid.UUID,
		permissionUUID uuid.UUID) error

	GrantPermission(
		ctx context.Context,
		email string,
		AppUUID uuid.UUID,
		permissionUUID uuid.UUID) (string, string, error)

	RevokePermission(
		ctx context.Context,
		email string,
		appUUID uuid.UUID,
		permissionUUID uuid.UUID,
	) (string, string, error)

	RefreshToken(
		ctx context.Context,
		actualRefreshToken string,
	) (accessToken string, refreshToken string, err error)
	GetAppPermissions(
		ctx context.Context,
		appUUID uuid.UUID,
	) ([]models.Permission, error)
}

func Register(gRPCServer *grpc.Server, auth Auth) {
	ssov2.RegisterAuthServer(gRPCServer, &serverAPI{auth: auth})
}
func (s *serverAPI) Login(
	ctx context.Context,
	in *ssov2.LoginRequest,
) (*ssov2.LoginResponse, error) {

	_, err := verfic.VerifyEmail(in.Email)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid email")
	}

	if in.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "No password")
	}

	appUUID, err := uuid.Parse(in.GetAppUuid())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "No App")
	}

	accessToken, refreshToken, err := s.auth.Login(ctx, in.GetEmail(), in.GetPassword(), appUUID)

	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid email or password")
		}
		return nil, status.Error(codes.Internal, "failed to login")
	}

	return &ssov2.LoginResponse{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (s *serverAPI) Register(
	ctx context.Context,
	in *ssov2.RegisterRequest,
) (*ssov2.OperationResponse, error) {

	_, err := verfic.VerifyEmail(in.Email)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid email")
	}

	if in.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	_, err = s.auth.RegisterNewUser(ctx, in.GetEmail(), in.GetPassword())
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}

		return nil, status.Error(codes.Internal, "failed to register user")
	}

	return &ssov2.OperationResponse{Success: true}, nil

}

func (s *serverAPI) Logout(ctx context.Context, in *ssov2.LogoutRequest) (*ssov2.OperationResponse, error) {

	appUUID, err := uuid.Parse(in.GetAppUUID())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "No App")
	}

	_, err = verfic.VerifyEmail(in.GetEmail())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "incorrect email")
	}

	err = s.auth.Logout(ctx, in.GetEmail(), appUUID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to logout user")
	}
	return &ssov2.OperationResponse{Success: true}, nil

}

func (s *serverAPI) AddPermission(ctx context.Context, in *ssov2.AddPermissionRequest) (*ssov2.AddPermissionResponse, error) {

	appUUID, err := uuid.Parse(in.GetAppUUID())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "No App")
	}

	permissionUUID, err := s.auth.AddPermission(ctx, appUUID, in.PermissionName)

	if err != nil {
		return nil, status.Error(codes.Internal, "failed to add permission")
	}
	return &ssov2.AddPermissionResponse{UUID: permissionUUID.String()}, nil
}

func (s *serverAPI) RemovePermission(ctx context.Context, in *ssov2.RemovePermissionRequest) (*ssov2.OperationResponse, error) {

	appUUID, err := uuid.Parse(in.GetAppUUID())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "No App")
	}

	permUUID, err := uuid.Parse(in.GetPermissionUUID())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "No Permission")
	}

	err = s.auth.RemovePermission(ctx, appUUID, permUUID)

	if err != nil {
		return nil, status.Error(codes.Internal, "failed to remove permission")
	}
	return &ssov2.OperationResponse{Success: true}, nil

}

func (s *serverAPI) GrantPermission(ctx context.Context, in *ssov2.GrantPermissionRequest) (*ssov2.LoginResponse, error) {

	_, err := verfic.VerifyEmail(in.GetEmail())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "incorrect email")
	}

	appUUID, err := uuid.Parse(in.GetAppUUID())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "No App")
	}

	permUUID, err := uuid.Parse(in.GetPermissionUUID())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "No Permission")
	}

	accessToken, refreshToken, err := s.auth.GrantPermission(ctx, in.GetEmail(), appUUID, permUUID)

	if err != nil {
		return nil, status.Error(codes.Internal, "failed to grant permission")
	}
	return &ssov2.LoginResponse{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (s *serverAPI) RevokePermission(ctx context.Context, in *ssov2.RevokePermissionRequest) (*ssov2.LoginResponse, error) {

	_, err := verfic.VerifyEmail(in.GetEmail())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "incorrect email")
	}

	appUUID, err := uuid.Parse(in.GetAppUUID())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "No App")
	}

	permUUID, err := uuid.Parse(in.GetPermissionUUID())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "No Permission")
	}

	accessToken, refreshToken, err := s.auth.RevokePermission(ctx, in.GetEmail(), appUUID, permUUID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to revoke permission")
	}
	return &ssov2.LoginResponse{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (s *serverAPI) RefreshToken(ctx context.Context, in *ssov2.RefreshTokenRequest) (*ssov2.LoginResponse, error) {

	accessToken, refreshToken, err := s.auth.RefreshToken(ctx, in.Token)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to grant permission")
	}
	return &ssov2.LoginResponse{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}
func (s *serverAPI) GetAppPermissions(ctx context.Context, in *ssov2.GetAppPermissionsRequest) (*ssov2.GetAppPermissionsResponse, error) {

	appUUID, err := uuid.Parse(in.GetAppUUID())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "incorrect app uuid")
	}

	perms, err := s.auth.GetAppPermissions(ctx, appUUID)
	if err != nil {
		return nil, status.Error(codes.Internal, "cant get app's permissions")
	}

	var permissions []*ssov2.Perm

	for _, perm := range perms {
		permissions = append(permissions, &ssov2.Perm{UUID: perm.UUID.String(), Name: perm.Name})
	}
	return &ssov2.GetAppPermissionsResponse{Permissions: permissions}, nil
}
