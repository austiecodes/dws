package clients

import (
	"fmt"

	"github.com/BurntSushi/toml"
	"github.com/austiecodes/dws/lib/resources"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type LogConfig struct {
	InfoLogFilePath    string `toml:"info_log_file_path"`
	WarningLogFilePath string `toml:"warning_log_file_path"`
	ErrorLogFilePath   string `toml:"error_log_file_path"`
}

type LoggerClient struct {
	config LogConfig
}

func NewLoggerClient() *LoggerClient {
	return &LoggerClient{}
}

func (l *LoggerClient) LoadConfig() error {
	var config struct {
		Log LogConfig `toml:"log"`
	}
	if _, err := toml.DecodeFile("conf/log.toml", &config); err != nil {
		return fmt.Errorf("error loading log config: %w", err)
	}
	l.config = config.Log
	return nil
}

func (l *LoggerClient) Init() error {
	infoLogger := &lumberjack.Logger{
		Filename:   l.config.InfoLogFilePath,
		MaxSize:    100,
		MaxBackups: 2,
		MaxAge:     28,
		Compress:   true,
	}

	warnLogger := &lumberjack.Logger{
		Filename:   l.config.WarningLogFilePath,
		MaxSize:    100,
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   true,
	}

	errorLogger := &lumberjack.Logger{
		Filename:   l.config.ErrorLogFilePath,
		MaxSize:    100,
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   true,
	}

	infoSyncer := zapcore.AddSync(infoLogger)
	warnSyncer := zapcore.AddSync(warnLogger)
	errorSyncer := zapcore.AddSync(errorLogger)

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoder := zapcore.NewJSONEncoder(encoderConfig)

	infoCore := zapcore.NewCore(encoder, infoSyncer, zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level == zapcore.InfoLevel
	}))

	warnCore := zapcore.NewCore(encoder, warnSyncer, zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level == zapcore.WarnLevel
	}))

	errorCore := zapcore.NewCore(encoder, errorSyncer, zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level >= zapcore.ErrorLevel
	}))

	core := zapcore.NewTee(infoCore, warnCore, errorCore)
	resources.Logger = zap.New(core, zap.AddCaller())
	return resources.Logger.Sync()
}
