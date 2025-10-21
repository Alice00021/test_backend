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

type CommandRepo struct {
	*postgres.Postgres
}

func NewCommandRepo(pg *postgres.Postgres) *CommandRepo {
	return &CommandRepo{pg}
}

func (r *CommandRepo) Create(ctx context.Context, e *entity.Command) (*entity.Command, error) {
	op := "CommandRepo - Create"

	sql, args, err := r.Builder.
		Insert("commands").
		Columns(
			"name, system_name, reagent, average_time, volume_waste, volume_drive_fluid, "+
				"volume_container, default_address").
		Values(
			e.Name, e.SystemName, e.Reagent, e.AverageTime, e.VolumeWaste, e.VolumeDriveFluid,
			e.VolumeContainer, e.DefaultAddress).
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

func (r *CommandRepo) GetById(ctx context.Context, id int64) (*entity.Command, error) {
	op := "CommandRepo - GetById"

	sql, args, err := r.Builder.
		Select(
			"id", "created_at", "updated_at", "deleted_at", "name",
			"system_name", "reagent", "average_time", "volume_waste", "volume_drive_fluid",
			"volume_container", "default_address",
		).
		From("commands").
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

	command, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[entity.Command])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entity.ErrCommandNotFound
		}

		return nil, fmt.Errorf("%s - pgx.CollectOneRow: %w", op, err)
	}
	return command, nil
}

func (r *CommandRepo) Update(ctx context.Context, inp *entity.Command) error {
	op := "CommandRepo - Update"

	sqlBuilder := r.Builder.
		Update("commands").
		Set("name", inp.Name).
		Set("reagent", inp.Reagent).
		Set("average_time", inp.AverageTime).
		Set("volume_waste", inp.VolumeWaste).
		Set("volume_drive_fluid", inp.VolumeDriveFluid).
		Set("volume_container", inp.VolumeContainer).
		Set("default_address", inp.DefaultAddress).
		Where(squirrel.Eq{"id": inp.ID})

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

func (r *CommandRepo) GetBySystemNames(ctx context.Context) (map[string]entity.Command, error) {
	op := "CommandRepo - GetBySystemNames"

	sql, args, err := r.Builder.
		Select(
			"id", "created_at", "updated_at", "deleted_at", "name",
			"system_name", "reagent", "average_time", "volume_waste", "volume_drive_fluid",
			"volume_container", "default_address",
		).
		From("commands").
		Where("deleted_at IS NULL").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s - r.Builder: %w", op, err)
	}

	client := r.GetClient(ctx)
	rows, err := client.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("%s - client.Query: %w", op, err)
	}

	defer rows.Close()

	items := make(map[string]entity.Command)

	for rows.Next() {
		var e entity.Command

		err = rows.Scan(
			&e.ID, &e.CreatedAt, &e.UpdatedAt, &e.DeletedAt, &e.Name, &e.SystemName,
			&e.Reagent, &e.AverageTime, &e.VolumeWaste, &e.VolumeDriveFluid,
			&e.VolumeContainer, &e.DefaultAddress,
		)
		if err != nil {
			return nil, fmt.Errorf("%s - rows.Scan: %w", op, err)
		}

		items[e.SystemName] = e
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s - rows.Err: %w", op, err)
	}
	return items, nil
}
