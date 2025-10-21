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

type UserRepo struct {
	*postgres.Postgres
}

func NewUserRepo(pg *postgres.Postgres) *UserRepo {
	return &UserRepo{pg}
}

func (r *UserRepo) Create(ctx context.Context, e *entity.User) (*entity.User, error) {
	op := "UserRepo - Create"

	sql, args, err := r.Builder.
		Insert("users").
		Columns(
			"name, surname, username, password, file_path, email, "+
				"verify_token, is_verified, rating", "role").
		Values(
			e.Name, e.Surname, e.Username, e.Password, e.FilePath, e.Email,
			e.VerifyToken, e.IsVerified, e.Rating, e.Role).
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

func (r *UserRepo) GetById(ctx context.Context, id int64) (*entity.User, error) {
	op := "UserRepo - GetById"

	sql, args, err := r.Builder.
		Select(
			"id", "created_at", "updated_at", "deleted_at", "name",
			"surname", "username", "password", "file_path", "email",
			"verify_token", "is_verified", "rating", "role",
		).
		From("users").
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

	user, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[entity.User])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entity.ErrUserNotFound
		}

		return nil, fmt.Errorf("%s - pgx.CollectOneRow: %w", op, err)
	}
	return user, nil
}

func (r *UserRepo) Update(ctx context.Context, e *entity.User) error {
	op := "UserRepo - Update"

	sqlBuilder := r.Builder.
		Update("users").
		Set("name", e.Name).
		Set("surname", e.Surname).
		Set("username", e.Username).
		Set("rating", e.Rating).
		Set("is_verified", e.IsVerified).
		Set("verify_token", e.VerifyToken).
		Set("password", e.Password).
		Set("file_path", e.FilePath).
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

func (r *UserRepo) GetByUserName(ctx context.Context, username string) (*entity.User, error) {
	op := "UserRepo - GetById"

	sql, args, err := r.Builder.
		Select(
			"id", "created_at", "updated_at", "deleted_at", "name",
			"surname", "username", "password", "file_path", "email",
			"verify_token", "is_verified", "rating", "role",
		).
		From("users").
		Where("deleted_at IS NULL").
		Where(squirrel.Eq{"username": username}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s - r.Builder: %w", op, err)
	}

	client := r.GetClient(ctx)
	rows, err := client.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("%s - client.Query: %w", op, err)
	}

	user, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[entity.User])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entity.ErrUserNotFound
		}

		return nil, fmt.Errorf("%s - pgx.CollectOneRow: %w", op, err)
	}
	return user, nil
}

func (r *UserRepo) GetAll(ctx context.Context, filter entity.FilterUserInput) ([]*entity.User, error) {
	op := "UserRepo - GetAll"

	sqlBuilder := r.Builder.
		Select(
			"id", "created_at", "updated_at", "deleted_at", "name",
			"surname", "username", "password", "file_path", "email",
			"verify_token", "is_verified", "rating", "role",
		).
		From("users").
		Where("deleted_at IS NULL")

	if filter.IsVerified != nil {
		sqlBuilder = sqlBuilder.Where(squirrel.Eq{"is_verified": *filter.IsVerified})
	}

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

	items := make([]*entity.User, 0, 64)

	for rows.Next() {
		e := entity.User{}

		if err = rows.Scan(
			&e.ID, &e.CreatedAt, &e.UpdatedAt, &e.DeletedAt, &e.Name,
			&e.Surname, &e.Username, &e.Password, &e.FilePath, &e.Email,
			&e.VerifyToken, &e.IsVerified, &e.Rating, &e.Role,
		); err != nil {
			return nil, fmt.Errorf("%s - row.Scan: %w", op, err)
		}

		items = append(items, &e)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s - rows error: %w", op, err)
	}

	return items, nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	op := "UserRepo - GetByEmail"

	sql, args, err := r.Builder.
		Select(
			"id", "created_at", "updated_at", "deleted_at", "name",
			"surname", "username", "password", "file_path", "email",
			"verify_token", "is_verified", "rating", "role",
		).
		From("users").
		Where("deleted_at IS NULL").
		Where(squirrel.Eq{"email": email}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s - r.Builder: %w", op, err)
	}

	client := r.GetClient(ctx)
	rows, err := client.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("%s - client.Query: %w", op, err)
	}

	user, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[entity.User])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entity.ErrUserNotFound
		}

		return nil, fmt.Errorf("%s - pgx.CollectOneRow: %w", op, err)
	}
	return user, nil
}

func (r *UserRepo) GetByVerifyToken(ctx context.Context, verifyToken string) (*entity.User, error) {
	op := "UserRepo - GetByVerifyToken"

	sql, args, err := r.Builder.
		Select(
			"id", "created_at", "updated_at", "deleted_at", "name",
			"surname", "username", "password", "file_path", "email",
			"verify_token", "is_verified", "rating", "role",
		).
		From("users").
		Where("deleted_at IS NULL").
		Where(squirrel.Eq{"verify_token": verifyToken}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s - r.Builder: %w", op, err)
	}

	client := r.GetClient(ctx)
	rows, err := client.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("%s - client.Query: %w", op, err)
	}

	user, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[entity.User])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entity.ErrUserNotFound
		}

		return nil, fmt.Errorf("%s - pgx.CollectOneRow: %w", op, err)
	}
	return user, nil
}
