package rbac

import (
	"embed"
	"errors"
	"fmt"
	"sync"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"

	libdb "github.com/austiecodes/dws/internal/lib/db"
)

//go:embed model.conf
var modelFS embed.FS

var (
	globalMu       sync.RWMutex
	globalEnforcer *casbin.Enforcer
)

// Init initializes the global Casbin enforcer with GORM adapter.
// It uses the embedded model.conf and connects to the existing database.
// Subsequent calls are no-ops and return the already-established enforcer.
func Init() (*casbin.Enforcer, error) {
	globalMu.Lock()
	defer globalMu.Unlock()

	if globalEnforcer != nil {
		return globalEnforcer, nil
	}

	db, err := libdb.Instance()
	if err != nil {
		return nil, fmt.Errorf("get db instance: %w", err)
	}

	adapter, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		return nil, fmt.Errorf("create gorm adapter: %w", err)
	}

	modelData, err := modelFS.ReadFile("model.conf")
	if err != nil {
		return nil, fmt.Errorf("read embedded model.conf: %w", err)
	}

	m, err := model.NewModelFromString(string(modelData))
	if err != nil {
		return nil, fmt.Errorf("parse casbin model: %w", err)
	}

	enforcer, err := casbin.NewEnforcer(m, adapter)
	if err != nil {
		return nil, fmt.Errorf("create enforcer: %w", err)
	}

	if err := enforcer.LoadPolicy(); err != nil {
		return nil, fmt.Errorf("load policy: %w", err)
	}

	globalEnforcer = enforcer
	return globalEnforcer, nil
}

// Instance returns the global Casbin enforcer if it has been initialized.
func Instance() (*casbin.Enforcer, error) {
	globalMu.RLock()
	enforcer := globalEnforcer
	globalMu.RUnlock()

	if enforcer == nil {
		return nil, errors.New("casbin enforcer not initialized")
	}
	return enforcer, nil
}

// MustInstance returns the global Casbin enforcer or panics if not initialized.
func MustInstance() *casbin.Enforcer {
	enforcer, err := Instance()
	if err != nil {
		panic(err)
	}
	return enforcer
}

// Close releases the global enforcer.
func Close() {
	globalMu.Lock()
	defer globalMu.Unlock()
	globalEnforcer = nil
}
