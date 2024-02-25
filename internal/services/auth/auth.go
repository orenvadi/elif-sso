package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/smtp"
	"time"

	"github.com/orenvadi/auth-grpc/internal/domain/models"
	"github.com/orenvadi/auth-grpc/internal/lib/jwt"
	"github.com/orenvadi/auth-grpc/internal/lib/jwt/logger/sl"
	"github.com/orenvadi/auth-grpc/internal/lib/rnd"
	"github.com/orenvadi/auth-grpc/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

type Auth struct {
	log                  *slog.Logger
	usrSaver             UserSaver
	usrProvider          UserProvider
	usrUpdater           UserUpdater
	appProvider          AppProvider
	passwordResetter     PasswordResetter
	emailConfirmProvider EmailConfirmProvider
	tokenTTL             time.Duration
}

type UserSaver interface {
	SaveUser(ctx context.Context, firstName, lastName, phoneNumber, email string, passwordHash []byte) (uid int64, err error)
}

type UserUpdater interface {
	UpdateUser(ctx context.Context, usr models.User) (err error)
}

type UserProvider interface {
	User(ctx context.Context, email string) (models.User, error)
	UserAllData(ctx context.Context, id int64) (models.User, error)
	IsAdmin(ctx context.Context, id int64) (bool, error)
	UserEmailConfirm(ctx context.Context, userID int64) error
}

type AppProvider interface {
	App(ctx context.Context, appID int64) (models.App, error)
}

type EmailConfirmProvider interface {
	SaveConfirmationCode(ctx context.Context, userID int64, code string) error
	ConfirmationCode(ctx context.Context, userID int64) (confCode models.ConfirmCode, err error)
	DeleteConfirmationCode(ctx context.Context, user_id int64) error
}

type PasswordResetter interface {
	IsEmailConfirmed(ctx context.Context, email string) (bool, error)
	ChangePassword(ctx context.Context, email string, newPasswordHash []byte) error
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
	userUpdater UserUpdater,
	appProvider AppProvider,
	emailConfirmProvider EmailConfirmProvider,
	passwordResetter PasswordResetter,
	tokenTTL time.Duration,
) *Auth {
	return &Auth{
		log:                  log,
		usrSaver:             userSaver,
		usrProvider:          userProvider,
		usrUpdater:           userUpdater, // из-за этой херни я  потерял 3 часа
		appProvider:          appProvider,
		emailConfirmProvider: emailConfirmProvider,
		passwordResetter:     passwordResetter,
		tokenTTL:             tokenTTL,
	}
}

// Login checks if user with given credentials exists in the system and returns access token.
//
// If user exists, but password is incorrect, returns error.
// If user doesn't exist, returns error.
func (a *Auth) Login(ctx context.Context, email, password string, appID int64) (accessToken string, err error) {
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

	accessToken, err = jwtn.NewToken(user, app, a.tokenTTL)
	if err != nil {
		log.Error("failed to generate token", sl.Err(err))

		return "", fmt.Errorf("%s: %w", op, err)
	}

	return accessToken, nil
}

