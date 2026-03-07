package repository

import (
	"context"
	stderrors "errors"
	"fmt"

	"github.com/adityakw90/service-access/internal/core/domain/model"
	"github.com/adityakw90/service-access/internal/core/domain/param"
	portrepository "github.com/adityakw90/service-access/internal/core/port/repository"
	"github.com/jackc/pgx/v5/pgconn"
)

// allowedOrderBySubject maps OrderBy string values to their typed enum for validation.
var allowedOrderBySubject = map[string]param.SubjectOrderBy{
	"subject_id":   param.OrderBySubjectID,
	"subject_type": param.OrderBySubjectType,
	"role_id":      param.OrderBySubjectRoleID,
	"assigned_at":  param.OrderBySubjectAssignedAt,
}

type subjectRepository struct {
	db dbExecutor
}

// NewSubjectRepository creates a new subject repository with a connection pool
func NewSubjectRepository(db dbExecutor) portrepository.SubjectRepository {
	return &subjectRepository{db: db}
}

func (r *subjectRepository) Create(ctx context.Context, subject *model.SubjectRole) error {
	const sql = `
		INSERT INTO subject_role (subject_id, subject_type, role_id)
		VALUES ($1, $2, $3)
		RETURNING assigned_at
	`

	err := r.db.QueryRow(ctx, sql, subject.SubjectID, subject.SubjectType, subject.RoleID).Scan(&subject.AssignedAt)
	if err != nil {
		// Check for unique constraint violation
		var pgErr *pgconn.PgError
		if stderrors.As(err, &pgErr) {
			if pgErr.ConstraintName == "uq_subject_role" {
				return fmt.Errorf("subject role assignment already exists")
			}
			if pgErr.ConstraintName == "fk_subject_role_role" {
				return fmt.Errorf("role with id %d not found", subject.RoleID)
			}
		}
		return fmt.Errorf("failed to create subject role: %w", err)
	}

	return nil
}

func (r *subjectRepository) Update(ctx context.Context, subject *model.SubjectRole) error {
	// subject_role has a composite primary key, so we can't really update it
	// The Update operation would essentially be a delete + create
	// For now, let's implement it as checking if the assignment exists and updating nothing if it does
	const sql = `
		UPDATE subject_role
		SET role_id = $3
		WHERE subject_id = $1 AND subject_type = $2
	`

	_, err := r.db.Exec(ctx, sql, subject.SubjectID, subject.SubjectType, subject.RoleID)
	if err != nil {
		// Check for foreign key violation on role_id
		var pgErr *pgconn.PgError
		if stderrors.As(err, &pgErr) {
			if pgErr.ConstraintName == "fk_subject_role_role" {
				return fmt.Errorf("role with id %d not found", subject.RoleID)
			}
		}
		return fmt.Errorf("failed to update subject role: %w", err)
	}

	return nil
}

