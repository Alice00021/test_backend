package persistent

import (
	"context"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"

	"test_go/internal/entity"
	"test_go/pkg/postgres"
)

type AuthorRepo struct {
	*postgres.Postgres
}

func NewAuthorRepo(pg *postgres.Postgres) *AuthorRepo {
	return &AuthorRepo{pg}
}

func (r *AuthorRepo) Create(ctx context.Context, e *entity.Author) (*entity.Author, error) {
	op := "AuthorRepo - Create"

	sql, args, err := r.Builder.
		Insert("authors").
		Columns("name, gender").
		Values(e.Name, e.Gender).
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

func (r *AuthorRepo) GetById(ctx context.Context, id int64) (*entity.Author, error) {
	op := "AuthorRepo - GetById"

	sql, args, err := r.Builder.
		Select(
			"id", "created_at", "updated_at", "deleted_at", "name",
			"gender",
		).
		From("authors").
		Where("deleted_at IS NULL").
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s - r.Builder: %w", op, err)
	}

	client := r.GetClient(ctx)
	rows, err := client.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("%s - client.Query: %w", op, err)
	}

	author, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[entity.Author])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entity.ErrAuthorNotFound
		}

		return nil, fmt.Errorf("%s - pgx.CollectOneRow: %w", op, err)
	}
	return author, nil
}

func (r *AuthorRepo) Update(ctx context.Context, e *entity.Author) error {
	op := "AuthorRepo - Update"

	sqlBuilder := r.Builder.
		Update("authors").
		Set("name", e.Name).
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

func (r *AuthorRepo) DeleteById(ctx context.Context, id int64) error {
	op := "AuthorRepo - DeleteById"

	builder := r.Builder.
		Update("authors").
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

func (r *AuthorRepo) GetAll(ctx context.Context) ([]*entity.Author, error) {
	op := "AuthorRepo - GetAll"

	sqlBuilder := r.Builder.
		Select(
			"id", "created_at", "updated_at", "deleted_at", "name",
			"gender",
		).
		From("authors").
		Where("deleted_at IS NULL")

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

	items := make([]*entity.Author, 0, 64)

	for rows.Next() {
		e := entity.Author{}

		if err = rows.Scan(
			&e.ID, &e.CreatedAt, &e.UpdatedAt, &e.DeletedAt, &e.Name, &e.Gender,
		); err != nil {
			return nil, fmt.Errorf("%s - row.Scan: %w", op, err)
		}

		items = append(items, &e)
	}

	return items, nil
}
