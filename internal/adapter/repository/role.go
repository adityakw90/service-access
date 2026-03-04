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

// allowedOrderByRole maps OrderBy string values to their typed enum for validation.
var allowedOrderByRole = map[string]param.RoleOrderBy{
	"id":          param.OrderByRoleID,
	"uid":         param.OrderByRoleUID,
	"group_id":    param.OrderByRoleGroupID,
	"name":        param.OrderByRoleName,
	"description": param.OrderByRoleDescription,
	"created_at":  param.OrderByRoleCreatedAt,
	"updated_at":  param.OrderByRoleUpdatedAt,
}

// allowedOrderByRolePermission maps OrderBy string values to their typed enum for validation.
var allowedOrderByRolePermission = map[string]param.RolePermissionOrderBy{
	"role_id":              param.OrderByRolePermissionRoleID,
	"group_permission_id": param.OrderByRolePermissionGroupPermissionID,
	"created_at":           param.OrderByRolePermissionCreatedAt,
}

type roleRepository struct {
	db dbExecutor
}

// NewRoleRepository creates a new role repository with a connection pool
func NewRoleRepository(db dbExecutor) portrepository.RoleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) Create(ctx context.Context, role *model.Role) error {
	// Validate UID is not empty
	if role.UID == "" {
		return errors.ErrInvalidEntity
	}

	const sql = `
		INSERT INTO role (uid, group_id, name, description)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(ctx, sql, role.UID, role.GroupID, role.Name, role.Description).
		Scan(&role.ID, &role.CreatedAt, &role.UpdatedAt)
	if err != nil {
		// Check for unique constraint violation on (group_id, name)
		var pgErr *pgconn.PgError
		if stderrors.As(err, &pgErr) && pgErr.ConstraintName == "uq_role_group_name" {
			return fmt.Errorf("role with name '%s' already exists in group", role.Name)
		}
		// Check for foreign key violation on group_id
		if stderrors.As(err, &pgErr) && pgErr.ConstraintName == "fk_role_group" {
			return fmt.Errorf("group with id %d not found", role.GroupID)
		}
		return fmt.Errorf("failed to create role: %w", err)
	}

	return nil
}

func (r *roleRepository) Update(ctx context.Context, role *model.Role) error {
	const sql = `
		UPDATE role
		SET name = $2, description = $3, updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, sql, role.ID, role.Name, role.Description)
	if err != nil {
		// Check for unique constraint violation on (group_id, name)
		var pgErr *pgconn.PgError
		if stderrors.As(err, &pgErr) && pgErr.ConstraintName == "uq_role_group_name" {
			return fmt.Errorf("role with name '%s' already exists in group", role.Name)
		}
		return fmt.Errorf("failed to update role: %w", err)
	}

	return nil
}

func (r *roleRepository) Delete(ctx context.Context, id int64) error {
	const sql = `DELETE FROM role WHERE id = $1`

	result, err := r.db.Exec(ctx, sql, id)
	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("role with id %d not found", id)
	}

	return nil
}

func (r *roleRepository) GetByID(ctx context.Context, id int64) (*model.Role, error) {
	const sql = `
		SELECT r.id, r.uid, r.group_id, g.uid as group_uid, r.name, r.description, r.created_at, r.updated_at
		FROM role r
		JOIN "group" g ON r.group_id = g.id
		WHERE r.id = $1
	`

	var role model.Role
	err := r.db.QueryRow(ctx, sql, id).Scan(
		&role.ID, &role.UID, &role.GroupID, &role.GroupUID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt,
	)
	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("role with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	return &role, nil
}

func (r *roleRepository) GetByUID(ctx context.Context, uid string) (*model.Role, error) {
	const sql = `
		SELECT r.id, r.uid, r.group_id, g.uid as group_uid, r.name, r.description, r.created_at, r.updated_at
		FROM role r
		JOIN "group" g ON r.group_id = g.id
		WHERE r.uid = $1
	`

	var role model.Role
	err := r.db.QueryRow(ctx, sql, uid).Scan(
		&role.ID, &role.UID, &role.GroupID, &role.GroupUID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt,
	)
	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("role with uid %s not found", uid)
		}
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	return &role, nil
}

