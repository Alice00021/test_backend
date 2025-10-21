package command

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"test_go/config"
	"test_go/internal/entity"
	"test_go/internal/repo"
	"test_go/pkg/logger"
	"test_go/pkg/transactional"
)

type useCase struct {
	transactional.Transactional
	repo        repo.CommandRepo
	jsonStorage config.LocalFileStorage
	l           logger.Interface
}

func New(t transactional.Transactional,
	repo repo.CommandRepo,
	jsonStorage config.LocalFileStorage,
	l logger.Interface,
) *useCase {
	return &useCase{
		Transactional: t,
		repo:          repo,
		jsonStorage:   jsonStorage,
		l:             l,
	}
}

func (uc *useCase) UpdateCommands(ctx context.Context) error {
	op := "CommandUseCase - UpdateCommands"

	file, err := os.Open(uc.jsonStorage.JsonPath)
	if err != nil {
		return fmt.Errorf("%s - os.Open: %w", op, err)
	}
	defer file.Close()

	var commands []entity.Command
	if err := json.NewDecoder(file).Decode(&commands); err != nil {
		return fmt.Errorf("%s - json.Decode: %w", op, err)
	}

	if err := uc.RunInTransaction(ctx, func(txCtx context.Context) error {
		mapCommands, err := uc.repo.GetBySystemNames(txCtx)
		if err != nil {
			return fmt.Errorf("%s - uc.repo.GetBySystemNames: %w", op, err)
		}

		for i := range commands {
			cmd := &commands[i]
			if existingCmd, exists := mapCommands[cmd.SystemName]; exists {
				cmd.ID = existingCmd.ID
				if err := uc.repo.Update(txCtx, cmd); err != nil {
					return fmt.Errorf("%s - uc.repo.Update: %w", op, err)
				}
			} else {
				if _, err := uc.repo.Create(txCtx, cmd); err != nil {
					return fmt.Errorf("%s - uc.repo.Create: %w", op, err)
				}
			}
		}
		return nil
	}); err != nil {
		return fmt.Errorf("%s - uc.RunInTransaction: %w", op, err)
	}
	return nil
}
