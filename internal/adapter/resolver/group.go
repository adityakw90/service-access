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

type groupResolver struct {
	db                 PostgrePool
	redisClient        *redis.Client
	redisPrefix        string
	redisCacheDuration time.Duration
	logger             monitoring.Logger
	tracer             monitoring.Tracer
}

type groupIdentity struct {
	id  int64
	uid string
}

func NewGroupResolver(
	db PostgrePool,
	redisClient *redis.Client,
	redisPrefix string,
	redisCacheDuration time.Duration,
	logger monitoring.Logger,
	tracer monitoring.Tracer,
) portResolver.GroupResolver {
	return &groupResolver{
		db:                 db,
		redisClient:        redisClient,
		redisPrefix:        redisPrefix,
		redisCacheDuration: redisCacheDuration,
		logger:             logger,
		tracer:             tracer,
	}
}

func (r *groupResolver) IDsByUIDs(ctx context.Context, uids []string) (map[string]int64, error) {
	newCtx, resvSpan := r.tracer.StartSpan(ctx, "groupResolver.IDsByUIDs")
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
		func(uid string) (*groupIdentity, error) {
			return r.fetchIDFromDB(newCtx, uid)
		},
		func(group *groupIdentity) int64 {
			return group.id
		},
		r.redisCacheDuration,
	)
	if err != nil {
		if errors.Is(err, domainerrors.ErrGroupNotFound) {
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
		attribute.StringSlice("groupUID", uids),
	))

	return result, nil
}

func (r *groupResolver) UIDsByIDs(ctx context.Context, ids []int64) (map[int64]string, error) {
	newCtx, resvSpan := r.tracer.StartSpan(ctx, "groupResolver.UIDsByIDs")
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
		func(id int64) (*groupIdentity, error) {
			return r.fetchUIDFromDB(newCtx, id)
		},
		func(group *groupIdentity) string {
			return group.uid
		},
		r.redisCacheDuration,
	)
	if err != nil {
		if errors.Is(err, domainerrors.ErrGroupNotFound) {
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
		attribute.Int64Slice("groupID", ids),
	))

	return result, nil
}

func (r *groupResolver) fetchIDFromDB(ctx context.Context, uid string) (*groupIdentity, error) {
	var iden groupIdentity

	rows, err := r.db.Query(ctx,
		`SELECT id, uid FROM "group" WHERE uid=$1`, uid,
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
		return nil, domainerrors.ErrGroupNotFound
	}

	return &iden, nil
}

func (r *groupResolver) fetchUIDFromDB(ctx context.Context, id int64) (*groupIdentity, error) {
	var iden groupIdentity

	rows, err := r.db.Query(ctx,
		`SELECT id, uid FROM "group" WHERE id=$1`, id,
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
		return nil, domainerrors.ErrGroupNotFound
	}

	return &iden, nil
}

func (r *groupResolver) Invalidate(ctx context.Context, uids ...string) error {
	if len(uids) == 0 {
		return nil
	}

	keys := make([]string, len(uids))
	for i, uid := range uids {
		keys[i] = r.redisPrefix + ":" + uid + ":id"
	}

	return r.redisClient.Del(ctx, keys...).Err()
}
