package request

import "test_go/internal/entity"

type CreateBookRequest struct {
	Title    string `json:"title" validate:"required"`
	AuthorId int64  `json:"authorId" validate:"required"`
}

func (req *CreateBookRequest) ToEntity() entity.CreateBookInput {
	return entity.CreateBookInput{
		Title:    req.Title,
		AuthorId: req.AuthorId,
	}
}

type UpdateBookRequest struct {
	CreateBookRequest
}

func (req *UpdateBookRequest) ToEntity() entity.UpdateBookInput {
	return entity.UpdateBookInput{
		Title:    req.Title,
		AuthorId: req.AuthorId,
	}
}