func (a *Auth) RegisterNewUser(ctx context.Context, firstName, lastName, phoneNumber, email, password string, appID int64) (userID int64, accessToken, refreshToken string, err error) {
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

	userID, err = a.usrSaver.SaveUser(ctx, firstName, lastName, phoneNumber, email, passwordHash)
	if err != nil {

		if errors.Is(err, storage.ErrUserExists) {
			a.log.Warn("user already exists", sl.Err(err))

			return 0, "", "", fmt.Errorf("%s: %w", op, ErrUserAlreadyExists)
		}

		log.Error("failed to save user", sl.Err(err))
		return 0, "", "", fmt.Errorf("%s: %w", op, err)
	}

	user := models.User{
		ID:           userID,
		FirstName:    firstName,
		LastName:     lastName,
		PhoneNumber:  phoneNumber,
		Email:        email,
		PasswordHash: passwordHash,
	}

	app, err := a.appProvider.App(ctx, appID)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			a.log.Warn("user not found", sl.Err(err))

			return -1, "", "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}

		log.Warn("user not found", sl.Err(err))
		return -1, "", "", fmt.Errorf("%s: %w", op, err)
	}

	accessToken, err = jwtn.NewToken(user, app, a.tokenTTL)
	if err != nil {
		log.Error("failed to generate token", sl.Err(err))

		return 0, "", "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user registered")

	// Sending confirmation code
	rndCode := rnd.GenerateRandomNumber()

	if err = a.emailConfirmProvider.SaveConfirmationCode(ctx, user.ID, rndCode); err != nil {
		return 0, "", "", fmt.Errorf("%s: %w", op, err)
	}

	if err = sendVerificationEmail(user.Email, rndCode); err != nil {
		return 0, "", "", fmt.Errorf("%s: %w", op, err)
	}

	return userID, accessToken, "", nil
}

func sendVerificationEmail(email, code string) error {
	auth := smtp.PlainAuth("", "eligdigital@gmail.com", "dqwqqgtxbbuwobgt", "smtp.gmail.com")
	to := []string{email}

	htmlMsg := `
    <html>
    <body>
        <h1 style="text-align: center;">Email Verification Code</h1>
        <p style="text-align: center; font-size: 20px;">Your verification code is:</p>
        <div style="text-align: center; font-size: 30px; border: 2px solid #000; padding: 10px; margin: 20px;">` + code + `</div>
    </body>
    </html>
    `

	err := smtp.SendMail("smtp.gmail.com:587", auth, "eligdigital@gmail.com", to, []byte(
		"From: eligdigital@gmail.com\r\n"+
			"To: "+email+"\r\n"+
			"Subject: Email Verification Code\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/html; charset=utf-8\r\n"+
			"\r\n"+
			htmlMsg,
	))
	if err != nil {
		return err
	}
	return nil
}

func (a *Auth) ConfirmUserEmail(ctx context.Context, confirmCode string, appID int64) (success bool, err error) {
	const op = "auth.ConfirmUserEmail"

	log := a.log.With(
		slog.String("op: ", op),
		// slog.String("user_email", email),
	)

	// Extract username from token

	app, err := a.appProvider.App(ctx, appID)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			a.log.Warn("user not found", sl.Err(err))

			return false, fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}

		log.Warn("user not found", sl.Err(err))
		return false, fmt.Errorf("%s: %w", op, err)
	}

	claims, err := jwtn.ValidateToken(ctx, app)
	if err != nil {
		return false, fmt.Errorf("invalid token claims")
	}
	userID := claims["uid"].(float64)
	uid := int64(userID)

	// logicccc

	log.Info("confirming user")

	confCodeFromDB, err := a.emailConfirmProvider.ConfirmationCode(ctx, uid)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	now := time.Now()

	if confirmCode != confCodeFromDB.Code {
		return false, fmt.Errorf("%s: invalid confirm code", op)
	}

	if now.Sub(confCodeFromDB.CreatedAt) > (5 * time.Minute) {
		return false, fmt.Errorf("%s: confirm code is expired", op)
	}

	if err = a.usrProvider.UserEmailConfirm(ctx, uid); err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	if err = a.emailConfirmProvider.DeleteConfirmationCode(ctx, confCodeFromDB.ID); err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	// logicccc end

	log.Info("user confirmed")
	return true, nil
}

