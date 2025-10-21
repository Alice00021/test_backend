package di

import (
	"test_go/internal/repo"
	"test_go/internal/repo/persistent"
	"test_go/pkg/postgres"
)

type Repo struct {
	UserRepo              repo.UserRepo
	BookRepo              repo.BookRepo
	AuthorRepo            repo.AuthorRepo
	CommandRepo           repo.CommandRepo
	OperationRepo         repo.OperationRepo
	OperationCommandsRepo repo.OperationCommandsRepo
}

func NewRepo(pg *postgres.Postgres) *Repo {
	return &Repo{
		UserRepo:              persistent.NewUserRepo(pg),
		BookRepo:              persistent.NewBookRepo(pg),
		AuthorRepo:            persistent.NewAuthorRepo(pg),
		CommandRepo:           persistent.NewCommandRepo(pg),
		OperationRepo:         persistent.NewOperationRepo(pg),
		OperationCommandsRepo: persistent.NewOperationCommandsRepo(pg),
	}
}
