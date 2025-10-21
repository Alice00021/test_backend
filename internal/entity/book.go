package entity

type Book struct {
	Entity
	Title    string
	AuthorId int64
	Author   Author
}

type CreateBookInput struct {
	Title    string `json:"name"`
	AuthorId int64  `json:"author_id"`
}

type UpdateBookInput struct {
	ID       int64  `json:"id"`
	Title    string `json:"name"`
	AuthorId int64  `json:"author_id"`
}

func NewBook(title string, authorId int64) *Book {
	return &Book{
		Title:    title,
		AuthorId: authorId,
	}
}
