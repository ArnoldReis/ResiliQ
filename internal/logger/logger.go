package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

/**
 * InitLogger inicializa o logger estruturado para toda a aplicação
 */
func InitLogger() {
	config := zap.NewProductionConfig()
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.OutputPaths = []string{"stdout"}

	var err error
	Log, err = config.Build()
	if err != nil {
		panic(err)
	}
}

/**
 * GetLogger retorna a instância global do logger
 */
func GetLogger() *zap.Logger {
	if Log == nil {
		InitLogger()
	}
	return Log
}
