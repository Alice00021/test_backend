package mongodb

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

func (r *OperationMongoRepo) Create(ctx context.Context, operation *entity.OperationMongo) (*entity.OperationMongo, error) {
	op := "OperationMongoRepo - Create"

	res, err := r.coll.InsertOne(ctx, operation)
	if err != nil {
		return nil, fmt.Errorf("%s - r.coll.InsertOne: %w", op, err)
	}
	return r.GetById(ctx, res.InsertedID.(primitive.ObjectID))
}

func (r *OperationMongoRepo) Update(ctx context.Context, operation *entity.OperationMongo) error {
	op := "OperationMongoRepo - Update"
	id := bson.M{"_id": operation.Id}

	update := bson.M{
		"$set": bson.M{
			"name":        operation.Name,
			"description": operation.Description,
			"averageTime": operation.AverageTime,
			"commands":    operation.Commands,
		},
	}
	_, err := r.coll.UpdateOne(ctx, id, update)
	if err != nil {
		return fmt.Errorf("%s - r.coll.UpdateOne: %w", op, err)
	}
	return err
}

func (r *OperationMongoRepo) GetById(ctx context.Context, id primitive.ObjectID) (*entity.OperationMongo, error) {
	var op entity.OperationMongo

	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&op)
	if err != nil {
		return nil, fmt.Errorf("OperationMongoRepo - GetById: %w", err)
	}

	return &op, nil
}

func (r *OperationMongoRepo) GetAll(ctx context.Context) ([]*entity.OperationMongo, error) {
	op := "OperationMongoRepo - GetAll"

	cursor, err := r.coll.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("%s - r.coll.Find: %w", op, err)
	}
	defer cursor.Close(ctx)

	var operations []*entity.OperationMongo
	for cursor.Next(ctx) {
		var oper entity.OperationMongo
		if err := cursor.Decode(&oper); err != nil {
			return nil, fmt.Errorf("%s - cursor.Decode: %w", op, err)
		}
		operations = append(operations, &oper)
	}
	return operations, nil
}

func (r *OperationMongoRepo) DeleteById(ctx context.Context, id primitive.ObjectID) error {
	operationId := bson.M{"_id": id}

	_, err := r.coll.DeleteOne(ctx, operationId)
	if err != nil {
		return fmt.Errorf("OperationMongoRepo - DeleteById - DeleteOne: %w", err)
	}

	return nil
}