func (r *roleRepository) List(ctx context.Context, pagination *param.PaginationParam, filter *param.RoleListFilterParam) (model.Roles, error) {
	baseSQL := `
		SELECT r.id, r.uid, r.group_id, g.uid as group_uid, r.name, r.description, r.created_at, r.updated_at
		FROM role r
		JOIN "group" g ON r.group_id = g.id
		WHERE 1=1
	`
	args := []interface{}{}
	argIdx := 1

	// Apply filters
	if filter != nil {
		if len(filter.IDs) > 0 {
			baseSQL += fmt.Sprintf(" AND r.id = ANY($%d)", argIdx)
			args = append(args, filter.IDs)
			argIdx++
		}
		if len(filter.UIDs) > 0 {
			baseSQL += fmt.Sprintf(" AND r.uid = ANY($%d)", argIdx)
			args = append(args, filter.UIDs)
			argIdx++
		}
		if filter.GroupID != nil {
			baseSQL += fmt.Sprintf(" AND r.group_id = $%d", argIdx)
			args = append(args, *filter.GroupID)
			argIdx++
		}
		if filter.GroupUID != nil {
			baseSQL += fmt.Sprintf(" AND g.uid = $%d", argIdx)
			args = append(args, *filter.GroupUID)
			argIdx++
		}
		if filter.Name != nil {
			baseSQL += fmt.Sprintf(" AND r.name = $%d", argIdx)
			args = append(args, *filter.Name)
			argIdx++
		}
		if filter.Query != nil {
			baseSQL += fmt.Sprintf(" AND (r.name ILIKE $%d OR r.description ILIKE $%d)", argIdx, argIdx+1)
			args = append(args, "%"+*filter.Query+"%", "%"+*filter.Query+"%")
			argIdx += 2
		}
	}

	// Apply sorting
	orderByValue := r.validateOrderBy(pagination, "created_at")
	orderBy := "r." + orderByValue
	if pagination != nil && pagination.Sort != nil {
		orderBy += " " + *pagination.Sort
	} else {
		orderBy += " DESC"
	}
	baseSQL += " ORDER BY " + orderBy

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
		return model.Roles{}, fmt.Errorf("failed to list roles: %w", err)
	}
	defer rows.Close()

	var items []model.Role
	for rows.Next() {
		var role model.Role
		err := rows.Scan(&role.ID, &role.UID, &role.GroupID, &role.GroupUID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt)
		if err != nil {
			return model.Roles{}, fmt.Errorf("failed to scan role: %w", err)
		}
		items = append(items, role)
	}

	if rows.Err() != nil {
		return model.Roles{}, fmt.Errorf("error iterating roles: %w", rows.Err())
	}

	// Get total count
	countSQL := `SELECT COUNT(*) FROM role r JOIN "group" g ON r.group_id = g.id WHERE 1=1`
	countArgs := []interface{}{}
	countArgIdx := 1
	if filter != nil {
		if len(filter.IDs) > 0 {
			countSQL += fmt.Sprintf(" AND r.id = ANY($%d)", countArgIdx)
			countArgs = append(countArgs, filter.IDs)
			countArgIdx++
		}
		if len(filter.UIDs) > 0 {
			countSQL += fmt.Sprintf(" AND r.uid = ANY($%d)", countArgIdx)
			countArgs = append(countArgs, filter.UIDs)
			countArgIdx++
		}
		if filter.GroupID != nil {
			countSQL += fmt.Sprintf(" AND r.group_id = $%d", countArgIdx)
			countArgs = append(countArgs, *filter.GroupID)
			countArgIdx++
		}
		if filter.GroupUID != nil {
			countSQL += fmt.Sprintf(" AND g.uid = $%d", countArgIdx)
			countArgs = append(countArgs, *filter.GroupUID)
			countArgIdx++
		}
		if filter.Name != nil {
			countSQL += fmt.Sprintf(" AND r.name = $%d", countArgIdx)
			countArgs = append(countArgs, *filter.Name)
			countArgIdx++
		}
		if filter.Query != nil {
			countSQL += fmt.Sprintf(" AND (r.name ILIKE $%d OR r.description ILIKE $%d)", countArgIdx, countArgIdx+1)
			countArgs = append(countArgs, "%"+*filter.Query+"%", "%"+*filter.Query+"%")
			countArgIdx += 2
		}
	}

	var total int64
	err = r.db.QueryRow(ctx, countSQL, countArgs...).Scan(&total)
	if err != nil {
		return model.Roles{}, fmt.Errorf("failed to get total count: %w", err)
	}

	return model.Roles{
		Items: items,
		Meta: model.Meta{
			Total: total,
			Page:  offset/limit + 1,
			Limit: limit,
		},
	}, nil
}

