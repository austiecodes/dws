package repository

import (
	"context"

	"gorm.io/gorm"

	libdb "github.com/austiecodes/dws/internal/lib/db"
)

func dbWithContext(ctx context.Context, db *gorm.DB) *gorm.DB {
	if db == nil {
		db = libdb.MustInstance()
	}
	return db.WithContext(ctx)
}
