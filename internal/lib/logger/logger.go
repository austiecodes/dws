package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Config struct {
	InfoLogFilePath    string `toml:"info_log_file_path"`
	WarningLogFilePath string `toml:"warning_log_file_path"`
	ErrorLogFilePath   string `toml:"error_log_file_path"`
}

func NewLogger(config *Config) (*zap.Logger, error) {
	infoLogger := &lumberjack.Logger{
		Filename:   config.InfoLogFilePath,
		MaxSize:    100,
		MaxBackups: 2,
		MaxAge:     28,
		Compress:   true,
	}

	warnLogger := &lumberjack.Logger{
		Filename:   config.WarningLogFilePath,
		MaxSize:    100,
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   true,
	}

	errorLogger := &lumberjack.Logger{
		Filename:   config.ErrorLogFilePath,
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

	return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel)), nil
}
