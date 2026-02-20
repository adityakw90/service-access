package resolver

import (
	"time"

	monitoring "github.com/adityakw90/go-monitoring"
	portResolver "github.com/adityakw90/service-access/internal/core/port/resolver"
	"github.com/redis/go-redis/v9"
)

type resolverProvider struct {
	db                 PostgrePool
	redisClient        *redis.Client
	redisPrefix        string
	redisCacheDuration time.Duration
	logger             monitoring.Logger
	tracer             monitoring.Tracer
}

func NewResolverProvider(
	db PostgrePool,
	redisClient *redis.Client,
	redisPrefix string,
	redisCacheDuration time.Duration,
	logger monitoring.Logger,
	tracer monitoring.Tracer,
) portResolver.ResolverProvider {
	return &resolverProvider{
		db:                 db,
		redisClient:        redisClient,
		redisPrefix:        redisPrefix,
		redisCacheDuration: redisCacheDuration,
		logger:             logger,
		tracer:             tracer,
	}
}

func (p *resolverProvider) Permission() portResolver.PermissionResolver {
	return NewPermissionResolver(
		p.db,
		p.redisClient,
		p.redisPrefix+":permission",
		p.redisCacheDuration,
		p.logger,
		p.tracer,
	)
}

func (p *resolverProvider) Group() portResolver.GroupResolver {
	return NewGroupResolver(
		p.db,
		p.redisClient,
		p.redisPrefix+":group",
		p.redisCacheDuration,
		p.logger,
		p.tracer,
	)
}

func (p *resolverProvider) Role() portResolver.RoleResolver {
	return NewRoleResolver(
		p.db,
		p.redisClient,
		p.redisPrefix+":role",
		p.redisCacheDuration,
		p.logger,
		p.tracer,
	)
}

func (p *resolverProvider) GroupPermission() portResolver.GroupPermissionResolver {
	return NewGroupPermissionResolver(
		p.db,
		p.redisClient,
		p.redisPrefix+":group_permission",
		p.redisCacheDuration,
		p.logger,
		p.tracer,
	)
}
