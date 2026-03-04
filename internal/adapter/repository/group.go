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

// allowedOrderByGroup maps OrderBy string values to their typed enum for validation.
var allowedOrderByGroup = map[string]param.GroupOrderBy{
	"id":          param.OrderByGroupID,
	"uid":         param.OrderByGroupUID,
	"name":        param.OrderByGroupName,
	"description": param.OrderByGroupDescription,
	"created_at":  param.OrderByGroupCreatedAt,
	"updated_at":  param.OrderByGroupUpdatedAt,
}

// allowedOrderByGroupPermission maps OrderBy string values to their typed enum for validation.
var allowedOrderByGroupPermission = map[string]param.GroupPermissionOrderBy{
	"group_id":      param.OrderByGroupPermissionGroupID,
	"permission_id": param.OrderByGroupPermissionPermissionID,
	"created_at":    param.OrderByGroupPermissionCreatedAt,
}

type groupRepository struct {
	db dbExecutor
}

// NewGroupRepository creates a new group repository with a connection pool
func NewGroupRepository(db dbExecutor) portrepository.GroupRepository {
	return &groupRepository{db: db}
}

func (r *groupRepository) Create(ctx context.Context, group *model.Group) error {
	// Validate UID is not empty
	if group.UID == "" {
		return errors.ErrInvalidEntity
	}

	const sql = `
		INSERT INTO "group" (uid, name, description)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRow(ctx, sql, group.UID, group.Name, group.Description).Scan(&group.ID, &group.CreatedAt, &group.UpdatedAt)
	if err != nil {
		// Check for unique constraint violation on name
		var pgErr *pgconn.PgError
		if stderrors.As(err, &pgErr) && pgErr.ConstraintName == "uq_group_name" {
			return fmt.Errorf("group with name '%s' already exists", group.Name)
		}
		return fmt.Errorf("failed to create group: %w", err)
	}

	return nil
}

func (r *groupRepository) Update(ctx context.Context, group *model.Group) error {
	const sql = `
		UPDATE "group"
		SET name = $2, description = $3, updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, sql, group.ID, group.Name, group.Description)
	if err != nil {
		// Check for unique constraint violation on name
		var pgErr *pgconn.PgError
		if stderrors.As(err, &pgErr) && pgErr.ConstraintName == "uq_group_name" {
			return fmt.Errorf("group with name '%s' already exists", group.Name)
		}
		return fmt.Errorf("failed to update group: %w", err)
	}

	return nil
}

func (r *groupRepository) Delete(ctx context.Context, id int64) error {
	const sql = `DELETE FROM "group" WHERE id = $1`

	result, err := r.db.Exec(ctx, sql, id)
	if err != nil {
		return fmt.Errorf("failed to delete group: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("group with id %d not found", id)
	}

	return nil
}

func (r *groupRepository) GetByID(ctx context.Context, id int64) (*model.Group, error) {
	const sql = `
		SELECT id, uid, name, description, created_at, updated_at
		FROM "group"
		WHERE id = $1
	`

	var g model.Group
	err := r.db.QueryRow(ctx, sql, id).Scan(
		&g.ID, &g.UID, &g.Name, &g.Description, &g.CreatedAt, &g.UpdatedAt,
	)
	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("group with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to get group: %w", err)
	}

	return &g, nil
}

func (r *groupRepository) GetByUID(ctx context.Context, uid string) (*model.Group, error) {
	const sql = `
		SELECT id, uid, name, description, created_at, updated_at
		FROM "group"
		WHERE uid = $1
	`

	var g model.Group
	err := r.db.QueryRow(ctx, sql, uid).Scan(
		&g.ID, &g.UID, &g.Name, &g.Description, &g.CreatedAt, &g.UpdatedAt,
	)
	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("group with uid %s not found", uid)
		}
		return nil, fmt.Errorf("failed to get group: %w", err)
	}

	return &g, nil
}

