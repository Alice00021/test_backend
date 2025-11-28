package mongodb

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"test_go/internal/entity"
	"test_go/pkg/mongodb"
)

type CommandMongoRepo struct {
	coll *mongo.Collection
}

func NewCommandRepo(client *mongodb.Client) *CommandMongoRepo {
	return &CommandMongoRepo{
		coll: client.Collection("commands"),
	}
}

func (r *CommandMongoRepo) Create(ctx context.Context, cmd *entity.Command) (*entity.Command, error) {
	op := "CommandMongoDBRepo - Create"

	_, err := r.coll.InsertOne(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("%s - r.coll.InsertOne: %w", op, err)
	}
	return cmd, nil
}

func (r *CommandMongoRepo) GetBySystemName(ctx context.Context, systemName string) (*entity.Command, error) {
	op := "CommandMongoDBRepo - GetBySystemName"
	var cmd entity.Command

	err := r.coll.FindOne(ctx, bson.M{"systemName": systemName}).Decode(&cmd)
	if err != nil {
		return nil, fmt.Errorf("%s - r.coll.FindOne: %w", op, err)
	}
	return &cmd, nil
}

func (r *CommandMongoRepo) GetAll(ctx context.Context) ([]entity.Command, error) {
	op := "CommandMongoDBRepo - GetAll"

	cursor, err := r.coll.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("%s - r.coll.Find: %w", op, err)
	}
	defer cursor.Close(ctx)

	var cmds []entity.Command
	for cursor.Next(ctx) {
		var cmd entity.Command
		if err := cursor.Decode(&cmd); err != nil {
			return nil, fmt.Errorf("%s - cursor.Decode: %w", op, err)
		}
		cmds = append(cmds, cmd)
	}
	return cmds, nil
}

func (r *CommandMongoRepo) Update(ctx context.Context, systemName string, cmd *entity.Command) error {
	_, err := r.coll.UpdateOne(
		ctx,
		bson.M{"systemName": systemName},
		bson.M{"$set": cmd},
	)
	return err
}