func (r *roleRepository) ListPermission(ctx context.Context, roleID int64, pagination *param.PaginationParam, filter *param.RolePermissionListFilterParam) (model.RolePermissions, error) {
	baseSQL := `
		SELECT rp.role_id, r.uid as role_uid,
		       rp.group_permission_id, gp.uid as group_permission_uid,
		       p.uid as permission_uid, p.resource, p.action, p.description,
		       rp.created_at
		FROM role_permission rp
		JOIN role r ON rp.role_id = r.id
		JOIN group_permission gp ON rp.group_permission_id = gp.id
		JOIN permission p ON gp.permission_id = p.id
		WHERE rp.role_id = $1
	`
	args := []interface{}{roleID}
	argIdx := 2

	// Apply filters
	if filter != nil {
		if len(filter.IDs) > 0 {
			baseSQL += fmt.Sprintf(" AND rp.group_permission_id = ANY($%d)", argIdx)
			args = append(args, filter.IDs)
			argIdx++
		}
		if len(filter.UIDs) > 0 {
			baseSQL += fmt.Sprintf(" AND gp.uid = ANY($%d)", argIdx)
			args = append(args, filter.UIDs)
			argIdx++
		}
		if len(filter.PermissionIDs) > 0 {
			baseSQL += fmt.Sprintf(" AND gp.permission_id = ANY($%d)", argIdx)
			args = append(args, filter.PermissionIDs)
			argIdx++
		}
		if len(filter.PermissionUIDs) > 0 {
			baseSQL += fmt.Sprintf(" AND p.uid = ANY($%d)", argIdx)
			args = append(args, filter.PermissionUIDs)
			argIdx++
		}
		if filter.Resource != nil {
			baseSQL += fmt.Sprintf(" AND p.resource = $%d", argIdx)
			args = append(args, *filter.Resource)
			argIdx++
		}
		if filter.Action != nil {
			baseSQL += fmt.Sprintf(" AND p.action = $%d", argIdx)
			args = append(args, *filter.Action)
			argIdx++
		}
		if filter.Query != nil {
			baseSQL += fmt.Sprintf(" AND (p.resource ILIKE $%d OR p.action ILIKE $%d OR p.description ILIKE $%d)", argIdx, argIdx+1, argIdx+2)
			args = append(args, "%"+*filter.Query+"%", "%"+*filter.Query+"%", "%"+*filter.Query+"%")
			argIdx += 3
		}
	}

	// Apply sorting
	orderByValue := r.validateOrderByRolePermission(pagination, "created_at")
	orderBy := "rp." + orderByValue
	if pagination != nil && pagination.Sort != nil {
		orderBy += " " + *pagination.Sort
	} else {
		orderBy += " DESC"
	}
	baseSQL += " ORDER BY " + orderBy

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
		return model.RolePermissions{}, fmt.Errorf("failed to list role permissions: %w", err)
	}
	defer rows.Close()

	var items []model.RolePermission
	for rows.Next() {
		var rp model.RolePermission
		err := rows.Scan(
			&rp.RoleID, &rp.RoleUID,
			&rp.GroupPermissionID, &rp.GroupPermissionUID,
			&rp.PermissionUID, &rp.PermissionResource, &rp.PermissionAction, &rp.PermissionDescription,
			&rp.CreatedAt,
		)
		if err != nil {
			return model.RolePermissions{}, fmt.Errorf("failed to scan role permission: %w", err)
		}
		items = append(items, rp)
	}

	if rows.Err() != nil {
		return model.RolePermissions{}, fmt.Errorf("error iterating role permissions: %w", rows.Err())
	}

	// Get total count
	countSQL := `SELECT COUNT(*) FROM role_permission rp JOIN role r ON rp.role_id = r.id JOIN group_permission gp ON rp.group_permission_id = gp.id JOIN permission p ON gp.permission_id = p.id WHERE rp.role_id = $1`
	countArgs := []interface{}{roleID}
	countArgIdx := 2
	if filter != nil {
		if len(filter.IDs) > 0 {
			countSQL += fmt.Sprintf(" AND rp.group_permission_id = ANY($%d)", countArgIdx)
			countArgs = append(countArgs, filter.IDs)
			countArgIdx++
		}
		if len(filter.UIDs) > 0 {
			countSQL += fmt.Sprintf(" AND gp.uid = ANY($%d)", countArgIdx)
			countArgs = append(countArgs, filter.UIDs)
			countArgIdx++
		}
		if len(filter.PermissionIDs) > 0 {
			countSQL += fmt.Sprintf(" AND gp.permission_id = ANY($%d)", countArgIdx)
			countArgs = append(countArgs, filter.PermissionIDs)
			countArgIdx++
		}
		if len(filter.PermissionUIDs) > 0 {
			countSQL += fmt.Sprintf(" AND p.uid = ANY($%d)", countArgIdx)
			countArgs = append(countArgs, filter.PermissionUIDs)
			countArgIdx++
		}
		if filter.Resource != nil {
			countSQL += fmt.Sprintf(" AND p.resource = $%d", countArgIdx)
			countArgs = append(countArgs, *filter.Resource)
			countArgIdx++
		}
		if filter.Action != nil {
			countSQL += fmt.Sprintf(" AND p.action = $%d", countArgIdx)
			countArgs = append(countArgs, *filter.Action)
			countArgIdx++
		}
		if filter.Query != nil {
			countSQL += fmt.Sprintf(" AND (p.resource ILIKE $%d OR p.action ILIKE $%d OR p.description ILIKE $%d)", countArgIdx, countArgIdx+1, countArgIdx+2)
			countArgs = append(countArgs, "%"+*filter.Query+"%", "%"+*filter.Query+"%", "%"+*filter.Query+"%")
			countArgIdx += 3
		}
	}

	var total int64
	err = r.db.QueryRow(ctx, countSQL, countArgs...).Scan(&total)
	if err != nil {
		return model.RolePermissions{}, fmt.Errorf("failed to get total count: %w", err)
	}

	return model.RolePermissions{
		Items: items,
		Meta: model.Meta{
			Total: total,
			Page:  offset/limit + 1,
			Limit: limit,
		},
	}, nil
}

