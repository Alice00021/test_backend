package persistent

import (
	"context"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"

	"github.com/Alice00021/test_common/pkg/postgres"
	"test_go/internal/entity"
)

type BookRepo struct {
	*postgres.Postgres
}

func NewBookRepo(pg *postgres.Postgres) *BookRepo {
	return &BookRepo{pg}
}

func (r *BookRepo) Create(ctx context.Context, e *entity.Book) (*entity.Book, error) {
	op := "BookRepo - Create"

	sql, args, err := r.Builder.
		Insert("books").
		Columns("title, author_id").
		Values(e.Title, e.AuthorId).
		Suffix(`RETURNING id`).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s - r.Builder: %w", op, err)
	}

	client := r.GetClient(ctx)

	var id int64
	err = client.QueryRow(ctx, sql, args...).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("%s - client.QueryRow: %w", op, err)
	}

	return r.GetById(ctx, id)
}

func (r *BookRepo) GetById(ctx context.Context, id int64) (*entity.Book, error) {
	op := "BookRepo - GetById"

	sql, args, err := r.Builder.
		Select(
			"b.id", "b.created_at", "b.updated_at", "b.deleted_at",
			"b.title", "b.author_id",
			"a.name", "a.gender",
		).
		From("books b").
		InnerJoin("authors a ON b.author_id = a.id").
		Where("b.deleted_at IS NULL").
		Where(squirrel.Eq{"b.id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s - r.Builder: %w", op, err)
	}

	client := r.GetClient(ctx)
	row := client.QueryRow(ctx, sql, args...)

	var e entity.Book
	if err = row.Scan(
		&e.ID, &e.CreatedAt, &e.UpdatedAt, &e.DeletedAt, &e.Title, &e.AuthorId,
		&e.Author.Name, &e.Author.Gender,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entity.ErrBookNotFound
		}

		return nil, fmt.Errorf("%s - row.Scan: %w", op, err)
	}

	e.Author.ID = e.AuthorId
	return &e, nil
}

func (r *BookRepo) Update(ctx context.Context, e *entity.Book) error {
	op := "BookRepo - Update"

	sqlBuilder := r.Builder.
		Update("books").
		Set("title", e.Title).
		Set("author_id", e.AuthorId).
		Where(squirrel.Eq{"id": e.ID})

	sql, args, err := sqlBuilder.ToSql()
	if err != nil {
		return fmt.Errorf("%s - r.Builder: %w", op, err)
	}

	client := r.GetClient(ctx)
	if _, err = client.Exec(ctx, sql, args...); err != nil {
		return fmt.Errorf("%s - client.Exec: %w", op, err)
	}

	return nil
}

func (r *BookRepo) DeleteById(ctx context.Context, id int64) error {
	op := "BookRepo - DeleteById"

	builder := r.Builder.
		Update("books").
		Set("deleted_at", "NOW()").
		Where(squirrel.Eq{"id": id}).
		Where("deleted_at IS NULL")

	sql, args, err := builder.ToSql()
	if err != nil {
		return fmt.Errorf("%s - r.Builder: %w", op, err)
	}

	client := r.GetClient(ctx)
	_, err = client.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("%s - client.Exec: %w", op, err)
	}

	return nil
}

func (r *BookRepo) GetAll(ctx context.Context) ([]*entity.Book, error) {
	op := "BookRepo - GetAll"

	sqlBuilder := r.Builder.
		Select(
			"b.id", "b.created_at", "b.updated_at", "b.deleted_at",
			"b.title", "b.author_id",
			"a.name", "a.gender",
		).
		From("books b").
		LeftJoin("authors a ON b.author_id = a.id").
		Where("b.deleted_at IS NULL")

	sqlBuilder = sqlBuilder.OrderBy("id DESC")

	sql, args, err := sqlBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s - r.Builder: %w", op, err)
	}

	client := r.GetClient(ctx)
	rows, err := client.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("%s - client.Query: %w", op, err)
	}
	defer rows.Close()

	items := make([]*entity.Book, 0, 64)

	for rows.Next() {
		e := entity.Book{}

		if err = rows.Scan(
			&e.ID, &e.CreatedAt, &e.UpdatedAt, &e.DeletedAt, &e.Title, &e.AuthorId,
			&e.Author.Name, &e.Author.Gender,
		); err != nil {
			return nil, fmt.Errorf("%s - row.Scan: %w", op, err)
		}

		e.Author.ID = e.AuthorId

		items = append(items, &e)
	}

	return items, nil
}
