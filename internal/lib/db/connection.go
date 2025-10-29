package db

import (
	"database/sql"
	"errors"
	"fmt"
	"sync"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	libconfig "github.com/austiecodes/dws/internal/lib/config"
)

var (
	globalMu sync.RWMutex
	globalDB *gorm.DB
)

// Init initialises the global Postgres connection. Subsequent calls are no-ops
// and will return the already-established handle.
func Init(cfg libconfig.DatabaseConfig) (*gorm.DB, error) {
	globalMu.Lock()
	defer globalMu.Unlock()

	if globalDB != nil {
		return globalDB, nil
	}

	db, err := openPostgres(cfg)
	if err != nil {
		return nil, err
	}

	globalDB = db
	return globalDB, nil
}

// Instance returns the global GORM handle if it has been initialised.
func Instance() (*gorm.DB, error) {
	globalMu.RLock()
	db := globalDB
	globalMu.RUnlock()

	if db == nil {
		return nil, errors.New("postgres connection not initialised")
	}
	return db, nil
}

// MustInstance returns the global GORM handle or panics if it has not been initialised.
func MustInstance() *gorm.DB {
	db, err := Instance()
	if err != nil {
		panic(err)
	}
	return db
}

// Close releases the global connection if it exists.
func Close() error {
	globalMu.Lock()
	defer globalMu.Unlock()

	if globalDB == nil {
		return nil
	}
	sqlDB, err := globalDB.DB()
	if err != nil {
		globalDB = nil
		return err
	}
	err = sqlDB.Close()
	globalDB = nil
	return err
}

func openPostgres(cfg libconfig.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.Name,
		cfg.SSLMode,
	)

	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open postgres connection: %w", err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("retrieve sql.DB: %w", err)
	}

	configurePool(sqlDB, cfg)
	return gormDB, nil
}

func configurePool(sqlDB *sql.DB, cfg libconfig.DatabaseConfig) {
	if cfg.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	}
	if cfg.ConnMaxIdleTime > 0 {
		sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
	}
}
