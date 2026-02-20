package repository

import (
	"context"
)

type UnitOfWork interface {
	Do(ctx context.Context, fn func(r RepositoryProvider) error) error
}
