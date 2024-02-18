package grpcauth

import (
	"context"
	"errors"
	"log"

	"github.com/bufbuild/protovalidate-go"
	"github.com/orenvadi/auth-grpc/internal/services/auth"
	"github.com/orenvadi/auth-grpc/internal/storage"
	ssov1 "github.com/orenvadi/auth-grpc/protos/gen/go/proto/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Auth interface {
	Login(ctx context.Context, email, password string, appID int) (accessToken string, err error)
	RegisterNewUser(ctx context.Context, firstName, lastName, phoneNumber, email, password string) (userID int64, accessToken, refreshToken string, err error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
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

	accessToken, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword(), int(req.GetAppId()))
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

	userID, accessToken, refreshToken, err := s.auth.RegisterNewUser(ctx, req.GetFirstName(), req.GetLastName(), req.GetPhoneNumber(), req.GetEmail(), req.GetPassword())
	if err != nil {
		// DONE handle various error types

		if errors.Is(err, storage.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}
		// cause it is internal service, and users have no access to us
		// we can return internal errors to the client services

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &ssov1.RegisterResponse{
		UserId:       userID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
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
			return nil, status.Error(codes.InvalidArgument, "app_id is required")

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
