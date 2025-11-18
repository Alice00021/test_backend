package repo

import (
	"context"
	"test_go/internal/entity"
)

type (
	UserRepo interface {
		Create(context.Context, *entity.User) (*entity.User, error)
		GetById(context.Context, int64) (*entity.User, error)
		Update(context.Context, *entity.User) error
		GetByUserName(context.Context, string) (*entity.User, error)
		GetAll(context.Context, entity.FilterUserInput) ([]*entity.User, error)
		GetByEmail(context.Context, string) (*entity.User, error)
		GetByVerifyToken(context.Context, string) (*entity.User, error)
	}

	AuthorRepo interface {
		Create(context.Context, *entity.Author) (*entity.Author, error)
		GetById(context.Context, int64) (*entity.Author, error)
		Update(context.Context, *entity.Author) error
		GetAll(context.Context) ([]*entity.Author, error)
		DeleteById(context.Context, int64) error
	}

	BookRepo interface {
		Create(context.Context, *entity.Book) (*entity.Book, error)
		GetById(context.Context, int64) (*entity.Book, error)
		Update(context.Context, *entity.Book) error
		GetAll(context.Context) ([]*entity.Book, error)
		DeleteById(context.Context, int64) error
	}

	CommandRepo interface {
		Create(context.Context, *entity.Command) (*entity.Command, error)
		GetById(context.Context, int64) (*entity.Command, error)
		Update(context.Context, *entity.Command) error
		GetBySystemNames(context.Context) (map[string]entity.Command, error)
	}

	OperationRepo interface {
		GetAll(context.Context) (map[int64]*entity.Operation, error)
		Create(context.Context, *entity.Operation) (*entity.Operation, error)
		GetById(context.Context, int64) (*entity.Operation, error)
		Update(context.Context, *entity.Operation) error
		DeleteById(context.Context, int64) error
	}

	OperationCommandsRepo interface {
		Create(context.Context, int64, []*entity.OperationCommand) error
		Update(context.Context, *entity.OperationCommand) error
		DeleteByOperationId(context.Context, int64) error
		DeleteIfNotInOperationCommandIds(context.Context, int64, []int64) error
	}
)
