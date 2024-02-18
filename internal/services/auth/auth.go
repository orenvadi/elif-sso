package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/orenvadi/auth-grpc/internal/domain/models"
	"github.com/orenvadi/auth-grpc/internal/lib/jwt"
	"github.com/orenvadi/auth-grpc/internal/lib/jwt/logger/sl"
	"github.com/orenvadi/auth-grpc/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

type Auth struct {
	log         *slog.Logger
	usrSaver    UserSaver
	usrProvider UserProvider
	appProvider AppProvider
	tokenTTL    time.Duration
}

type UserSaver interface {
	SaveUser(ctx context.Context, firstName, lastName, phoneNumber, email string, passwordHash []byte) (uid int64, err error)
}

type UserProvider interface {
	User(ctx context.Context, email string) (models.User, error)
	IsAdmin(ctx context.Context, id int64) (bool, error)
}

type AppProvider interface {
	App(ctx context.Context, appID int) (models.App, error)
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidAppID       = errors.New("invalid app_id")
	ErrUserAlreadyExists  = errors.New("user already exists")
)

// New return a new instance of the Auth service.
func New(
	log *slog.Logger,
	userSaver UserSaver,
	userProvider UserProvider,
	appProvider AppProvider,
	tokenTTL time.Duration,
) *Auth {
	return &Auth{
		log:         log,
		usrSaver:    userSaver,
		usrProvider: userProvider,
		appProvider: appProvider,
		tokenTTL:    tokenTTL,
	}
}

// Login checks if user with given credentials exists in the system and returns access token.
//
// If user exists, but password is incorrect, returns error.
// If user doesn't exist, returns error.
func (a *Auth) Login(ctx context.Context, email, password string, appID int) (accessToken string, err error) {
	const op = "auth.Login"

	log := a.log.With(
		slog.String("op: ", op),
		// slog.String("email: ", email), // do not do that
	)

	log.Info("attempting to login user")

	user, err := a.usrProvider.User(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			a.log.Warn("user not found", sl.Err(err))

			return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}

		log.Warn("user not found", sl.Err(err))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password)); err != nil {
		log.Info("invalid credentials", sl.Err(err))

		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	app, err := a.appProvider.App(ctx, appID)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	log.Info("user logged in successfully")

	accessToken, err = jwt.NewToken(user, app, a.tokenTTL)
	if err != nil {
		log.Error("failed to generate token", sl.Err(err))

		return "", fmt.Errorf("%s: %w", op, err)
	}

	return accessToken, nil
}

func (a *Auth) RegisterNewUser(ctx context.Context, firstName, lastName, phoneNumber, email, password string) (userID int64, accessToken, refreshToken string, err error) {
	const op = "auth.RegisterNewUser"

	log := a.log.With(
		slog.String("op: ", op),
		// slog.String("email: ", email), // do not do that
	)

	log.Info("registering user")

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate password hash", sl.Err(err))
		return 0, "", "", fmt.Errorf("%s: %w", op, err)
	}

	id, err := a.usrSaver.SaveUser(ctx, firstName, lastName, phoneNumber, email, passwordHash)
	if err != nil {

		if errors.Is(err, storage.ErrUserExists) {
			a.log.Warn("user already exists", sl.Err(err))

			return 0, "", "", fmt.Errorf("%s: %w", op, ErrUserAlreadyExists)
		}

		log.Error("failed to save user", sl.Err(err))
		return 0, "", "", fmt.Errorf("%s: %w", op, err)
	}

	user := models.User{
		ID:           id,
		FirstName:    firstName,
		LastName:     lastName,
		PhoneNumber:  phoneNumber,
		Email:        email,
		PasswordHash: passwordHash,
	}

	app := models.App{}

	accessToken, err = jwt.NewToken(user, app, a.tokenTTL)
	if err != nil {
		log.Error("failed to generate token", sl.Err(err))

		return 0, "", "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user registered")

	return id, accessToken, "", nil
}

func (a *Auth) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const op = "Auth.IsAdmin"

	log := a.log.With(
		slog.String("op", op),
		slog.Int64("user_id", userID),
	)

	log.Info("checking if user is admin")

	isAdmin, err := a.usrProvider.IsAdmin(ctx, userID)
	if err != nil {

		if errors.Is(err, storage.ErrAppNotFound) {
			a.log.Warn("app not found", sl.Err(err))

			return false, fmt.Errorf("%s: %w", op, err)
		}

		return false, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("checked if user is admin", slog.Bool("is_admin", isAdmin))

	return isAdmin, nil
}
