package operation

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
	opRepo  repo.OperationRepo
	opcRepo repo.OperationCommandsRepo
	cRepo   repo.CommandRepo
	l       logger.Interface
}

func New(
	t transactional.Transactional,
	opRepo repo.OperationRepo,
	opCmdRepo repo.OperationCommandsRepo,
	cmdRepo repo.CommandRepo,
	l logger.Interface,
) *useCase {
	return &useCase{
		Transactional: t,
		opRepo:        opRepo,
		opcRepo:       opCmdRepo,
		cRepo:         cmdRepo,
		l:             l,
	}
}

func (uc *useCase) CreateOperation(ctx context.Context, inp entity.CreateOperationInput) (*entity.Operation, error) {
	op := "OperationUseCase - CreateOperation"

	var operation entity.Operation
	if err := uc.RunInTransaction(ctx, func(txCtx context.Context) error {
		e := &entity.Operation{
			Name:        inp.Name,
			Description: inp.Description,
			Commands:    []*entity.OperationCommand{},
		}

		mapCommands, err := uc.cRepo.GetBySystemNames(txCtx)
		if err != nil {
			return fmt.Errorf("%s - uc.cRepo.GetBySystemNames: %w", op, err)
		}

		var operationCommands []*entity.OperationCommand
		mapContainer := make(map[entity.Address]entity.Container)
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

			operationCommand := &entity.OperationCommand{
				Command: command,
				Address: commandInput.Address,
			}

			operationCommands = append(operationCommands, operationCommand)
			totalTime += command.AverageTime
		}
		e.AverageTime = totalTime

		res, err := uc.opRepo.Create(txCtx, e)
		if err != nil {
			return fmt.Errorf("%s - uc.opRepo.Create: %w", op, err)
		}

		if err := uc.opcRepo.Create(txCtx, res.ID, operationCommands); err != nil {
			return fmt.Errorf("%s - uc.opсRepo.Create: %w", op, err)
		}

		operation = *res
		operation.Commands = operationCommands

		return nil
	}); err != nil {
		return nil, fmt.Errorf("%s - uc.RunInTransaction: %w", op, err)
	}

	return &operation, nil
}

func (uc *useCase) UpdateOperation(ctx context.Context, inp entity.UpdateOperationInput) error {
	op := "OperationUseCase - UpdateOperation"

	return uc.RunInTransaction(ctx, func(txCtx context.Context) error {
		mapCommands, err := uc.cRepo.GetBySystemNames(txCtx)
		if err != nil {
			return fmt.Errorf("%s - uc.cRepo.GetBySystemNames: %w", op, err)
		}

		var (
			operationCommands []*entity.OperationCommand
			mapContainer      = make(map[entity.Address]entity.Container)
			totalTime         int64
			commandsToCreate  []*entity.OperationCommand
			idsToKeep         []int64
		)

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
				return entity.ErrCommandNotFound
			} else {
				container.Volume += command.VolumeContainer
			}

			if !container.IsValidVolume() {
				return entity.ErrCommandVolumeExceeded
			}
			mapContainer[commandInput.Address] = container

			operationCommand := &entity.OperationCommand{
				ID:          0,
				OperationID: inp.ID,
				Command:     command,
				Address:     commandInput.Address,
			}

			operationCommands = append(operationCommands, operationCommand)
			totalTime += command.AverageTime

			// Если команда уже существует — обновляем
			if commandInput.ID != nil {
				operationCommand.ID = *commandInput.ID
				if err := uc.opcRepo.Update(txCtx, operationCommand); err != nil {
					return fmt.Errorf("%s - uc.opcRepo.Update: %w", op, err)
				}
				idsToKeep = append(idsToKeep, *commandInput.ID)
			} else {
				commandsToCreate = append(commandsToCreate, operationCommand)
			}
		}

		if err := uc.opcRepo.DeleteIfNotInOperationCommandIds(txCtx, inp.ID, idsToKeep); err != nil {
			return fmt.Errorf("%s - uc.opcRepo.DeleteExceptIDs: %w", op, err)
		}

		if len(commandsToCreate) > 0 {
			if err := uc.opcRepo.Create(txCtx, inp.ID, commandsToCreate); err != nil {
				return fmt.Errorf("%s - uc.opcRepo.Create: %w", op, err)
			}
		}

		operation := &entity.Operation{
			Entity:      entity.Entity{ID: inp.ID},
			Name:        inp.Name,
			Description: inp.Description,
			AverageTime: totalTime,
		}

		if err := uc.opRepo.Update(txCtx, operation); err != nil {
			return fmt.Errorf("%s - uc.opRepo.Update: %w", op, err)
		}

		return nil
	})
}

func (uc *useCase) GetOperations(ctx context.Context) (map[int64]*entity.Operation, error) {
	operations, err := uc.opRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("OperationUseCase - GetOperations - uc.opRepo.GetAll: %w", err)
	}

	return operations, nil
}

func (uc *useCase) GetOperation(ctx context.Context, id int64) (*entity.Operation, error) {
	book, err := uc.opRepo.GetById(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("OperationUseCase - GetOperation - uc.opRepo.GetById: %w", err)
	}

	return book, nil
}

func (uc *useCase) DeleteOperation(ctx context.Context, id int64) error {
	op := "OperationUseCase - DeleteOperation"

	return uc.RunInTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.opRepo.DeleteById(txCtx, id); err != nil {
			return fmt.Errorf("%s - uc.opRepo.DeleteById: %w", op, err)
		}

		if err := uc.opcRepo.DeleteByOperationId(txCtx, id); err != nil {
			return fmt.Errorf("%s - uc.opсRepo.DeleteByOperationId: %w", op, err)
		}
		return nil
	})
}
