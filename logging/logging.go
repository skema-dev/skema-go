package logging

import (
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logger *zap.SugaredLogger

	levelmap = map[string]zapcore.Level{
		"info":  zap.InfoLevel,
		"debug": zap.DebugLevel,
		"warn":  zap.WarnLevel,
		"error": zap.ErrorLevel,
		"panic": zap.PanicLevel,
		"fatal": zap.FatalLevel,
	}
)

func init() {
	Init("debug", "console")
}

func Init(level string, encoding string, opts ...string) {
	levelKey := strings.ToLower(level)
	levelValue, ok := levelmap[levelKey]
	if !ok {
		levelValue = zapcore.DebugLevel
	}

	outputPath := ""
	if len(opts) > 0 {
		outputPath = opts[0]
	}

	zapConfig := zap.NewProductionConfig()
	zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	if outputPath != "" {
		// clean previous data, regardless it's file or path
		os.RemoveAll(outputPath)
		// recreate the folder (not the file) to be written
		dir := filepath.Dir(outputPath)
		if _, err := os.Stat(dir); err != nil {
			os.MkdirAll(dir, os.ModePerm)
		}
		zapConfig.OutputPaths = append(zapConfig.OutputPaths, outputPath)
		zapConfig.ErrorOutputPaths = append(zapConfig.ErrorOutputPaths, outputPath)
	}

	zapConfig.Level = zap.NewAtomicLevelAt(levelValue)
	zapConfig.Encoding = encoding
	zapLogger, _ := zapConfig.Build()
	logger = zapLogger.Sugar()
}

func Infow(msg string, args ...interface{}) {
	logger.Infow(msg, args...)
	logger.Sync()
}

func Debugw(msg string, args ...interface{}) {
	logger.Debugw(msg, args...)
	logger.Sync()
}

func Warnw(msg string, args ...interface{}) {
	logger.Warnw(msg, args...)
	logger.Sync()
}

func Errorw(msg string, args ...interface{}) {
	logger.Errorw(msg, args...)
	logger.Sync()
}

func Panicw(msg string, args ...interface{}) {
	logger.Panicw(msg, args...)
	logger.Sync()
}

func Fatalw(msg string, args ...interface{}) {
	logger.Fatalw(msg, args...)
	logger.Sync()
}

func Infof(format string, args ...interface{}) {
	logger.Infof(format, args...)
	logger.Sync()
}

func Debugf(format string, args ...interface{}) {
	logger.Debugf(format, args...)
	logger.Sync()
}

func Warnf(format string, args ...interface{}) {
	logger.Warnf(format, args...)
	logger.Sync()
}

func Errorf(format string, args ...interface{}) {
	logger.Errorf(format, args...)
	logger.Sync()
}

func Panicf(format string, args ...interface{}) {
	logger.Panicf(format, args...)
	logger.Sync()
}

func Fatalf(format string, args ...interface{}) {
	logger.Fatalf(format, args...)
	logger.Sync()
}
