package operation

import (
	"context"
	"fmt"
	"github.com/Alice00021/test_common/pkg/logger"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"test_go/internal/entity"
	"test_go/internal/repo"
)

type useCaseMongo struct {
	opRepo repo.OperationMongoRepo
	cRepo  repo.CommandMongoRepo
	l      logger.Interface
}

func NewMongo(
	opRepo repo.OperationMongoRepo,
	cRepo repo.CommandMongoRepo,
	l logger.Interface,
) *useCaseMongo {
	return &useCaseMongo{
		opRepo: opRepo,
		cRepo:  cRepo,
		l:      l,
	}
}
func (uc *useCaseMongo) CreateOperation(ctx context.Context, inp entity.CreateOperationInput) (*entity.OperationMongo, error) {
	op := "OperationUseCase - CreateOperation"

	mapCommands, err := uc.cRepo.GetBySystemName(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s - uc.cRepo.GetBySystemNames: %w", op, err)
	}

	mapContainer := make(map[entity.Address]entity.Container)
	var operationCommands []*entity.OperationCommand
	var totalTime int64

	for _, commandInput := range inp.Commands {
		command, ok := mapCommands[commandInput.SystemName]
		if !ok {
			return nil, entity.ErrCommandNotFound
		}

		container, ok := mapContainer[commandInput.Address]
		if !ok {
			container = entity.Container{
				Address:     commandInput.Address,
				ReagentType: command.Reagent,
				Volume:      command.VolumeContainer,
			}
		} else if container.ReagentType != command.Reagent {
			return nil, entity.ErrCommandDuplicateAddress
		} else {
			container.Volume += command.VolumeContainer
		}

		if !container.IsValidVolume() {
			return nil, entity.ErrCommandVolumeExceeded
		}
		mapContainer[commandInput.Address] = container

		operationCommands = append(operationCommands, &entity.OperationCommand{
			Command: command,
			Address: commandInput.Address,
		})
		totalTime += command.AverageTime
	}

	opDoc := &entity.OperationMongo{
		Name:        inp.Name,
		Description: inp.Description,
		AverageTime: totalTime,
		Commands:    operationCommands,
	}

	res, err := uc.opRepo.Create(ctx, opDoc)
	if err != nil {
		return nil, fmt.Errorf("%s - uc.opRepo.Create: %w", op, err)
	}

	return res, nil
}

func (uc *useCaseMongo) UpdateOperation(ctx context.Context, inp entity.UpdateOperationInputMongo) error {
	op := "OperationUseCase - UpdateOperation"

	mapCommands, err := uc.cRepo.GetBySystemName(ctx)
	if err != nil {
		return fmt.Errorf("%s - uc.cRepo.GetBySystemName: %w", op, err)
	}

	mapContainer := make(map[entity.Address]entity.Container)
	var operationCommands []*entity.OperationCommand
	var totalTime int64

	for _, commandInput := range inp.Commands {
		command, ok := mapCommands[commandInput.SystemName]
		if !ok {
			return entity.ErrCommandNotFound
		}

		container, ok := mapContainer[commandInput.Address]
		if !ok {
			container = entity.Container{
				Address:     commandInput.Address,
				ReagentType: command.Reagent,
				Volume:      command.VolumeContainer,
			}
		} else if container.ReagentType != command.Reagent {
			return entity.ErrCommandDuplicateAddress
		} else {
			container.Volume += command.VolumeContainer
		}

		if !container.IsValidVolume() {
			return entity.ErrCommandVolumeExceeded
		}
		mapContainer[commandInput.Address] = container

		operationCommands = append(operationCommands, &entity.OperationCommand{
			Command: command,
			Address: commandInput.Address,
		})
		totalTime += command.AverageTime
	}

	operation := &entity.OperationMongo{
		Id:          inp.ID,
		Name:        inp.Name,
		Description: inp.Description,
		AverageTime: totalTime,
		Commands:    operationCommands,
	}

	if err := uc.opRepo.Update(ctx, operation); err != nil {
		return fmt.Errorf("%s - uc.opRepo.Update: %w", op, err)
	}

	return nil
}

func (uc *useCaseMongo) GetOperations(ctx context.Context) ([]*entity.OperationMongo, error) {
	operations, err := uc.opRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("OperationUseCase - GetOperations - uc.opRepo.GetAll: %w", err)
	}

	return operations, nil
}

func (uc *useCaseMongo) DeleteOperation(ctx context.Context, id primitive.ObjectID) error {
	op := "OperationUseCase - DeleteOperation"

	if err := uc.opRepo.DeleteById(ctx, id); err != nil {
		return fmt.Errorf("%s - uc.opRepo.DeleteById: %w", op, err)
	}
	return nil
}
