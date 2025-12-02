package author

import (
	"context"
	"fmt"
	"github.com/Alice00021/test_common/pkg/logger"
	"github.com/Alice00021/test_common/pkg/transactional"

	"test_go/internal/entity"
	"test_go/internal/repo"
)

type useCase struct {
	transactional.Transactional
	repo repo.AuthorRepo
	l    logger.Interface
}

func New(t transactional.Transactional,
	repo repo.AuthorRepo,
	l logger.Interface,
) *useCase {
	return &useCase{
		Transactional: t,
		repo:          repo,
		l:             l,
	}
}

func (uc *useCase) CreateAuthor(ctx context.Context, inp entity.CreateAuthorInput) (*entity.Author, error) {
	op := "AuthorUseCase - CreateAuthor"

	var author entity.Author
	if err := uc.RunInTransaction(ctx, func(txCtx context.Context) error {
		e := entity.NewAuthor(
			inp.Name, inp.Gender,
		)
		res, err := uc.repo.Create(txCtx, e)
		if err != nil {
			return fmt.Errorf("uc.repo.Create: %w", err)
		}

		author = *res

		return nil
	}); err != nil {
		return nil, fmt.Errorf("%s - uc.RunInTransaction: %w", op, err)
	}

	return &author, nil
}

func (uc *useCase) UpdateAuthor(ctx context.Context, inp entity.UpdateAuthorInput) error {
	if err := uc.RunInTransaction(ctx, func(txCtx context.Context) error {
		e := &entity.Author{
			Entity: entity.Entity{ID: inp.ID},
			Name:   inp.Name,
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

func (uc *useCase) GetAuthor(ctx context.Context, id int64) (*entity.Author, error) {
	author, err := uc.repo.GetById(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("AuthorUseCase - GetAuthor - uc.repo.GetById: %w", err)
	}

	return author, nil
}

func (uc *useCase) GetAuthors(ctx context.Context) ([]*entity.Author, error) {
	authors, err := uc.repo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("AuthorUseCase - GetAuthors - uc.repo.GetAll: %w", err)
	}

	return authors, nil
}

func (uc *useCase) DeleteAuthor(ctx context.Context, id int64) error {
	op := "AuthorUseCase - DeleteAuthor"

	if err := uc.RunInTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.repo.DeleteById(txCtx, id); err != nil {
			return fmt.Errorf("uc.repo.DeleteById: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("%s - uc.RunInTransaction: %w", op, err)
	}

	return nil
}
