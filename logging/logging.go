package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	DebugLevel = int8(zap.DebugLevel)
	InfoLevel  = int8(zap.InfoLevel)
	WarnLevel  = int8(zap.WarnLevel)
	ErrorLevel = int8(zap.ErrorLevel)
	PanicLevel = int8(zap.PanicLevel)
	FatalLevel = int8(zap.FatalLevel)
)

var (
	logger       *zap.Logger
	simpleLogger *zap.SugaredLogger
)

func Init(level int8, encoding string) {
	zapConfig := zap.NewProductionConfig()
	zapConfig.Level = zap.NewAtomicLevelAt(zapcore.Level(level))
	zapConfig.Encoding = encoding

	logger, _ = zapConfig.Build()
}

func Infow(msg string, args ...interface{}) {
	logger.Sugar().Infow(msg, args...)
	logger.Sync()
}

func Debugw(msg string, args ...interface{}) {
	logger.Sugar().Debugw(msg, args...)
	logger.Sync()
}

func Warnw(msg string, args ...interface{}) {
	logger.Sugar().Warnw(msg, args...)
	logger.Sync()
}

func Errorw(msg string, args ...interface{}) {
	logger.Sugar().Errorw(msg, args...)
	logger.Sync()
}

func Panicw(msg string, args ...interface{}) {
	logger.Sugar().Panicw(msg, args...)
	logger.Sync()
}

func Fatalw(msg string, args ...interface{}) {
	logger.Sugar().Fatalw(msg, args...)
	logger.Sync()
}

func Infof(format string, args ...interface{}) {
	logger.Sugar().Infof(format, args...)
	logger.Sync()
}

func Debugf(format string, args ...interface{}) {
	logger.Sugar().Debugf(format, args...)
	logger.Sync()
}

func Warnf(format string, args ...interface{}) {
	logger.Sugar().Warnf(format, args...)
	logger.Sync()
}

func Errorf(format string, args ...interface{}) {
	logger.Sugar().Errorf(format, args...)
	logger.Sync()
}

func Panicf(format string, args ...interface{}) {
	logger.Sugar().Panicf(format, args...)
	logger.Sync()
}

func Fatalf(format string, args ...interface{}) {
	logger.Sugar().Fatalf(format, args...)
	logger.Sync()
}
