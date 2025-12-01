package entity

import "go.mongodb.org/mongo-driver/bson/primitive"

type OperationMongo struct {
	Id          primitive.ObjectID
	Name        string
	Description string
	AverageTime int64
	Commands    []*OperationCommand
}

type UpdateOperationInputMongo struct {
	ID          primitive.ObjectID `json:"id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Commands    []*CommandInput    `json:"commands"`
}
