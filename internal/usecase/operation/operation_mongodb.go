package operation

import (
	"context"
	"fmt"
	"test_go/internal/repo"
	"test_go/pkg/logger"
)

type useCaseMongo struct {
	opRepo repo.OperationMongoRepo
	//opcRepo repo.OperationCommandsRepo
	//cRepo   repo.CommandRepo
	l logger.Interface
}

func NewMongo(
	opRepo repo.OperationMongoRepo,
//opCmdRepo repo.OperationCommandsRepo,
//cmdRepo repo.CommandRepo,
	l logger.Interface,
) *useCaseMongo {
	return &useCaseMongo{
		opRepo: opRepo,
		//opcRepo:       opCmdRepo,
		//cRepo:         cmdRepo,
		l: l,
	}
}
func (uc *useCaseMongo) DeleteOperation(ctx context.Context, id int64) error {
	op := "OperationUseCase - DeleteOperation"

	if err := uc.opRepo.DeleteById(ctx, id); err != nil {
		return fmt.Errorf("%s - uc.opRepo.DeleteById: %w", op, err)
	}
	return nil
}
