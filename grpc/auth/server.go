package grpcauth

import (
	"context"
	"errors"
	"log"

	"github.com/bufbuild/protovalidate-go"
	"github.com/orenvadi/auth-grpc/internal/domain/models"
	"github.com/orenvadi/auth-grpc/internal/services/auth"
	"github.com/orenvadi/auth-grpc/internal/storage"
	ssov1 "github.com/orenvadi/auth-grpc/protos/gen/go/proto/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Auth interface {
	Login(ctx context.Context, email, password string, appID int64) (accessToken string, err error)
	RegisterNewUser(ctx context.Context, firstName, lastName, phoneNumber, email, password string, appID int64) (userID int64, accessToken, refreshToken string, err error)
	UpdateUser(ctx context.Context, firstName, lastName, phoneNumber, email string, appID int64) error
	IsAdmin(ctx context.Context, userID int64) (bool, error)
	ConfirmUserEmail(ctx context.Context, code string, appID int64) (success bool, err error)
	GetUserData(ctx context.Context, appID int64) (models.User, error)
	SendCodeToResetPassword(ctx context.Context, email string) error
	SetNewPassword(ctx context.Context, confirmCode, email string, newPassword string) error
}

type serverAPI struct {
	ssov1.UnimplementedAuthServer
	auth Auth
}

func Register(gRPC *grpc.Server, auth Auth) {
	ssov1.RegisterAuthServer(gRPC, &serverAPI{auth: auth})
}

const (
	emptyValue = 0
)

func (s *serverAPI) Login(ctx context.Context, req *ssov1.LoginRequest) (*ssov1.LoginResponse, error) {
	// DONE add rpc validation using 3rd party package
	v, err := protovalidate.New()
	if err != nil {
		log.Fatalln("error protovalidate", err)
	}

	// validating
	if err := v.Validate(req); err != nil {
		switch {

		case req.GetAppId() == emptyValue:
			return nil, status.Error(codes.InvalidArgument, "app_id is required")

		default:
			return nil, status.Error(codes.InvalidArgument, err.Error())

		}
	}

	// DONE: implement login via auth service

	accessToken, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword(), req.GetAppId())
	if err != nil {
		// DONE handle various error types

		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid email or password")
		}

		// cause it is internal service, and users have no access to us
		// we can return internal errors to the client services

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &ssov1.LoginResponse{
		AccessToken: accessToken,
	}, nil
}

func (s *serverAPI) Register(ctx context.Context, req *ssov1.RegisterRequest) (*ssov1.RegisterResponse, error) {
	v, err := protovalidate.New()
	if err != nil {
		log.Fatalln("error protovalidate", err)
	}

	// validating
	if err := v.Validate(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	userID, accessToken, refreshToken, err := s.auth.RegisterNewUser(ctx, req.GetFirstName(), req.GetLastName(), req.GetPhoneNumber(), req.GetEmail(), req.GetPassword(), req.GetAppId())
	if err != nil {
		// DONE handle various error types

		if errors.Is(err, storage.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}
		// because it is internal service, and users have no access to us
		// we can return internal errors to the client services

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &ssov1.RegisterResponse{
		UserId:       userID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *serverAPI) ConfirmUserEmail(ctx context.Context, req *ssov1.ConfirmUserEmailRequest) (confirmUserEmailResponse *ssov1.ConfirmUserEmailResponse, err error) {
	v, err := protovalidate.New()
	if err != nil {
		log.Fatalln("error protovalidate", err)
	}

	// validating
	if err := v.Validate(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	confirmSuccess, err := s.auth.ConfirmUserEmail(ctx, req.GetConfirmCode(), req.GetAppId())
	if err != nil {
		return nil, err
	}

	return &ssov1.ConfirmUserEmailResponse{
		Success: confirmSuccess,
	}, nil
}

// this took me 8 hours to debug
func (s *serverAPI) UpdateUser(ctx context.Context, req *ssov1.UpdateUserRequest) (updateUserResponse *ssov1.UpdateUserResponse, err error) {
	v, err := protovalidate.New()
	if err != nil {
		log.Fatalln("error protovalidate", err)
	}

	// validating
	if err := v.Validate(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err = s.auth.UpdateUser(ctx, req.GetFirstName(), req.GetLastName(), req.GetPhoneNumber(), req.GetEmail(), req.GetAppId()); err != nil {
		return nil, err
	}

	return &ssov1.UpdateUserResponse{
		Success: true,
	}, nil
}

func (s *serverAPI) IsAdmin(ctx context.Context, req *ssov1.IsAdminRequest) (*ssov1.IsAdminResponse, error) {
	v, err := protovalidate.New()
	if err != nil {
		log.Fatalln("error protovalidate", err)
	}

	// validating
	if err := v.Validate(req); err != nil {
		switch {

		case req.GetUserId() == emptyValue:
			return nil, status.Error(codes.InvalidArgument, "user_id is required")

		default:
			return nil, status.Error(codes.InvalidArgument, err.Error())

		}
	}

	isAdmin, err := s.auth.IsAdmin(ctx, req.UserId)
	if err != nil {
		// DONE handle various error types

		if errors.Is(err, storage.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}

		// cause it is internal service, and users have no access to us
		// we can return internal errors to the client services

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &ssov1.IsAdminResponse{
		IsAdmin: isAdmin,
	}, nil
}

func (s *serverAPI) GetUserData(ctx context.Context, req *ssov1.GetUserDataRequest) (*ssov1.GetUserDataResponse, error) {
	user, err := s.auth.GetUserData(ctx, req.AppId)
	if err != nil {
		return nil, err
	}

	return &ssov1.GetUserDataResponse{
		Id:          user.ID,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		PhoneNumber: user.PhoneNumber,
		CreatedAt:   timestamppb.New(user.CreatedAt),
		UpdatedAt:   timestamppb.New(user.UpdatedAt),
		Email:       user.Email,
	}, nil
}

func (s *serverAPI) SendCodeToResetPassword(ctx context.Context, req *ssov1.SendCodeToResetPasswordRequest) (*ssov1.SendCodeToResetPasswordResponse, error) {
	v, err := protovalidate.New()
	if err != nil {
		log.Fatalln("error protovalidate", err)
	}

	// validating
	if err := v.Validate(req); err != nil {
		switch {
		default:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}

	if err = s.auth.SendCodeToResetPassword(ctx, req.GetEmail()); err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}

	return &ssov1.SendCodeToResetPasswordResponse{Success: true}, nil
}

func (s *serverAPI) SetNewPassword(ctx context.Context, req *ssov1.SetNewPasswordRequest) (*ssov1.SetNewPasswordResponse, error) {
	v, err := protovalidate.New()
	if err != nil {
		log.Fatalln("error protovalidate", err)
	}

	// validating
	if err := v.Validate(req); err != nil {
		switch {
		default:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}

	if err = s.auth.SetNewPassword(ctx, req.GetConfirmCode(), req.GetEmail(), req.GetNewPassword()); err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}

	return &ssov1.SetNewPasswordResponse{Success: true}, nil
}
