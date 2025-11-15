package logger

import (
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(level string) (*zap.Logger, error) {
	lvl := parseLevel(level)

	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(lvl)
	cfg.Encoding = "json"

	enc := cfg.EncoderConfig
	enc.TimeKey = "ts"
	enc.EncodeTime = zapcore.ISO8601TimeEncoder
	enc.EncodeLevel = zapcore.LowercaseLevelEncoder
	enc.EncodeCaller = zapcore.ShortCallerEncoder
	cfg.EncoderConfig = enc

	return cfg.Build()
}

func parseLevel(s string) zapcore.Level {
	switch strings.ToLower(s) {
	case "debug":
		return zap.DebugLevel
	case "warn":
		return zap.WarnLevel
	case "error":
		return zap.ErrorLevel
	case "dpanic":
		return zap.DPanicLevel
	case "panic":
		return zap.PanicLevel
	case "fatal":
		return zap.FatalLevel
	default:
		return zap.InfoLevel
	}
}
