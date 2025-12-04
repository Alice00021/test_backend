package request

import "test_go/internal/entity"

type CreateUserRequest struct {
	Name     string `json:"name" validate:"required"`
	Surname  string `json:"surname" validate:"required"`
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
	Email    string `json:"email" validate:"required"`
}

func (req *CreateUserRequest) ToEntity() entity.CreateUserInput {
	return entity.CreateUserInput{
		Name:     req.Name,
		Surname:  req.Surname,
		Username: req.Username,
		Password: req.Password,
		Email:    req.Email,
	}
}

type AuthenticateRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type VerifyEmailRequest struct {
	Token string `form:"token" validate:"required"`
}

type ValidateTokenRequest struct {
	AccessToken string `json:"accessToken" binding:"required"`
}
