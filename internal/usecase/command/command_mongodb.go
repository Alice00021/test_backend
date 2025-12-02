package command

import (
	"context"
	"encoding/json"
	"os"
	"test_go/config"

	"fmt"
	"github.com/Alice00021/test_common/pkg/logger"
	"test_go/internal/entity"
	"test_go/internal/repo"
)

type useCaseMongo struct {
	repo        repo.CommandMongoRepo
	jsonStorage config.LocalFileStorage
	l           logger.Interface
}

func NewMongo(
	repo repo.CommandMongoRepo,
	jsonStorage config.LocalFileStorage,
	l logger.Interface,
) *useCaseMongo {
	return &useCaseMongo{
		repo:        repo,
		jsonStorage: jsonStorage,
		l:           l,
	}
}

func (uc *useCaseMongo) UpdateCommands(ctx context.Context) error {
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

	existing, err := uc.repo.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("%s - repo.GetAll: %w", op, err)
	}

	existingMap := make(map[string]entity.Command)
	for _, cmd := range existing {
		existingMap[cmd.SystemName] = cmd
	}

	for i := range commands {
		cmd := &commands[i]

		if ex, ok := existingMap[cmd.SystemName]; ok {
			cmd.ID = ex.ID
			if err := uc.repo.Update(ctx, cmd.SystemName, cmd); err != nil {
				return fmt.Errorf("%s - repo.Update: %w", op, err)
			}
		} else {
			if _, err := uc.repo.Create(ctx, cmd); err != nil {
				return fmt.Errorf("%s - repo.Create: %w", op, err)
			}
		}
	}

	return nil
}

func (uc *useCaseMongo) GetCommands(ctx context.Context) (map[string]entity.Command, error) {
	commands, err := uc.repo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("CommandUseCase - GetCommands - repo.GetAll: %w", err)
	}

	result := make(map[string]entity.Command)
	for _, cmd := range commands {
		result[cmd.SystemName] = cmd
	}

	return result, nil
}
