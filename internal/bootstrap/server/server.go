package server

import (
	"fmt"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/austiecodes/dws/internal/app/auth"
	"github.com/austiecodes/dws/internal/app/container"
	"github.com/austiecodes/dws/internal/app/image"
	"github.com/austiecodes/dws/internal/app/task"
	"github.com/austiecodes/dws/internal/router"
	"github.com/austiecodes/dws/lib/resources"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
)

type ServerConfig struct {
	Port        int    `toml:"port"`
	SessionName string `toml:"session_name"`
	SessionKey  string `toml:"session_key"`
	AESKey      string `toml:"aes_key"`
}

type Server struct {
	config ServerConfig
	engine *gin.Engine
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) LoadConfig() error {
	var config struct {
		App ServerConfig `toml:"app"`
	}
	if _, err := toml.DecodeFile("conf/app.toml", &config); err != nil {
		return fmt.Errorf("error loading server config: %w", err)
	}
	s.config = config.App
	return nil
}

func (s *Server) Init() error {
	s.engine = gin.New()
	// load middlewares
	s.engine.Use(ginzap.Ginzap(resources.Logger, time.RFC3339, true))
	s.engine.Use(ginzap.RecoveryWithZap(resources.Logger, true))
	store := cookie.NewStore([]byte(s.config.SessionKey))
	s.engine.Use(sessions.Sessions(s.config.SessionName, store))
	// setup routes
	router.InitRouter([]router.Router{
		auth.NewRouter(),
		container.NewRouter(),
		image.NewRouter(),
		task.NewRouter(),
	})
	return nil
}

func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.config.Port)
	return s.engine.Run(addr)
}
