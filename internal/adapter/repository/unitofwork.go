package repository

import (
	"context"
	"fmt"

	portrepository "github.com/adityakw90/service-access/internal/core/port/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type unitOfWork struct {
	db *pgxpool.Pool
}

func NewUnitOfWork(db *pgxpool.Pool) portrepository.UnitOfWork {
	return &unitOfWork{db: db}
}

func (u *unitOfWork) begin(ctx context.Context) (pgx.Tx, error) {
	tx, err := u.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	return tx, nil
}

func (u *unitOfWork) commit(ctx context.Context, tx pgx.Tx) error {
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (u *unitOfWork) rollback(ctx context.Context, tx pgx.Tx) error {
	if err := tx.Rollback(ctx); err != nil {
		return fmt.Errorf("failed to rollback transaction: %w", err)
	}
	return nil
}

func (u *unitOfWork) Do(ctx context.Context, fn func(r portrepository.Repositories) error) error {
	tx, err := u.begin(ctx)
	if err != nil {
		return err
	}
	defer u.rollback(ctx, tx)

	r := &repositories{
		db: tx,
	}

	if err := fn(r); err != nil {
		return err
	}

	return u.commit(ctx, tx)
}
