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

type groupPermissionResolver struct {
	db                 PostgrePool
	redisClient        *redis.Client
	redisPrefix        string
	redisCacheDuration time.Duration
	logger             monitoring.Logger
	tracer             monitoring.Tracer
}

type groupPermissionIdentity struct {
	id  int64
	uid string
}

func NewGroupPermissionResolver(
	db PostgrePool,
	redisClient *redis.Client,
	redisPrefix string,
	redisCacheDuration time.Duration,
	logger monitoring.Logger,
	tracer monitoring.Tracer,
) portResolver.GroupPermissionResolver {
	return &groupPermissionResolver{
		db:                 db,
		redisClient:        redisClient,
		redisPrefix:        redisPrefix,
		redisCacheDuration: redisCacheDuration,
		logger:             logger,
		tracer:             tracer,
	}
}

func (r *groupPermissionResolver) IDsByUIDs(ctx context.Context, uids []string) (map[string]int64, error) {
	newCtx, resvSpan := r.tracer.StartSpan(ctx, "groupPermissionResolver.IDsByUIDs")
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
		func(uid string) (*groupPermissionIdentity, error) {
			return r.fetchIDFromDB(newCtx, uid)
		},
		func(gp *groupPermissionIdentity) int64 {
			return gp.id
		},
		r.redisCacheDuration,
	)
	if err != nil {
		if errors.Is(err, domainerrors.ErrGroupPermissionNotFound) {
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
		attribute.StringSlice("groupPermissionUID", uids),
	))

	return result, nil
}

func (r *groupPermissionResolver) UIDsByIDs(ctx context.Context, ids []int64) (map[int64]string, error) {
	newCtx, resvSpan := r.tracer.StartSpan(ctx, "groupPermissionResolver.UIDsByIDs")
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
		func(id int64) (*groupPermissionIdentity, error) {
			return r.fetchUIDFromDB(newCtx, id)
		},
		func(gp *groupPermissionIdentity) string {
			return gp.uid
		},
		r.redisCacheDuration,
	)
	if err != nil {
		if errors.Is(err, domainerrors.ErrGroupPermissionNotFound) {
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
		attribute.Int64Slice("groupPermissionID", ids),
	))

	return result, nil
}

func (r *groupPermissionResolver) IDsByGroupIDAndPermissionIDs(ctx context.Context, params []param.GroupPermissionMapGroupIDPermissionID) (map[param.GroupPermissionMapGroupIDPermissionID]int64, error) {
	// Unimplemented for now
	return nil, nil
}

func (r *groupPermissionResolver) GroupIDsAndPermissionIDsByIDs(ctx context.Context, ids []int64) (map[int64]param.GroupPermissionMapGroupIDPermissionID, error) {
	// Unimplemented for now
	return nil, nil
}

func (r *groupPermissionResolver) fetchIDFromDB(ctx context.Context, uid string) (*groupPermissionIdentity, error) {
	var iden groupPermissionIdentity

	rows, err := r.db.Query(ctx,
		`SELECT id, uid FROM "group_permission" WHERE uid=$1`, uid,
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
		return nil, domainerrors.ErrGroupPermissionNotFound
	}

	return &iden, nil
}

func (r *groupPermissionResolver) fetchUIDFromDB(ctx context.Context, id int64) (*groupPermissionIdentity, error) {
	var iden groupPermissionIdentity

	rows, err := r.db.Query(ctx,
		`SELECT id, uid FROM "group_permission" WHERE id=$1`, id,
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
		return nil, domainerrors.ErrGroupPermissionNotFound
	}

	return &iden, nil
}
