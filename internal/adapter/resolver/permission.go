package resolver

import (
	"context"
	"errors"
	"strconv"
	"time"

	monitoring "github.com/adityakw90/go-monitoring"
	domainerrors "github.com/adityakw90/service-access/internal/core/domain/errors"
	"github.com/adityakw90/service-access/internal/core/domain/param"
	portResolver "github.com/adityakw90/service-access/internal/core/port/resolver"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type permissionResolver struct {
	db                 PostgrePool
	redisClient        *redis.Client
	redisPrefix        string
	redisCacheDuration time.Duration
	logger             monitoring.Logger
	tracer             monitoring.Tracer
}

type permissionIdentity struct {
	id  int64
	uid string
}

func NewPermissionResolver(
	db PostgrePool,
	redisClient *redis.Client,
	redisPrefix string,
	redisCacheDuration time.Duration,
	logger monitoring.Logger,
	tracer monitoring.Tracer,
) portResolver.PermissionResolver {
	return &permissionResolver{
		db:                 db,
		redisClient:        redisClient,
		redisPrefix:        redisPrefix,
		redisCacheDuration: redisCacheDuration,
		logger:             logger,
		tracer:             tracer,
	}
}

func (r *permissionResolver) IDsByUIDs(ctx context.Context, uids []string) (map[string]int64, error) {
	newCtx, resvSpan := r.tracer.StartSpan(ctx, "permissionResolver.IDsByUIDs")
	defer resvSpan.End()

	result, err := mapperID(
		newCtx,
		r.logger,
		r.redisClient,
		uids,
		func(res string) int64 {
			d, _ := strconv.ParseInt(res, 10, 64)
			return d
		},
		func(uid string) string {
			return r.redisPrefix + ":" + uid + ":id"
		},
		func(uid string) (*permissionIdentity, error) {
			return r.fetchIDFromDB(newCtx, uid)
		},
		func(permission *permissionIdentity) int64 {
			return permission.id
		},
		r.redisCacheDuration,
	)
	if err != nil {
		if errors.Is(err, domainerrors.ErrPermissionNotFound) {
			r.logger.Debug("Failed", map[string]interface{}{
				"error.message": err.Error(),
			})
		} else {
			r.logger.Error("error", map[string]interface{}{
				"error.message": err.Error(),
			})
		}
		resvSpan.AddEvent("Error", trace.WithAttributes(
			attribute.String("error.message", err.Error()),
		))
		return nil, err
	}

	resvSpan.AddEvent("success", trace.WithAttributes(
		attribute.StringSlice("permissionUID", uids),
	))

	return result, nil
}

func (r *permissionResolver) UIDsByIDs(ctx context.Context, ids []int64) (map[int64]string, error) {
	newCtx, resvSpan := r.tracer.StartSpan(ctx, "permissionResolver.UIDsByIDs")
	defer resvSpan.End()

	result, err := mapperID(
		newCtx,
		r.logger,
		r.redisClient,
		ids,
		func(res string) string { return res },
		func(id int64) string {
			return r.redisPrefix + ":id:" + strconv.FormatInt(id, 10) + ":uid"
		},
		func(id int64) (*permissionIdentity, error) {
			return r.fetchUIDFromDB(newCtx, id)
		},
		func(permission *permissionIdentity) string {
			return permission.uid
		},
		r.redisCacheDuration,
	)
	if err != nil {
		if errors.Is(err, domainerrors.ErrPermissionNotFound) {
			r.logger.Debug("Failed", map[string]interface{}{
				"error.message": err.Error(),
			})
		} else {
			r.logger.Error("error", map[string]interface{}{
				"error.message": err.Error(),
			})
		}
		resvSpan.AddEvent("Error", trace.WithAttributes(
			attribute.String("error.message", err.Error()),
		))
		return nil, err
	}

	resvSpan.AddEvent("success", trace.WithAttributes(
		attribute.Int64Slice("permissionID", ids),
	))

	return result, nil
}

func (r *permissionResolver) IDsByResourceActions(ctx context.Context, resourceActions []param.PermissionMapResourceAction) (map[param.PermissionMapResourceAction]int64, error) {
	// pending, still deciding to set this as cacheable
	// because resource action can be changed
	return nil, nil
}

func (r *permissionResolver) ResourceActionsByIDs(ctx context.Context, ids []int64) (map[int64]param.PermissionMapResourceAction, error) {
	// pending, still deciding to set this as cacheable
	// because resource action can be changed
	return nil, nil
}

func (r *permissionResolver) fetchIDFromDB(ctx context.Context, uid string) (*permissionIdentity, error) {
	var iden permissionIdentity

	rows, err := r.db.Query(ctx,
		`SELECT id, uid FROM "permission" WHERE uid=$1`, uid,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&iden.id, &iden.uid)
		if err != nil {
			return nil, err
		}
	}

	if iden.id == 0 {
		return nil, domainerrors.ErrPermissionNotFound
	}

	return &iden, nil
}

func (r *permissionResolver) fetchUIDFromDB(ctx context.Context, id int64) (*permissionIdentity, error) {
	var iden permissionIdentity

	rows, err := r.db.Query(ctx,
		`SELECT id, uid FROM "permission" WHERE id=$1`, id,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&iden.id, &iden.uid)
		if err != nil {
			return nil, err
		}
	}

	if iden.id == 0 {
		return nil, domainerrors.ErrPermissionNotFound
	}

	return &iden, nil
}
