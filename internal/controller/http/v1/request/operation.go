package request

import "test_go/internal/entity"

type CreateOperationRequest struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Commands    []*entity.CommandInput `json:"commands"`
}

func (req *CreateOperationRequest) ToEntity() entity.CreateOperationInput {
	return entity.CreateOperationInput{
		Name:        req.Name,
		Description: req.Description,
		Commands:    req.Commands,
	}
}

type UpdateOperationRequest struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Commands    []*entity.CommandInput `json:"commands"`
}

func (req *UpdateOperationRequest) ToEntity() entity.UpdateOperationInputMongo {
	return entity.UpdateOperationInputMongo{
		Name:        req.Name,
		Description: req.Description,
		Commands:    req.Commands,
	}
}
