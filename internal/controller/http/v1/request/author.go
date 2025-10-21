package request

import "test_go/internal/entity"

type CreateAuthorRequest struct {
	Name   string `json:"name" validate:"required"`
	Gender bool   `json:"gender" validate:"required"`
}

func (req *CreateAuthorRequest) ToEntity() entity.CreateAuthorInput {
	return entity.CreateAuthorInput{
		Name:   req.Name,
		Gender: req.Gender,
	}
}

type UpdateAuthorRequest struct {
	Name string `json:"name" validate:"required"`
}

func (req *UpdateAuthorRequest) ToEntity() entity.UpdateAuthorInput {
	return entity.UpdateAuthorInput{
		Name: req.Name,
	}
}