func (r *roleRepository) AddPermission(ctx context.Context, roleID int64, groupPermissionID int64) error {
	const sql = `
		INSERT INTO role_permission (role_id, group_permission_id)
		VALUES ($1, $2)
		ON CONFLICT (role_id, group_permission_id) DO NOTHING
	`

	_, err := r.db.Exec(ctx, sql, roleID, groupPermissionID)
	if err != nil {
		// Check for foreign key violations
		var pgErr *pgconn.PgError
		if stderrors.As(err, &pgErr) {
			if pgErr.ConstraintName == "fk_role_permission_role" {
				return fmt.Errorf("role with id %d not found", roleID)
			}
			if pgErr.ConstraintName == "fk_role_permission_group_permission" {
				return fmt.Errorf("group permission with id %d not found", groupPermissionID)
			}
		}
		return fmt.Errorf("failed to add permission to role: %w", err)
	}

	return nil
}

// validateOrderBy validates the OrderBy value against allowed Role columns using O(1) map lookup.
func (r *roleRepository) validateOrderBy(pagination *param.PaginationParam, defaultOrderBy string) string {
	if pagination != nil && pagination.OrderBy != nil {
		if _, ok := allowedOrderByRole[*pagination.OrderBy]; ok {
			return *pagination.OrderBy
		}
	}
	return defaultOrderBy
}

// validateOrderByRolePermission validates the OrderBy value against allowed RolePermission columns using O(1) map lookup.
func (r *roleRepository) validateOrderByRolePermission(pagination *param.PaginationParam, defaultOrderBy string) string {
	if pagination != nil && pagination.OrderBy != nil {
		if _, ok := allowedOrderByRolePermission[*pagination.OrderBy]; ok {
			return *pagination.OrderBy
		}
	}
	return defaultOrderBy
}

func (r *roleRepository) RemovePermission(ctx context.Context, roleID int64, groupPermissionID int64) error {
	const sql = `DELETE FROM role_permission WHERE role_id = $1 AND group_permission_id = $2`

	result, err := r.db.Exec(ctx, sql, roleID, groupPermissionID)
	if err != nil {
		return fmt.Errorf("failed to remove permission from role: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("group permission %d not found in role %d", groupPermissionID, roleID)
	}

	return nil
}

func (r *roleRepository) ReplacePermission(ctx context.Context, roleID int64, groupPermissionIDs []int64) error {
	tx, ok := r.db.(pgx.Tx)
	if !ok {
		return fmt.Errorf("ReplacePermission requires a transaction")
	}

	// First, delete all existing permissions for the role
	const deleteSQL = `DELETE FROM role_permission WHERE role_id = $1`
	_, err := tx.Exec(ctx, deleteSQL, roleID)
	if err != nil {
		return fmt.Errorf("failed to remove existing permissions: %w", err)
	}

	// Then, add the new permissions
	const insertSQL = `
		INSERT INTO role_permission (role_id, group_permission_id)
		VALUES ($1, $2)
	`
	for _, groupPermissionID := range groupPermissionIDs {
		_, err := tx.Exec(ctx, insertSQL, roleID, groupPermissionID)
		if err != nil {
			var pgErr *pgconn.PgError
			if stderrors.As(err, &pgErr) {
				if pgErr.ConstraintName == "fk_role_permission_role" {
					return fmt.Errorf("role with id %d not found", roleID)
				}
				if pgErr.ConstraintName == "fk_role_permission_group_permission" {
					return fmt.Errorf("group permission with id %d not found", groupPermissionID)
				}
			}
			return fmt.Errorf("failed to add permission %d: %w", groupPermissionID, err)
		}
	}

	return nil
}
