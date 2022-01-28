package redis

import (
	"context"

	"go.uber.org/zap"
)

type zapLogger struct {
	lg *zap.SugaredLogger
}

func (l *zapLogger) Printf(_ context.Context, format string, v ...interface{}) {
	l.lg.Debugf(format, v...)
}
