package repository

import (
	"context"
	stderrors "errors"
	"fmt"

	"github.com/adityakw90/service-access/internal/core/domain/errors"
	"github.com/adityakw90/service-access/internal/core/domain/model"
	"github.com/adityakw90/service-access/internal/core/domain/param"
	portrepository "github.com/adityakw90/service-access/internal/core/port/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// allowedOrderByPermission maps OrderBy string values to their typed enum for validation.
var allowedOrderByPermission = map[string]param.PermissionOrderBy{
	"id":          param.OrderByPermissionID,
	"uid":         param.OrderByPermissionUID,
	"resource":    param.OrderByPermissionResource,
	"action":      param.OrderByPermissionAction,
	"description": param.OrderByPermissionDescription,
	"created_at":  param.OrderByPermissionCreatedAt,
	"updated_at":  param.OrderByPermissionUpdatedAt,
}

type permissionRepository struct {
	db dbExecutor
}

// NewPermissionRepository creates a new permission repository with a connection pool
func NewPermissionRepository(db dbExecutor) portrepository.PermissionRepository {
	return &permissionRepository{db: db}
}

func (r *permissionRepository) Create(ctx context.Context, permission *model.Permission) error {
	// Validate UID is not empty
	if permission.UID == "" {
		return errors.ErrInvalidEntity
	}

	const sql = `
		INSERT INTO permission (uid, resource, action, description)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(ctx, sql, permission.UID, permission.Resource, permission.Action, permission.Description).
		Scan(&permission.ID, &permission.CreatedAt, &permission.UpdatedAt)
	if err != nil {
		// Check for unique constraint violation on (resource, action)
		var pgErr *pgconn.PgError
		if stderrors.As(err, &pgErr) && pgErr.ConstraintName == "uq_permission_resource_action" {
			return fmt.Errorf("permission with resource '%s' and action '%s' already exists", permission.Resource, permission.Action)
		}
		return fmt.Errorf("failed to create permission: %w", err)
	}

	return nil
}

func (r *permissionRepository) Update(ctx context.Context, permission *model.Permission) error {
	const sql = `
		UPDATE permission
		SET resource = $2, action = $3, description = $4, updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, sql, permission.ID, permission.Resource, permission.Action, permission.Description)
	if err != nil {
		// Check for unique constraint violation on (resource, action)
		var pgErr *pgconn.PgError
		if stderrors.As(err, &pgErr) && pgErr.ConstraintName == "uq_permission_resource_action" {
			return fmt.Errorf("permission with resource '%s' and action '%s' already exists", permission.Resource, permission.Action)
		}
		return fmt.Errorf("failed to update permission: %w", err)
	}

	return nil
}

