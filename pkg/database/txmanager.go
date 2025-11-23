package database

import (
	"context"
	"database/sql"

	trmsql "github.com/avito-tech/go-transaction-manager/drivers/sql/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
)

type DB struct {
	db     *sql.DB
	getter *trmsql.CtxGetter
}

func NewDB(db *sql.DB) *DB {
	return &DB{
		db:     db,
		getter: trmsql.DefaultCtxGetter,
	}
}

func (db *DB) Conn(ctx context.Context) trmsql.Tr {
	return db.getter.DefaultTrOrDB(ctx, db.db)
}

type TransactionManagerInterface interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}

type TransactionManager struct {
	manager *manager.Manager
}

func NewTransactionManager(db *sql.DB) (*TransactionManager, error) {
	trManager, err := manager.New(trmsql.NewDefaultFactory(db))

	if err != nil {
		return nil, err
	}

	return &TransactionManager{manager: trManager}, nil
}

func (tm *TransactionManager) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	return tm.manager.Do(ctx, fn)
}
