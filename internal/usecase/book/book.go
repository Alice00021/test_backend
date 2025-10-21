package book

import (
	"context"
	"fmt"
	"test_go/internal/entity"
	"test_go/internal/repo"
	"test_go/pkg/logger"
	"test_go/pkg/transactional"
)

type useCase struct {
	transactional.Transactional
	repo repo.BookRepo
	l    logger.Interface
}

func New(t transactional.Transactional,
	repo repo.BookRepo,
	l logger.Interface,
) *useCase {
	return &useCase{
		Transactional: t,
		repo:          repo,
		l:             l,
	}
}

func (uc *useCase) CreateBook(ctx context.Context, inp entity.CreateBookInput) (*entity.Book, error) {
	op := "BookUseCase - CreateBook"

	var book entity.Book
	if err := uc.RunInTransaction(ctx, func(txCtx context.Context) error {
		e := entity.NewBook(
			inp.Title, inp.AuthorId,
		)
		res, err := uc.repo.Create(txCtx, e)
		if err != nil {
			return fmt.Errorf("uc.repo.Create: %w", err)
		}

		book = *res

		return nil
	}); err != nil {
		return nil, fmt.Errorf("%s - uc.RunInTransaction: %w", op, err)
	}

	return &book, nil
}

func (uc *useCase) UpdateBook(ctx context.Context, inp entity.UpdateBookInput) error {
	if err := uc.RunInTransaction(ctx, func(txCtx context.Context) error {
		e := &entity.Book{
			Entity:   entity.Entity{ID: inp.ID},
			Title:    inp.Title,
			AuthorId: inp.AuthorId,
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

func (uc *useCase) GetBook(ctx context.Context, id int64) (*entity.Book, error) {
	book, err := uc.repo.GetById(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("BookUseCase - GetBook - uc.repo.GetById: %w", err)
	}

	return book, nil
}

func (uc *useCase) GetBooks(ctx context.Context) ([]*entity.Book, error) {
	books, err := uc.repo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("BookUseCase - GetBooks - uc.repo.GetAll: %w", err)
	}

	return books, nil
}

func (uc *useCase) DeleteBook(ctx context.Context, id int64) error {
	op := "BookUseCase - DeleteBook"

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