// UpdateUser updates user information.
func (a *Auth) UpdateUser(ctx context.Context, firstName, lastName, phoneNumber, email string, appID int64) error {
	const op = "auth.UpdateUser"

	log := a.log.With(
		slog.String("op: ", op),
		// slog.String("user_email", email),
	)

	// Extract username from token

	app, err := a.appProvider.App(ctx, appID)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			a.log.Warn("user not found", sl.Err(err))

			return fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}

		log.Warn("user not found", sl.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	claims, err := jwtn.ValidateToken(ctx, app)
	if err != nil {
		return fmt.Errorf("invalid token claims")
	}
	userID := claims["uid"].(float64)
	uid := int64(userID)

	log.Info("updating user")

	// Retrieve the user from the storage
	user, err := a.usrProvider.UserAllData(ctx, uid)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			a.log.Warn("user not found", sl.Err(err))
			return ErrInvalidCredentials // or ErrUserNotFound
		}

		log.Warn("user not found", sl.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	// Update user information
	user.FirstName = firstName
	user.LastName = lastName
	user.PhoneNumber = phoneNumber
	user.Email = email

	// Hash and update password if provided
	// if password != "" {
	// 	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	// 	if err != nil {
	// 		log.Error("failed to generate password hash", sl.Err(err))
	// 		return fmt.Errorf("%s: %w", op, err)
	// 	}
	// 	user.PasswordHash = passwordHash
	// }

	// log.Info("upd: ", sl.Err(fmt.Errorf(fmt.Sprintf("%v", user))))

	// Save updated user information to the storage

	err = a.usrUpdater.UpdateUser(ctx, user)
	if err != nil {
		log.Error("failed to update user", sl.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user updated successfully")

	return nil
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

func (a *Auth) GetUserData(ctx context.Context, appID int64) (models.User, error) {
	const op = "auth.GetUserData"

	log := a.log.With(
		slog.String("op: ", op),
		// slog.String("user_email", email),
	)

	// Extract username from token

	app, err := a.appProvider.App(ctx, appID)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			a.log.Warn("user not found", sl.Err(err))

			return models.User{}, fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}

		log.Warn("user not found", sl.Err(err))
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	claims, err := jwtn.ValidateToken(ctx, app)
	if err != nil {
		return models.User{}, fmt.Errorf("invalid token claims")
	}
	userID := claims["uid"].(float64)
	uid := int64(userID)

	user, err := a.usrProvider.UserAllData(ctx, uid)
	if err != nil {
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (a *Auth) SendCodeToResetPassword(ctx context.Context, email string) error {
	const op = "auth.SendCodeToResetPassword"

	isEmailConfirmed, err := a.passwordResetter.IsEmailConfirmed(ctx, email)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if !isEmailConfirmed {
		return fmt.Errorf("%s: %w", op, errors.New("email is not confirmed"))
	}

	user, err := a.usrProvider.User(ctx, email)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	// Sending confirmation code
	rndCode := rnd.GenerateRandomNumber()

	if err = sendVerificationEmail(email, rndCode); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err = a.emailConfirmProvider.SaveConfirmationCode(ctx, user.ID, rndCode); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *Auth) SetNewPassword(ctx context.Context, confirmCode, email string, newPassword string) error {
	const op = "auth.SetNewPassword"

	log := a.log.With(
		slog.String("op: ", op),
		// slog.String("email: ", email), // do not do that
	)

	user, err := a.usrProvider.User(ctx, email)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	userAllData, err := a.usrProvider.UserAllData(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate password hash", sl.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	uid := userAllData.ID

	confCodeFromDB, err := a.emailConfirmProvider.ConfirmationCode(ctx, uid)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	location, _ := time.LoadLocation("Asia/Bishkek")
	now := time.Now().In(location)
	// now := time.Now()

	if confirmCode != confCodeFromDB.Code {
		return fmt.Errorf("%s: invalid confirm code", op)
	}

	if elapsedTime := now.Sub(confCodeFromDB.CreatedAt.In(location)); elapsedTime > (5 * time.Minute) {
		return fmt.Errorf("%s: confirm code is expired", op)
	}

	if err = a.usrProvider.UserEmailConfirm(ctx, uid); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err = a.emailConfirmProvider.DeleteConfirmationCode(ctx, confCodeFromDB.ID); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err = a.passwordResetter.ChangePassword(ctx, userAllData.Email, passwordHash); err != nil {
		log.Error("failed to change pass", sl.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
