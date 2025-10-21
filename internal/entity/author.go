package entity

type Author struct {
	Entity
	Name   string
	Gender bool
}

type CreateAuthorInput struct {
	Name   string `json:"name"`
	Gender bool   `json:"gender"`
}

type UpdateAuthorInput struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func NewAuthor(name string, gender bool) *Author {
	return &Author{
		Name:   name,
		Gender: gender,
	}
}