func (r *permissionRepository) Delete(ctx context.Context, id int64) error {
	const sql = `DELETE FROM permission WHERE id = $1`

	result, err := r.db.Exec(ctx, sql, id)
	if err != nil {
		return fmt.Errorf("failed to delete permission: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("permission with id %d not found", id)
	}

	return nil
}

func (r *permissionRepository) GetByID(ctx context.Context, id int64) (*model.Permission, error) {
	const sql = `
		SELECT id, uid, resource, action, description, created_at, updated_at
		FROM permission
		WHERE id = $1
	`

	var p model.Permission
	err := r.db.QueryRow(ctx, sql, id).Scan(
		&p.ID, &p.UID, &p.Resource, &p.Action, &p.Description, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("permission with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}

	return &p, nil
}

func (r *permissionRepository) GetByUID(ctx context.Context, uid string) (*model.Permission, error) {
	const sql = `
		SELECT id, uid, resource, action, description, created_at, updated_at
		FROM permission
		WHERE uid = $1
	`

	var p model.Permission
	err := r.db.QueryRow(ctx, sql, uid).Scan(
		&p.ID, &p.UID, &p.Resource, &p.Action, &p.Description, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("permission with uid %s not found", uid)
		}
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}

	return &p, nil
}

func (r *permissionRepository) List(ctx context.Context, pagination *param.PaginationParam, filter *param.PermissionListFilterParam) (model.Permissions, error) {
	baseSQL := `
		SELECT id, uid, resource, action, description, created_at, updated_at
		FROM permission
		WHERE 1=1
	`
	args := []interface{}{}
	argIdx := 1

	// Apply filters
	if filter != nil {
		if len(filter.IDs) > 0 {
			baseSQL += fmt.Sprintf(" AND id = ANY($%d)", argIdx)
			args = append(args, filter.IDs)
			argIdx++
		}
		if len(filter.UIDs) > 0 {
			baseSQL += fmt.Sprintf(" AND uid = ANY($%d)", argIdx)
			args = append(args, filter.UIDs)
			argIdx++
		}
		if filter.Resource != nil {
			baseSQL += fmt.Sprintf(" AND resource = $%d", argIdx)
			args = append(args, *filter.Resource)
			argIdx++
		}
		if filter.Action != nil {
			baseSQL += fmt.Sprintf(" AND action = $%d", argIdx)
			args = append(args, *filter.Action)
			argIdx++
		}
		if filter.Query != nil {
			baseSQL += fmt.Sprintf(" AND (resource ILIKE $%d OR action ILIKE $%d OR description ILIKE $%d)", argIdx, argIdx+1, argIdx+2)
			args = append(args, "%"+*filter.Query+"%", "%"+*filter.Query+"%", "%"+*filter.Query+"%")
			argIdx += 3
		}
	}

	// Apply sorting
	orderByValue := r.validateOrderBy(pagination, "created_at")
	if pagination != nil && pagination.Sort != nil {
		orderByValue += " " + *pagination.Sort
	} else {
		orderByValue += " DESC"
	}
	baseSQL += " ORDER BY " + orderByValue

	// Apply pagination
	limit := 10
	offset := 0
	if pagination != nil {
		if pagination.Limit != nil {
			limit = *pagination.Limit
		}
		if pagination.Page != nil {
			offset = (*pagination.Page - 1) * limit
		}
	}
	baseSQL += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	// Execute query
	rows, err := r.db.Query(ctx, baseSQL, args...)
	if err != nil {
		return model.Permissions{}, fmt.Errorf("failed to list permissions: %w", err)
	}
	defer rows.Close()

	var items []model.Permission
	for rows.Next() {
		var p model.Permission
		err := rows.Scan(&p.ID, &p.UID, &p.Resource, &p.Action, &p.Description, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return model.Permissions{}, fmt.Errorf("failed to scan permission: %w", err)
		}
		items = append(items, p)
	}

	if rows.Err() != nil {
		return model.Permissions{}, fmt.Errorf("error iterating permissions: %w", rows.Err())
	}

	// Get total count
	countSQL := `SELECT COUNT(*) FROM permission WHERE 1=1`
	countArgs := []interface{}{}
	countArgIdx := 1
	if filter != nil {
		if len(filter.IDs) > 0 {
			countSQL += fmt.Sprintf(" AND id = ANY($%d)", countArgIdx)
			countArgs = append(countArgs, filter.IDs)
			countArgIdx++
		}
		if len(filter.UIDs) > 0 {
			countSQL += fmt.Sprintf(" AND uid = ANY($%d)", countArgIdx)
			countArgs = append(countArgs, filter.UIDs)
			countArgIdx++
		}
		if filter.Resource != nil {
			countSQL += fmt.Sprintf(" AND resource = $%d", countArgIdx)
			countArgs = append(countArgs, *filter.Resource)
			countArgIdx++
		}
		if filter.Action != nil {
			countSQL += fmt.Sprintf(" AND action = $%d", countArgIdx)
			countArgs = append(countArgs, *filter.Action)
			countArgIdx++
		}
		if filter.Query != nil {
			countSQL += fmt.Sprintf(" AND (resource ILIKE $%d OR action ILIKE $%d OR description ILIKE $%d)", countArgIdx, countArgIdx+1, countArgIdx+2)
			countArgs = append(countArgs, "%"+*filter.Query+"%", "%"+*filter.Query+"%", "%"+*filter.Query+"%")
			countArgIdx += 3
		}
	}

	var total int64
	err = r.db.QueryRow(ctx, countSQL, countArgs...).Scan(&total)
	if err != nil {
		return model.Permissions{}, fmt.Errorf("failed to get total count: %w", err)
	}

	return model.Permissions{
		Items: items,
		Meta: model.Meta{
			Total: total,
			Page:  offset/limit + 1,
			Limit: limit,
		},
	}, nil
}

// validateOrderBy validates the OrderBy value against allowed columns using O(1) map lookup.
func (r *permissionRepository) validateOrderBy(pagination *param.PaginationParam, defaultOrderBy string) string {
	if pagination != nil && pagination.OrderBy != nil {
		if _, ok := allowedOrderByPermission[*pagination.OrderBy]; ok {
			return *pagination.OrderBy
		}
	}
	return defaultOrderBy
}