func (r *groupRepository) List(ctx context.Context, pagination *param.PaginationParam, filter *param.GroupListFilterParam) (model.Groups, error) {
	baseSQL := `
		SELECT id, uid, name, description, created_at, updated_at
		FROM "group"
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
		if filter.Name != nil {
			baseSQL += fmt.Sprintf(" AND name = $%d", argIdx)
			args = append(args, *filter.Name)
			argIdx++
		}
		if filter.Query != nil {
			baseSQL += fmt.Sprintf(" AND (name ILIKE $%d OR description ILIKE $%d)", argIdx, argIdx+1)
			args = append(args, "%"+*filter.Query+"%", "%"+*filter.Query+"%")
			argIdx += 2
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
		return model.Groups{}, fmt.Errorf("failed to list groups: %w", err)
	}
	defer rows.Close()

	var items []model.Group
	for rows.Next() {
		var g model.Group
		err := rows.Scan(&g.ID, &g.UID, &g.Name, &g.Description, &g.CreatedAt, &g.UpdatedAt)
		if err != nil {
			return model.Groups{}, fmt.Errorf("failed to scan group: %w", err)
		}
		items = append(items, g)
	}

	if rows.Err() != nil {
		return model.Groups{}, fmt.Errorf("error iterating groups: %w", rows.Err())
	}

	// Get total count
	countSQL := `SELECT COUNT(*) FROM "group" WHERE 1=1`
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
		if filter.Name != nil {
			countSQL += fmt.Sprintf(" AND name = $%d", countArgIdx)
			countArgs = append(countArgs, *filter.Name)
			countArgIdx++
		}
		if filter.Query != nil {
			countSQL += fmt.Sprintf(" AND (name ILIKE $%d OR description ILIKE $%d)", countArgIdx, countArgIdx+1)
			countArgs = append(countArgs, "%"+*filter.Query+"%", "%"+*filter.Query+"%")
			countArgIdx += 2
		}
	}

	var total int64
	err = r.db.QueryRow(ctx, countSQL, countArgs...).Scan(&total)
	if err != nil {
		return model.Groups{}, fmt.Errorf("failed to get total count: %w", err)
	}

	return model.Groups{
		Items: items,
		Meta: model.Meta{
			Total: total,
			Page:  offset/limit + 1,
			Limit: limit,
		},
	}, nil
}

func (r *groupRepository) ListPermission(ctx context.Context, groupID int64, pagination *param.PaginationParam, filter *param.GroupPermissionListFilterParam) (model.GroupPermissions, error) {
	baseSQL := `
		SELECT gp.id, gp.uid, gp.group_id, g.uid as group_uid,
		       gp.permission_id, p.uid as permission_uid, p.resource, p.action, p.description,
		       gp.created_at
		FROM group_permission gp
		JOIN "group" g ON gp.group_id = g.id
		JOIN permission p ON gp.permission_id = p.id
		WHERE gp.group_id = $1
	`
	args := []interface{}{groupID}
	argIdx := 2

	// Apply filters
	if filter != nil {
		if len(filter.IDs) > 0 {
			baseSQL += fmt.Sprintf(" AND gp.id = ANY($%d)", argIdx)
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
	orderByValue := r.validateOrderByGroupPermission(pagination, "created_at")
	orderBy := "gp." + orderByValue
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
		return model.GroupPermissions{}, fmt.Errorf("failed to list group permissions: %w", err)
	}
	defer rows.Close()

	var items []model.GroupPermission
	for rows.Next() {
		var gp model.GroupPermission
		err := rows.Scan(
			&gp.ID, &gp.UID, &gp.GroupID, &gp.GroupUID,
			&gp.PermissionID, &gp.PermissionUID, &gp.PermissionResource, &gp.PermissionAction, &gp.PermissionDescription,
			&gp.CreatedAt,
		)
		if err != nil {
			return model.GroupPermissions{}, fmt.Errorf("failed to scan group permission: %w", err)
		}
		items = append(items, gp)
	}

	if rows.Err() != nil {
		return model.GroupPermissions{}, fmt.Errorf("error iterating group permissions: %w", rows.Err())
	}

	// Get total count
	countSQL := `SELECT COUNT(*) FROM group_permission gp JOIN permission p ON gp.permission_id = p.id WHERE gp.group_id = $1`
	countArgs := []interface{}{groupID}
	countArgIdx := 2
	if filter != nil {
		if len(filter.IDs) > 0 {
			countSQL += fmt.Sprintf(" AND gp.id = ANY($%d)", countArgIdx)
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
		return model.GroupPermissions{}, fmt.Errorf("failed to get total count: %w", err)
	}

	return model.GroupPermissions{
		Items: items,
		Meta: model.Meta{
			Total: total,
			Page:  offset/limit + 1,
			Limit: limit,
		},
	}, nil
}

func (r *groupRepository) GetPermissionByID(ctx context.Context, groupPermissionID int64) (*model.GroupPermission, error) {
	const sql = `
		SELECT gp.id, gp.uid, gp.group_id, g.uid as group_uid,
		       gp.permission_id, p.uid as permission_uid, p.resource, p.action, p.description,
		       gp.created_at
		FROM group_permission gp
		JOIN "group" g ON gp.group_id = g.id
		JOIN permission p ON gp.permission_id = p.id
		WHERE gp.id = $1
	`

	var gp model.GroupPermission
	err := r.db.QueryRow(ctx, sql, groupPermissionID).Scan(
		&gp.ID, &gp.UID, &gp.GroupID, &gp.GroupUID,
		&gp.PermissionID, &gp.PermissionUID, &gp.PermissionResource, &gp.PermissionAction, &gp.PermissionDescription,
		&gp.CreatedAt,
	)
	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("group permission with id %d not found", groupPermissionID)
		}
		return nil, fmt.Errorf("failed to get group permission: %w", err)
	}

	return &gp, nil
}

func (r *groupRepository) GetPermissionByGroupIDAndPermissionUID(ctx context.Context, groupID int64, permissionUID string) (*model.GroupPermission, error) {
	const sql = `
		SELECT gp.id, gp.uid, gp.group_id, g.uid as group_uid,
		       gp.permission_id, p.uid as permission_uid, p.resource, p.action, p.description,
		       gp.created_at
		FROM group_permission gp
		JOIN "group" g ON gp.group_id = g.id
		JOIN permission p ON gp.permission_id = p.id
		WHERE gp.group_id = $1 AND p.uid = $2
	`

	var gp model.GroupPermission
	err := r.db.QueryRow(ctx, sql, groupID, permissionUID).Scan(
		&gp.ID, &gp.UID, &gp.GroupID, &gp.GroupUID,
		&gp.PermissionID, &gp.PermissionUID, &gp.PermissionResource, &gp.PermissionAction, &gp.PermissionDescription,
		&gp.CreatedAt,
	)
	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("group permission for group %d and permission uid %s not found", groupID, permissionUID)
		}
		return nil, fmt.Errorf("failed to get group permission: %w", err)
	}

	return &gp, nil
}

func (r *groupRepository) AddPermission(ctx context.Context, groupID int64, permissionID int64, uid string) error {
	const sql = `
		INSERT INTO group_permission (uid, group_id, permission_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (group_id, permission_id) DO NOTHING
	`

	_, err := r.db.Exec(ctx, sql, uid, groupID, permissionID)
	if err != nil {
		// Check for foreign key violations
		var pgErr *pgconn.PgError
		if stderrors.As(err, &pgErr) {
			if pgErr.ConstraintName == "fk_group_permission_group" {
				return fmt.Errorf("group with id %d not found", groupID)
			}
			if pgErr.ConstraintName == "fk_group_permission_permission" {
				return fmt.Errorf("permission with id %d not found", permissionID)
			}
		}
		return fmt.Errorf("failed to add permission to group: %w", err)
	}

	return nil
}

func (r *groupRepository) RemovePermission(ctx context.Context, groupID int64, permissionID int64) error {
	const sql = `DELETE FROM group_permission WHERE group_id = $1 AND permission_id = $2`

	result, err := r.db.Exec(ctx, sql, groupID, permissionID)
	if err != nil {
		return fmt.Errorf("failed to remove permission from group: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("permission %d not found in group %d", permissionID, groupID)
	}

	return nil
}

func (r *groupRepository) ReplacePermission(ctx context.Context, groupID int64, permissionIDs []int64, uids []string) error {
	tx, ok := r.db.(pgx.Tx)
	if !ok {
		return fmt.Errorf("ReplacePermission requires a transaction")
	}

	if len(permissionIDs) != len(uids) {
		return fmt.Errorf("permissionIDs and uids must have same length")
	}

	// First, delete all existing permissions for the group
	const deleteSQL = `DELETE FROM group_permission WHERE group_id = $1`
	_, err := tx.Exec(ctx, deleteSQL, groupID)
	if err != nil {
		return fmt.Errorf("failed to remove existing permissions: %w", err)
	}

	// Then, add the new permissions
	const insertSQL = `
		INSERT INTO group_permission (uid, group_id, permission_id)
		VALUES ($1, $2, $3)
	`
	for i, permissionID := range permissionIDs {
		_, err := tx.Exec(ctx, insertSQL, uids[i], groupID, permissionID)
		if err != nil {
			var pgErr *pgconn.PgError
			if stderrors.As(err, &pgErr) {
				if pgErr.ConstraintName == "fk_group_permission_group" {
					return fmt.Errorf("group with id %d not found", groupID)
				}
				if pgErr.ConstraintName == "fk_group_permission_permission" {
					return fmt.Errorf("permission with id %d not found", permissionID)
				}
			}
			return fmt.Errorf("failed to add permission %d: %w", permissionID, err)
		}
	}

	return nil
}

// validateOrderBy validates the OrderBy value against allowed Group columns using O(1) map lookup.
func (r *groupRepository) validateOrderBy(pagination *param.PaginationParam, defaultOrderBy string) string {
	if pagination != nil && pagination.OrderBy != nil {
		if _, ok := allowedOrderByGroup[*pagination.OrderBy]; ok {
			return *pagination.OrderBy
		}
	}
	return defaultOrderBy
}

// validateOrderByGroupPermission validates the OrderBy value against allowed GroupPermission columns using O(1) map lookup.
func (r *groupRepository) validateOrderByGroupPermission(pagination *param.PaginationParam, defaultOrderBy string) string {
	if pagination != nil && pagination.OrderBy != nil {
		if _, ok := allowedOrderByGroupPermission[*pagination.OrderBy]; ok {
			return *pagination.OrderBy
		}
	}
	return defaultOrderBy
}
