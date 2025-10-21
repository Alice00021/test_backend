package request

import "test_go/internal/entity"

type ChangePasswordRequest struct {
	OldPassword     string `json:"oldPassword" validate:"required"`
	NewPassword     string `json:"newPassword" validate:"required"`
	ConfirmPassword string `json:"confirmPassword" validate:"required"`
}

func (req *ChangePasswordRequest) ToEntity() entity.ChangePasswordInput {
	return entity.ChangePasswordInput{
		OldPassword:     req.OldPassword,
		NewPassword:     req.NewPassword,
		ConfirmPassword: req.ConfirmPassword,
	}
}

type UpdateRatingRequest struct {
	Rating float32 `json:"rating" validate:"required" min:"0" max:"100"`
}
