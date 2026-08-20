package logx

import (
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logger *zap.Logger
	once   sync.Once
)

// Init configures the global zap logger.
func Init(level, format string) {
	once.Do(func() {
		var lvl zapcore.Level
		_ = lvl.Set(level)
		cfg := zap.NewProductionConfig()
		cfg.Level = zap.NewAtomicLevelAt(lvl)
		cfg.Encoding = "json"
		if format == "console" {
			cfg.Encoding = "console"
			cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		}
		cfg.DisableStacktrace = true
		l, err := cfg.Build()
		if err != nil {
			l, _ = zap.NewDevelopment()
		}
		logger = l
	})
}

// L returns the configured logger (lazily initialising with defaults).
func L() *zap.Logger {
	if logger == nil {
		Init("info", "json")
	}
	return logger
}

// Sugar returns the sugared logger for ergonomic key/value logging.
func Sugar() *zap.SugaredLogger { return L().Sugar() }

func Info(msg string, keysAndValues ...any)  { Sugar().Infow(msg, keysAndValues...) }
func Warn(msg string, keysAndValues ...any)  { Sugar().Warnw(msg, keysAndValues...) }
func Error(msg string, keysAndValues ...any) { Sugar().Errorw(msg, keysAndValues...) }
func Fatal(msg string, keysAndValues ...any) { Sugar().Fatalw(msg, keysAndValues...) }

// Sync flushes buffered log output.
func Sync() { _ = L().Sync() }
