package mongodb

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"test_go/internal/entity"
	"test_go/pkg/mongodb"
)

type OperationMongoRepo struct {
	coll *mongo.Collection
}

func NewOperationRepo(client *mongodb.Client) *OperationMongoRepo {
	return &OperationMongoRepo{
		coll: client.Collection("operations"),
	}
}

func (r *OperationMongoRepo) Create(ctx context.Context, operation *entity.Operation) (*entity.Operation, error) {
	op := "OperationMongoRepo - Create"

	_, err := r.coll.InsertOne(ctx, operation)
	if err != nil {
		return nil, fmt.Errorf("%s - r.coll.InsertOne: %w", op, err)
	}
	return operation, nil
}

func (r *OperationMongoRepo) Update(ctx context.Context, operation *entity.Operation) error {
	op := "OperationMongoRepo - Update"
	id := bson.M{"_id": operation.ID}

	update := bson.M{
		"$set": bson.M{
			"name":        operation.Name,
			"description": operation.Description,
			"averageTime": operation.AverageTime,
		},
	}
	_, err := r.coll.UpdateOne(ctx, id, update)
	if err != nil {
		return fmt.Errorf("%s - r.coll.UpdateOne: %w", op, err)
	}
	return err
}

func (r *OperationMongoRepo) DeleteById(ctx context.Context, id int64) error {
	operationId := bson.M{"_id": id}

	_, err := r.coll.DeleteOne(ctx, operationId)
	if err != nil {
		return fmt.Errorf("OperationMongoRepo - DeleteById - DeleteOne: %w", err)
	}

	return nil
}
