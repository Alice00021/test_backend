package mongodb

import (
	"context"
	"fmt"
	"github.com/Alice00021/test_common/pkg/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"test_go/internal/entity"
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
	op := "CommandMongoRepo - Create"

	res, err := r.coll.InsertOne(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("%s - r.coll.InsertOne: %w", op, err)
	}
	return r.GetById(ctx, res.InsertedID.(primitive.ObjectID))
}

func (r *CommandMongoRepo) GetById(ctx context.Context, id primitive.ObjectID) (*entity.Command, error) {
	var cmd entity.Command

	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&cmd)
	if err != nil {
		return nil, fmt.Errorf("CommandMongoRepo - GetById: %w", err)
	}

	return &cmd, nil
}

func (r *CommandMongoRepo) GetBySystemName(ctx context.Context) (map[string]entity.Command, error) {
	op := "CommandMongoRepo - GetBySystemName"

	cursor, err := r.coll.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("%s - r.coll.Find: %w", op, err)
	}
	defer cursor.Close(ctx)

	items := make(map[string]entity.Command)

	for cursor.Next(ctx) {
		var cmd entity.Command
		if err := cursor.Decode(&cmd); err != nil {
			return nil, fmt.Errorf("%s - cursor.Decode: %w", op, err)
		}
		items[cmd.SystemName] = cmd
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("%s - cursor.Err: %w", op, err)
	}

	return items, nil
}

func (r *CommandMongoRepo) GetAll(ctx context.Context) ([]entity.Command, error) {
	op := "CommandMongoRepo - GetAll"

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