func (r *subjectRepository) Delete(ctx context.Context, subjectID string, subjectType string, roleID int64) error {
	const sql = `DELETE FROM subject_role WHERE subject_id = $1 AND subject_type = $2 AND role_id = $3`

	result, err := r.db.Exec(ctx, sql, subjectID, subjectType, roleID)
	if err != nil {
		return fmt.Errorf("failed to delete subject role: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("subject role not found")
	}

	return nil
}

func (r *subjectRepository) GetRoles(ctx context.Context, subjectID string, subjectType string) ([]model.SubjectRole, error) {
	const sql = `
		SELECT sr.subject_id, sr.subject_type, sr.role_id, r.uid as role_uid, sr.assigned_at
		FROM subject_role sr
		JOIN role r ON sr.role_id = r.id
		WHERE sr.subject_id = $1 AND sr.subject_type = $2
		ORDER BY sr.assigned_at DESC
	`

	rows, err := r.db.Query(ctx, sql, subjectID, subjectType)
	if err != nil {
		return nil, fmt.Errorf("failed to get subject roles: %w", err)
	}
	defer rows.Close()

	var items []model.SubjectRole
	for rows.Next() {
		var sr model.SubjectRole
		err := rows.Scan(&sr.SubjectID, &sr.SubjectType, &sr.RoleID, &sr.RoleUID, &sr.AssignedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan subject role: %w", err)
		}
		items = append(items, sr)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("error iterating subject roles: %w", rows.Err())
	}

	return items, nil
}

func (r *subjectRepository) GetAllRoles(ctx context.Context, subjectID string, subjectType string) ([]model.Role, error) {
	const sql = `
		SELECT DISTINCT r.id, r.uid, r.group_id, g.uid as group_uid,
		                r.name, r.description, r.created_at, r.updated_at
		FROM subject_role sr
		JOIN role r ON sr.role_id = r.id
		JOIN "group" g ON r.group_id = g.id
		WHERE sr.subject_id = $1 AND sr.subject_type = $2
		ORDER BY r.uid
	`

	rows, err := r.db.Query(ctx, sql, subjectID, subjectType)
	if err != nil {
		return nil, fmt.Errorf("failed to get all roles for subject: %w", err)
	}
	defer rows.Close()

	var roles []model.Role
	for rows.Next() {
		var role model.Role
		err := rows.Scan(
			&role.ID, &role.UID, &role.GroupID, &role.GroupUID,
			&role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}
		roles = append(roles, role)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("error iterating roles: %w", rows.Err())
	}

	return roles, nil
}

func (r *subjectRepository) List(ctx context.Context, pagination *param.PaginationParam, filter *param.SubjectListFilterParam) (model.SubjectRoles, error) {
	baseSQL := `
		SELECT sr.subject_id, sr.subject_type, sr.role_id, r.uid as role_uid, sr.assigned_at
		FROM subject_role sr
		JOIN role r ON sr.role_id = r.id
		WHERE 1=1
	`
	args := []interface{}{}
	argIdx := 1

	// Apply filters
	if filter != nil {
		if filter.SubjectID != nil {
			baseSQL += fmt.Sprintf(" AND sr.subject_id = $%d", argIdx)
			args = append(args, *filter.SubjectID)
			argIdx++
		}
		if filter.SubjectType != nil {
			baseSQL += fmt.Sprintf(" AND sr.subject_type = $%d", argIdx)
			args = append(args, *filter.SubjectType)
			argIdx++
		}
		if filter.RoleID != nil {
			baseSQL += fmt.Sprintf(" AND sr.role_id = $%d", argIdx)
			args = append(args, *filter.RoleID)
			argIdx++
		}
		if filter.RoleUID != nil {
			baseSQL += fmt.Sprintf(" AND r.uid = $%d", argIdx)
			args = append(args, *filter.RoleUID)
			argIdx++
		}
		if filter.Query != nil {
			baseSQL += fmt.Sprintf(" AND (sr.subject_id ILIKE $%d OR sr.subject_type ILIKE $%d)", argIdx, argIdx+1)
			args = append(args, "%"+*filter.Query+"%", "%"+*filter.Query+"%")
			argIdx += 2
		}
	}

	// Apply sorting
	orderByValue := r.validateOrderBy(pagination, "assigned_at")
	orderBy := "sr." + orderByValue
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
		return model.SubjectRoles{}, fmt.Errorf("failed to list subject roles: %w", err)
	}
	defer rows.Close()

	var items []model.SubjectRole
	for rows.Next() {
		var sr model.SubjectRole
		err := rows.Scan(&sr.SubjectID, &sr.SubjectType, &sr.RoleID, &sr.RoleUID, &sr.AssignedAt)
		if err != nil {
			return model.SubjectRoles{}, fmt.Errorf("failed to scan subject role: %w", err)
		}
		items = append(items, sr)
	}

	if rows.Err() != nil {
		return model.SubjectRoles{}, fmt.Errorf("error iterating subject roles: %w", rows.Err())
	}

	// Get total count
	countSQL := `SELECT COUNT(*) FROM subject_role sr JOIN role r ON sr.role_id = r.id WHERE 1=1`
	countArgs := []interface{}{}
	countArgIdx := 1
	if filter != nil {
		if filter.SubjectID != nil {
			countSQL += fmt.Sprintf(" AND sr.subject_id = $%d", countArgIdx)
			countArgs = append(countArgs, *filter.SubjectID)
			countArgIdx++
		}
		if filter.SubjectType != nil {
			countSQL += fmt.Sprintf(" AND sr.subject_type = $%d", countArgIdx)
			countArgs = append(countArgs, *filter.SubjectType)
			countArgIdx++
		}
		if filter.RoleID != nil {
			countSQL += fmt.Sprintf(" AND sr.role_id = $%d", countArgIdx)
			countArgs = append(countArgs, *filter.RoleID)
			countArgIdx++
		}
		if filter.RoleUID != nil {
			countSQL += fmt.Sprintf(" AND r.uid = $%d", countArgIdx)
			countArgs = append(countArgs, *filter.RoleUID)
			countArgIdx++
		}
		if filter.Query != nil {
			countSQL += fmt.Sprintf(" AND (sr.subject_id ILIKE $%d OR sr.subject_type ILIKE $%d)", countArgIdx, countArgIdx+1)
			countArgs = append(countArgs, "%"+*filter.Query+"%", "%"+*filter.Query+"%")
			countArgIdx += 2
		}
	}

	var total int64
	err = r.db.QueryRow(ctx, countSQL, countArgs...).Scan(&total)
	if err != nil {
		return model.SubjectRoles{}, fmt.Errorf("failed to get total count: %w", err)
	}

	return model.SubjectRoles{
		Items: items,
		Meta: model.Meta{
			Total: total,
			Page:  offset/limit + 1,
			Limit: limit,
		},
	}, nil
}

// validateOrderBy validates the OrderBy value against allowed Subject columns using O(1) map lookup.
func (r *subjectRepository) validateOrderBy(pagination *param.PaginationParam, defaultOrderBy string) string {
	if pagination != nil && pagination.OrderBy != nil {
		if _, ok := allowedOrderBySubject[*pagination.OrderBy]; ok {
			return *pagination.OrderBy
		}
	}
	return defaultOrderBy
}
