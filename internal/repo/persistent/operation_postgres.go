package persistent

import (
	"context"
	"fmt"
	"github.com/Alice00021/test_common/pkg/postgres"
	"github.com/Masterminds/squirrel"
	"test_go/internal/entity"
)

type OperationRepo struct {
	*postgres.Postgres
}

func NewOperationRepo(pg *postgres.Postgres) *OperationRepo {
	return &OperationRepo{pg}
}

func (r *OperationRepo) Create(ctx context.Context, e *entity.Operation) (*entity.Operation, error) {
	op := "OperationRepo - Create"

	sql, args, err := r.Builder.
		Insert("operations").
		Columns(
			"name, description, average_time").
		Values(
			e.Name, e.Description, e.AverageTime).
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

func (r *OperationRepo) GetAll(ctx context.Context) (map[int64]*entity.Operation, error) {
	op := "OperationRepo - GetAll"

	sqlBuilder := r.Builder.
		Select(
			"op.id", "op.created_at", "op.updated_at", "op.deleted_at",
			"op.name", "op.description", "op.average_time",
			"opc.command_id", "c.name", "c.system_name", "c.default_address", "opc.address",
		).
		From("operations op").
		LeftJoin("operation_commands opc ON opc.operation_id = op.id").
		LeftJoin("commands c ON opc.command_id = c.id").
		Where("op.deleted_at IS NULL").
		OrderBy("op.id DESC")

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

	items := make(map[int64]*entity.Operation)

	for rows.Next() {
		var (
			e                  entity.Operation
			commandId          *int64
			commandName        *string
			commandSystemName  *string
			commandDefaultAddr *string
			commandAddress     *string
		)

		if err = rows.Scan(
			&e.ID, &e.CreatedAt, &e.UpdatedAt, &e.DeletedAt,
			&e.Name, &e.Description, &e.AverageTime,
			&commandId, &commandName, &commandSystemName, &commandDefaultAddr, &commandAddress,
		); err != nil {
			return nil, fmt.Errorf("%s - row.Scan: %w", op, err)
		}

		if _, ok := items[e.ID]; !ok {
			items[e.ID] = &entity.Operation{
				Entity: entity.Entity{
					ID:        e.ID,
					CreatedAt: e.CreatedAt,
					UpdatedAt: e.UpdatedAt,
					DeletedAt: e.DeletedAt,
				},
				Name:        e.Name,
				Description: e.Description,
				AverageTime: e.AverageTime,
				Commands:    make([]*entity.OperationCommand, 0),
			}
		}

		operation := items[e.ID]

		if commandId != nil {
			operation.Commands = append(operation.Commands, &entity.OperationCommand{
				ID: *commandId,
				Command: entity.Command{
					Entity: entity.Entity{
						ID: *commandId,
					},
					Name:           *commandName,
					SystemName:     *commandSystemName,
					DefaultAddress: entity.Address(*commandDefaultAddr),
				},
				Address: entity.Address(*commandAddress),
			})
		}
	}

	return items, nil
}

func (r *OperationRepo) GetById(ctx context.Context, id int64) (*entity.Operation, error) {
	op := "OperationRepo - GetById"

	sql, args, err := r.Builder.
		Select(
			"op.id", "op.created_at", "op.updated_at", "op.deleted_at",
			"op.name", "op.description", "op.average_time",
			"c.id", "c.system_name", "c.default_address",
		).
		From("operations op").
		LeftJoin("operation_commands opc ON opc.operation_id = op.id").
		LeftJoin("commands c ON opc.command_id = c.id").
		Where("op.deleted_at IS NULL").
		Where(squirrel.Eq{"op.id": id}).
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

	var e entity.Operation
	e.Commands = []*entity.OperationCommand{}

	for rows.Next() {
		var (
			commandId                  *int64
			systemName, defaultAddress *string
		)

		if err := rows.Scan(
			&e.ID, &e.CreatedAt, &e.UpdatedAt, &e.DeletedAt,
			&e.Name, &e.Description, &e.AverageTime,
			&commandId, &systemName, &defaultAddress,
		); err != nil {
			return nil, fmt.Errorf("%s - rows.Scan: %w", op, err)
		}

		if commandId != nil {
			addr := entity.Address("")
			if defaultAddress != nil {
				addr = entity.Address(*defaultAddress)
			}

			sysName := ""
			if systemName != nil {
				sysName = *systemName
			}

			e.Commands = append(e.Commands, &entity.OperationCommand{
				Command: entity.Command{
					Entity:         entity.Entity{ID: *commandId},
					SystemName:     sysName,
					DefaultAddress: addr,
				},
				Address: addr,
			})
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s - rows.Err: %w", op, err)
	}

	if len(e.Commands) == 0 {
		e.Commands = nil
	}

	return &e, nil
}

func (r *OperationRepo) Update(ctx context.Context, e *entity.Operation) error {
	op := "OperationRepo - Update"

	sqlBuilder := r.Builder.
		Update("operations").
		Set("name", e.Name).
		Set("description", e.Description).
		Set("average_time", e.AverageTime).
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

func (r *OperationRepo) DeleteById(ctx context.Context, id int64) error {
	op := "OperationRepo - Delete"

	builder := r.Builder.
		Update("operations").
		Set("deleted_at", "NOW()").
		Where("deleted_at IS NULL").
		Where(squirrel.Eq{"id": id})

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
