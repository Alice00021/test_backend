package entity

type UserRole string

const (
	UserRoleAdmin  UserRole = "ADMIN"
	UserRoleClient UserRole = "CLIENT"
)

type User struct {
	Entity
	Name        string
	Surname     string
	Username    string
	Password    string
	Role        UserRole
	Email       string
	IsVerified  bool
	VerifyToken *string
	FilePath    *string
	Rating      float32
}

type CreateUserInput struct {
	Name     string   `json:"name"`
	Surname  string   `json:"surname"`
	Username string   `json:"username"`
	Password string   `json:"password"`
	Email    string   `json:"email"`
	Role     UserRole `json:"role"`
}

type UpdateUserInput struct {
	ID       int64   `json:"id"`
	Name     string  `json:"name"`
	Surname  string  `json:"surname"`
	Username string  `json:"username"`
	Rating   float32 `json:"rating"`
}
type ChangePasswordInput struct {
	ID              int64  `json:"id"`
	OldPassword     string `json:"oldPassword"`
	NewPassword     string `json:"newPassword"`
	ConfirmPassword string `json:"confirmPassword"`
}

func NewUser(name, surname, username, password, email string) *User {
	return &User{
		Name:       name,
		Surname:    surname,
		Username:   username,
		Password:   password,
		Email:      email,
		IsVerified: false,
	}
}

type FilterUserInput struct {
	IsVerified *bool `json:"is_verified"`
}
