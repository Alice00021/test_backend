package di

import (
	"github.com/Alice00021/test_common/pkg/jwt"
	"github.com/Alice00021/test_common/pkg/logger"
	"github.com/Alice00021/test_common/pkg/transactional"
	"sync"
	"test_go/config"
	"test_go/internal/usecase"
	"test_go/internal/usecase/auth"
	"test_go/internal/usecase/author"
	"test_go/internal/usecase/book"
	"test_go/internal/usecase/command"
	"test_go/internal/usecase/export"
	"test_go/internal/usecase/operation"
	"test_go/internal/usecase/user"
)

type UseCase struct {
	Auth           usecase.Auth
	User           usecase.User
	Book           usecase.Book
	Author         usecase.Author
	Export         usecase.Export
	Command        usecase.Command
	Operation      usecase.Operation
	CommandMongo   usecase.CommandMongo
	OperationMongo usecase.OperationMongo
}

func NewUseCase(
	t transactional.Transactional,
	repo *Repo,
	l logger.Interface,
	conf *config.Config,
	jwtManager *jwt.JWTManager,
) *UseCase {
	txMtx := &sync.Mutex{}
	authUc := auth.New(t, l, repo.UserRepo, jwtManager, conf.LocalFileStorage.BasePath, &conf.EmailConfig, txMtx)
	userUc := user.New(t, l, repo.UserRepo, jwtManager, conf.LocalFileStorage.BasePath, &conf.EmailConfig, txMtx)
	authorUc := author.New(t, repo.AuthorRepo, l)
	bookUc := book.New(t, repo.BookRepo, l)
	commandUc := command.New(t, repo.CommandRepo, conf.LocalFileStorage, l)
	commandMongoUc := command.NewMongo(repo.CommandMongoRepo, conf.LocalFileStorage, l)
	OperationMongoUc := operation.NewMongo(repo.OperationMongoRepo, repo.CommandMongoRepo, l)
	operationUc := operation.New(t, repo.OperationRepo, repo.OperationCommandsRepo, repo.CommandRepo, l)
	exportUc := export.New(authorUc, bookUc, commandUc, operationUc, l, conf.LocalFileStorage.ExportPath)

	return &UseCase{
		Auth:           authUc,
		Author:         authorUc,
		Book:           bookUc,
		User:           userUc,
		Export:         exportUc,
		Command:        commandUc,
		CommandMongo:   commandMongoUc,
		Operation:      operationUc,
		OperationMongo: OperationMongoUc,
	}
}
