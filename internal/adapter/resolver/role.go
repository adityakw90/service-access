package resolver

import (
	"context"
	"errors"
	"strconv"
	"time"

	monitoring "github.com/adityakw90/go-monitoring"
	domainerrors "github.com/adityakw90/service-access/internal/core/domain/errors"
	portResolver "github.com/adityakw90/service-access/internal/core/port/resolver"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type roleResolver struct {
	db                 PostgrePool
	redisClient        *redis.Client
	redisPrefix        string
	redisCacheDuration time.Duration
	logger             monitoring.Logger
	tracer             monitoring.Tracer
}

type roleIdentity struct {
	id  int64
	uid string
}

func NewRoleResolver(
	db PostgrePool,
	redisClient *redis.Client,
	redisPrefix string,
	redisCacheDuration time.Duration,
	logger monitoring.Logger,
	tracer monitoring.Tracer,
) portResolver.RoleResolver {
	return &roleResolver{
		db:                 db,
		redisClient:        redisClient,
		redisPrefix:        redisPrefix,
		redisCacheDuration: redisCacheDuration,
		logger:             logger,
		tracer:             tracer,
	}
}

func (r *roleResolver) IDsByUIDs(ctx context.Context, uids []string) (map[string]int64, error) {
	newCtx, resvSpan := r.tracer.StartSpan(ctx, "roleResolver.IDsByUIDs")
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
		func(uid string) (*roleIdentity, error) {
			return r.fetchIDFromDB(newCtx, uid)
		},
		func(role *roleIdentity) int64 {
			return role.id
		},
		r.redisCacheDuration,
	)
	if err != nil {
		if errors.Is(err, domainerrors.ErrRoleNotFound) {
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
		attribute.StringSlice("roleUID", uids),
	))

	return result, nil
}

func (r *roleResolver) UIDsByIDs(ctx context.Context, ids []int64) (map[int64]string, error) {
	newCtx, resvSpan := r.tracer.StartSpan(ctx, "roleResolver.UIDsByIDs")
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
		func(id int64) (*roleIdentity, error) {
			return r.fetchUIDFromDB(newCtx, id)
		},
		func(role *roleIdentity) string {
			return role.uid
		},
		r.redisCacheDuration,
	)
	if err != nil {
		if errors.Is(err, domainerrors.ErrRoleNotFound) {
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
		attribute.Int64Slice("roleID", ids),
	))

	return result, nil
}

func (r *roleResolver) fetchIDFromDB(ctx context.Context, uid string) (*roleIdentity, error) {
	var iden roleIdentity

	rows, err := r.db.Query(ctx,
		`SELECT id, uid FROM "role" WHERE uid=$1`, uid,
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
		return nil, domainerrors.ErrRoleNotFound
	}

	return &iden, nil
}

func (r *roleResolver) fetchUIDFromDB(ctx context.Context, id int64) (*roleIdentity, error) {
	var iden roleIdentity

	rows, err := r.db.Query(ctx,
		`SELECT id, uid FROM "role" WHERE id=$1`, id,
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
		return nil, domainerrors.ErrRoleNotFound
	}

	return &iden, nil
}
