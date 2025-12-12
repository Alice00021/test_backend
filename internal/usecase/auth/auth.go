package auth

import (
	"context"
	"crypto/rsa"
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
	cfg             config.Auth
	PrivateKey      *rsa.PrivateKey
	PublicKey       *rsa.PublicKey
	storageBasePath string
	emailConfig     *config.EmailConfig
	mtx             *sync.Mutex
}

func New(t transactional.Transactional,
	l logger.Interface,
	repo repo.UserRepo,
	cfg config.Auth,
	sbp string,
	emailConfig *config.EmailConfig,
	mtx *sync.Mutex,
) *useCase {
	//encryptionRsa, err := jwt.New(cfg.PrivateKeyFile, "")
	//if err != nil {
	//	l.Fatal("AuthUseCase - New - jwt.New error - %s", err)
	//}

	privateKey, err := jwt.DecodePrivateKey([]byte(cfg.PrivateKey))
	if err != nil {
		l.Fatal("AuthUseCase - New - jwt.DecodePrivateKey error - %s", err)
	}

	return &useCase{
		Transactional: t,
		l:             l,
		repo:          repo,
		cfg:           cfg,
		PrivateKey:    privateKey,
		PublicKey:     &privateKey.PublicKey,
		//PrivateKey:      encryptionRsa.PrivateKey,
		//PublicKey:       encryptionRsa.PublicKey,
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

		userInfo := &entity.UserInfoToken{
			ID:   0,
			Role: user.Role,
		}

		tokenPair, err := uc.generateTokens(userInfo)
		if err != nil {
			return fmt.Errorf("uc.generateTokens: %w", err)
		}

		verifyToken := tokenPair.AccessToken

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

	userInfo := &entity.UserInfoToken{
		ID:   user.ID,
		Role: user.Role,
	}

	tokenPair, err := uc.generateTokens(userInfo)
	if err != nil {
		return nil, fmt.Errorf("uc.generateTokens: %w", err)
	}

	return tokenPair, nil
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
	op := "AuthUseCase - RefreshTokens"

	claims, err := jwt.ValidateToken(refreshToken, uc.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("%s - jwt.ValidateToken: %w", op, entity.ErrInvalidRefreshToken)
	}

	data, ok := claims["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("%s - invalid token data: %w", op, entity.ErrInvalidRefreshToken)
	}

	userID := int64(data["id"].(float64))

	user, err := uc.repo.GetById(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("%s - uc.repo.GetById: %w", op, entity.ErrUserNotFound)
	}

	userInfo := &entity.UserInfoToken{
		ID:   user.ID,
		Role: user.Role,
	}

	tokenPair, err := uc.generateTokens(userInfo)
	if err != nil {
		return nil, fmt.Errorf("%s - uc.generateTokens: %w", op, err)
	}

	return tokenPair, nil
}

func (s *useCase) ValidateToken(ctx context.Context, token string) (*entity.UserInfoToken, error) {
	claims, err := jwt.ValidateToken(token, s.PublicKey)
	if err != nil {
		return nil, err
	}

	if data, ok := claims["data"].(map[string]interface{}); ok {
		id := data["id"]
		role := data["role"]

		return &entity.UserInfoToken{
			ID:   int64(id.(float64)),
			Role: entity.UserRole(role.(string)),
		}, nil
	}

	return nil, jwt.ErrInvalidToken

}

func (s *useCase) generateTokens(user *entity.UserInfoToken) (*entity.TokenPair, error) {
	data := make(map[string]interface{})
	data["id"] = user.ID
	data["role"] = user.Role

	accessToken, err := jwt.GenerateToken(s.cfg.AccessTokenExpiresIn, data, s.PrivateKey, "")
	if err != nil {
		return nil, fmt.Errorf("jwt token: %v", err)
	}

	refreshToken, err := jwt.GenerateToken(s.cfg.RefreshTokenExpiresIn, data, s.PrivateKey, "")
	if err != nil {
		return nil, fmt.Errorf("jwt token: %v", err)
	}

	return &entity.TokenPair{RefreshToken: refreshToken, AccessToken: accessToken}, nil
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
