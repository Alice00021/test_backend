package di

import (
	m "github.com/Alice00021/test_common/pkg/mongodb"
	"github.com/Alice00021/test_common/pkg/postgres"
	"test_go/internal/repo"
	mongodb "test_go/internal/repo/mongodb"
	"test_go/internal/repo/persistent"
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
