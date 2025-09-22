package server

import (
	"fmt"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
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
	//
	logger, _ := zap.NewProduction()
	s.engine.Use(ginzap.Ginzap(logger, time.RFC3339, true))
	s.engine.Use(ginzap.RecoveryWithZap(logger, true))
	store := cookie.NewStore([]byte(s.config.SessionKey))
	s.engine.Use(sessions.Sessions(s.config.SessionName, store))
	// setup routes
	return nil
}

func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.config.Port)
	return s.engine.Run(addr)
}
