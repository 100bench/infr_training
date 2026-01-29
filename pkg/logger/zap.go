package logger

import (
	"context"

	"go.uber.org/zap"
)

type zapLogger struct {
	log *zap.Logger
}

func NewZap(cfg zap.Config) (Logger, error) {
	l, err := cfg.Build()
	if err != nil {
		return nil, err
	}
	return &zapLogger{log: l}, nil
}

func NewDevelopmentLogger() (Logger, error) {
      return NewZap(zap.NewDevelopmentConfig())
  }

func NewProductionLogger() (Logger, error) {
	return NewZap(zap.NewProductionConfig())
}

func (z *zapLogger) Debug(msg string, fields ...Field) {
	z.log.Debug(msg, toZap(fields)...)
}

func (z *zapLogger) Info(msg string, fields ...Field) {
	z.log.Info(msg, toZap(fields)...)
}

func (z *zapLogger) Warn(msg string, fields ...Field) {
	z.log.Warn(msg, toZap(fields)...)
}

func (z *zapLogger) Error(msg string, fields ...Field) {
	z.log.Error(msg, toZap(fields)...)
}

func (z *zapLogger) With(fields ...Field) Logger {
	return &zapLogger{
		log: z.log.With(toZap(fields)...),
	}
}

func (z *zapLogger) WithContext(ctx context.Context) Logger {
	// пример: вытаскиваем request_id
	if v := ctx.Value("request_id"); v != nil {
		return z.With(String("request_id", v.(string)))
	}
	return z
}

func toZap(fields []Field) []zap.Field {
	out := make([]zap.Field, 0, len(fields))
	for _, f := range fields {
		// Type switch для оптимизации
		switch v := f.Value.(type) {
		case string:
			out = append(out, zap.String(f.Key, v))
		case int:
			out = append(out, zap.Int(f.Key, v))
		case int64:
			out = append(out, zap.Int64(f.Key, v))
		case error:
			out = append(out, zap.Error(v))
		default:
			out = append(out, zap.Any(f.Key, v))
		}
    }
	return out
}
