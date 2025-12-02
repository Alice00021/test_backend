package user

import (
	"context"
	"errors"
	"fmt"
	"github.com/Alice00021/test_common/pkg/auth"
	"github.com/Alice00021/test_common/pkg/jwt"
	"github.com/Alice00021/test_common/pkg/logger"
	"github.com/Alice00021/test_common/pkg/transactional"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"test_go/config"
	"test_go/internal/entity"
	"test_go/internal/repo"
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

func (uc *useCase) GetUserByName(ctx context.Context, name string) (*entity.User, error) {
	user, err := uc.repo.GetByUserName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("UserUseCase - GetUserByName - uc.repo.GetByUserName: %w", err)
	}

	return user, nil
}

func (uc *useCase) GetUser(ctx context.Context, id int64) (*entity.User, error) {
	user, err := uc.repo.GetById(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("UserUseCase - GetUser - uc.repo.GetById: %w", err)
	}

	return user, nil
}

func (uc *useCase) GetUsers(ctx context.Context, filter entity.FilterUserInput) ([]*entity.User, error) {
	users, err := uc.repo.GetAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("UserUseCase - GetUsers - uc.repo.GetAll: %w", err)
	}

	return users, nil
}

func (uc *useCase) UpdateUser(ctx context.Context, inp entity.UpdateUserInput) error {
	if err := uc.RunInTransaction(ctx, func(txCtx context.Context) error {
		e := &entity.User{
			Entity:   entity.Entity{ID: inp.ID},
			Name:     inp.Name,
			Surname:  inp.Surname,
			Username: inp.Username,
			Rating:   inp.Rating,
		}
		if err := uc.repo.Update(txCtx, e); err != nil {
			return fmt.Errorf("uc.repo.Update: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("%s - uc.RunInTransaction: %w", err)
	}

	return nil
}

func (uc *useCase) ChangePassword(ctx context.Context, inp entity.ChangePasswordInput) error {
	if err := uc.RunInTransaction(ctx, func(txCtx context.Context) error {
		if inp.NewPassword != inp.ConfirmPassword {
			return entity.ErrPasswordMismatch
		}

		user, err := uc.repo.GetById(txCtx, inp.ID)
		if err != nil {
			return fmt.Errorf("UserUseCase - ChangePassword - uc.repo.GetById: %w", err)
		}

		if !auth.CheckPasswordHash(inp.OldPassword, user.Password) {
			return fmt.Errorf("UserUseCase - ChangePassword - auth.CheckPasswordHash: %w", err)
		}

		hashedPassword, err := auth.HashPassword(inp.NewPassword)
		if err != nil {
			return fmt.Errorf("UserUseCase - ChangePassword - auth.HashPassword: %w", err)
		}

		user.Password = hashedPassword

		if err := uc.repo.Update(txCtx, user); err != nil {
			return fmt.Errorf("uc.repo.Update: %w", err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("%s - uc.RunInTransaction: %w", err)
	}

	return nil
}

func (uc *useCase) SetProfilePhoto(ctx context.Context, id int64, file *multipart.FileHeader) error {
	if err := uc.RunInTransaction(ctx, func(txCtx context.Context) error {
		allowedPhoto := map[string]bool{
			".jpg":  true,
			".jpeg": true,
			".png":  true,
		}

		ext := strings.ToLower(filepath.Ext(file.Filename))

		if !allowedPhoto[ext] {
			return errors.New("invalid file type, only images (jpg, jpeg, png) are allowed")
		}

		if err := os.MkdirAll(uc.storageBasePath, 0755); err != nil {
			return fmt.Errorf("failed to create upload directory: %v", err)
		}

		user, err := uc.repo.GetById(ctx, id)
		if err != nil {
			return fmt.Errorf("UserUseCase - SetProfilePhoto - uc.repo.GetById: %w", err)
		}

		filename := fmt.Sprintf("%s_profile%s", user.Username, ext)
		filePath := filepath.Join(uc.storageBasePath, filename)

		src, err := file.Open()
		if err != nil {
			return entity.ErrOpenFile
		}
		defer src.Close()

		dst, err := os.Create(filePath)
		if err != nil {
			return entity.ErrCreateFile
		}

		if _, err := io.Copy(dst, src); err != nil {
			return entity.ErrSaveFile
		}
		defer dst.Close()

		user.FilePath = &filePath
		if err := uc.repo.Update(txCtx, user); err != nil {
			return fmt.Errorf("uc.repo.Update: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("%s - uc.RunInTransaction: %w", err)
	}

	return nil
}

func (uc *useCase) UpdateRating(ctx context.Context, id int64, rating float32) error {
	uc.mtx.Lock()
	defer uc.mtx.Unlock()

	if err := uc.RunInTransaction(ctx, func(txCtx context.Context) error {
		var user entity.User
		user.ID = id
		user.Rating = rating

		if err := uc.repo.Update(txCtx, &user); err != nil {
			return fmt.Errorf("uc.repo.Update: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("%s - uc.RunInTransaction: %w", err)
	}

	return nil
}
