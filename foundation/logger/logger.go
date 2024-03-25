package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(name string) (*zap.SugaredLogger, error) {
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{"stdout"}
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.DisableStacktrace = true
	config.InitialFields = map[string]any{
		"services": name,
	}

	log, err := config.Build()
	if err != nil {
		return nil, err
	}
	return log.Sugar(), nil
}
