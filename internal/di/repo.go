package di

import (
	"test_go/internal/repo"
	mongodb "test_go/internal/repo/mongodb"
	"test_go/internal/repo/persistent"
	m "test_go/pkg/mongodb"
	"test_go/pkg/postgres"
)

type Repo struct {
	UserRepo              repo.UserRepo
	BookRepo              repo.BookRepo
	AuthorRepo            repo.AuthorRepo
	CommandRepo           repo.CommandRepo
	OperationRepo         repo.OperationRepo
	OperationCommandsRepo repo.OperationCommandsRepo
	CommandMongoRepo      repo.CommandMongoRepo
	OperationMongoRepo    repo.OperationMongoRepo
}

func NewRepo(pg *postgres.Postgres, mongoClient *m.Client) *Repo {
	return &Repo{
		UserRepo:              persistent.NewUserRepo(pg),
		BookRepo:              persistent.NewBookRepo(pg),
		AuthorRepo:            persistent.NewAuthorRepo(pg),
		CommandRepo:           persistent.NewCommandRepo(pg),
		OperationRepo:         persistent.NewOperationRepo(pg),
		OperationCommandsRepo: persistent.NewOperationCommandsRepo(pg),
		CommandMongoRepo:      mongodb.NewCommandRepo(mongoClient),
		OperationMongoRepo:    mongodb.NewOperationRepo(mongoClient),
	}
}
