package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/Alice00021/test_common/pkg/auth"
	"github.com/Alice00021/test_common/pkg/jwt"
	"github.com/Alice00021/test_common/pkg/logger"
	"github.com/Alice00021/test_common/pkg/transactional"
	"sync"
	"test_go/config"
	"test_go/internal/entity"
	"test_go/internal/repo"

	"gopkg.in/gomail.v2"
)

type useCase struct {
	transactional.Transactional
	l               logger.Interface
	repo            repo.UserRepo
	jwtManager      *jwt.JWTManager
	storageBasePath string
	emailConfig     *config.EmailConfig
	mtx             *sync.Mutex
}

func New(t transactional.Transactional,
	l logger.Interface,
	repo repo.UserRepo,
	jwtManager *jwt.JWTManager,
	sbp string,
	emailConfig *config.EmailConfig,
	mtx *sync.Mutex,
) *useCase {
	return &useCase{
		Transactional:   t,
		l:               l,
		repo:            repo,
		jwtManager:      jwtManager,
		storageBasePath: sbp,
		emailConfig:     emailConfig,
		mtx:             mtx,
	}
}

func (uc *useCase) Register(ctx context.Context, inp entity.CreateUserInput) (*entity.User, error) {
	op := "AuthUseCase - Register"

	var user entity.User
	if err := uc.RunInTransaction(ctx, func(txCtx context.Context) error {
		_, err := uc.repo.GetByEmail(txCtx, inp.Email)
		if err == nil {
			return entity.ErrEmailAlreadyUsed
		}

		if !errors.Is(err, entity.ErrUserNotFound) {
			return fmt.Errorf("uc.repo.GetByEmail: %w", err)
		}

		verifyToken, err := generateVerifyToken()
		if err != nil {
			return entity.ErrGenerateVerifyToken
		}

		e := entity.NewUser(
			inp.Name, inp.Surname, inp.Username, inp.Password, inp.Email,
		)
		e.VerifyToken = &verifyToken

		e.Rating = 50

		hashedPassword, err := auth.HashPassword(e.Password)
		if err != nil {
			return err
		}

		e.Password = hashedPassword

		res, err := uc.repo.Create(txCtx, e)
		if err != nil {
			return fmt.Errorf("uc.repo.Create: %w", err)
		}

		if err := uc.sendVerificationEmail(e.Email, verifyToken); err != nil {
			return fmt.Errorf("uc.sendVerificationEmail: %w", err)
		}
		user = *res
		return nil
	}); err != nil {
		return nil, fmt.Errorf("%s - uc.RunInTransaction: %w", op, err)
	}

	return &user, nil
}

func (uc *useCase) Login(ctx context.Context, username string, password string) (*entity.TokenPair, error) {
	op := "AuthUseCase - Login"

	var tokenPair entity.TokenPair
	user, err := uc.repo.GetByUserName(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("%s - uc.repo.GetByUserName: %w", op, err)
	}

	if !user.IsVerified {
		return nil, fmt.Errorf("%s - %w", op, entity.ErrEmailNotVerified)
	}

	if !auth.CheckPasswordHash(password, user.Password) {
		return nil, fmt.Errorf("%s - invalid credentials", op)
	}

	accessToken, err := uc.jwtManager.GenerateAccessToken(user.ID, user.Username)
	if err != nil {
		return nil, fmt.Errorf("%s - uc.jwtManager.GenerateAccessToken: %w", op, err)
	}

	refreshToken, err := uc.jwtManager.GenerateRefreshToken(user.ID, user.Username)
	if err != nil {
		return nil, fmt.Errorf("%s - uc.jwtManager.GenerateRefreshToken: %w", op, err)
	}
	tokenPair.AccessToken = accessToken
	tokenPair.RefreshToken = refreshToken

	return &tokenPair, nil
}

func (uc *useCase) VerifyEmail(ctx context.Context, token string) error {
	if err := uc.RunInTransaction(ctx, func(txCtx context.Context) error {
		user, err := uc.repo.GetByVerifyToken(txCtx, token)
		if err != nil {
			return fmt.Errorf("uc.repo.GetByVerifyToken: %w", err)
		}

		user.IsVerified = true
		user.VerifyToken = nil

		if err := uc.repo.Update(txCtx, user); err != nil {
			return fmt.Errorf("uc.repo.Update: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("%s - uc.RunInTransaction: %w", err)
	}

	return nil
}

func (uc *useCase) RefreshTokens(ctx context.Context, refreshToken string) (*entity.TokenPair, error) {

	claims, err := uc.jwtManager.ParseToken(refreshToken)
	if err != nil {
		return nil, entity.ErrInvalidRefreshToken
	}
	user, err := uc.repo.GetByUserName(ctx, claims.Username)
	if err != nil {
		return nil, entity.ErrUserNotFound
	}
	var token entity.TokenPair

	acessToken, err := uc.jwtManager.GenerateAccessToken(user.ID, user.Username)
	if err != nil {
		return nil, fmt.Errorf("uc.jwtManager.GenerateAccessToken: %w", err)
	}

	newRefreshToken, err := uc.jwtManager.GenerateRefreshToken(user.ID, user.Username)
	if err != nil {
		return nil, fmt.Errorf("uc.jwtManager.GenerateRefreshToken: %w", err)
	}

	token.AccessToken = acessToken
	token.RefreshToken = newRefreshToken

	return &token, nil
}

func (uc *useCase) Validation(ctx context.Context, tokenString string) (*entity.UserInfoToken, error) {
	op := "AuthUseCase - Validation"

	claims, err := uc.jwtManager.ParseToken(tokenString)
	if err != nil {
		if errors.Is(err, entity.ErrExpiredToken) {
			return nil, entity.ErrExpiredToken
		}
		return nil, entity.ErrInvalidToken
	}

	user, err := uc.repo.GetById(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("%s - uc.repo.GetByID: %w", op, err)
	}

	userInfo := &entity.UserInfoToken{
		ID:   user.ID,
		Role: user.Role,
	}

	return userInfo, nil
}

func generateVerifyToken() (string, error) {
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (uc *useCase) sendVerificationEmail(email, token string) error {
	message := gomail.NewMessage()
	message.SetHeader("From", uc.emailConfig.SenderEmail)
	message.SetHeader("To", email)
	message.SetHeader("Subject", "Email Verification")

	verificationLink := fmt.Sprintf("%s?token=%s", uc.emailConfig.VerifyBaseURL, token)
	body := fmt.Sprintf("Please verify your email by clicking the following link: %s", verificationLink)
	message.SetBody("text/plain", body)

	d := gomail.NewDialer(uc.emailConfig.SMTPHost, uc.emailConfig.SMTPPort, uc.emailConfig.SenderEmail, uc.emailConfig.SenderPassword)

	return d.DialAndSend(message)
}
