package utils

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Logger *zap.Logger // Global logger variable

func InitLogger(path string, debug bool) (*zap.Logger, error) {
	// Encoder config
	encoderConfig := zap.NewProductionEncoderConfig()
	if debug {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
	}
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// set format log
	encoder := zapcore.NewJSONEncoder(encoderConfig)
	if debug {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// File sink dengan rotasi log
	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   path + time.Now().Format("20060102") + ".log",
		MaxSize:    10, // MB
		MaxBackups: 7,
		MaxAge:     28, // days
		Compress:   true,
	})

	// Stdout sink
	consoleWriter := zapcore.AddSync(os.Stdout)

	// Gabungkan ke dalam satu core
	core := zapcore.NewTee(
		zapcore.NewCore(encoder, fileWriter, zap.InfoLevel),
		zapcore.NewCore(encoder, consoleWriter, zap.InfoLevel),
	)

	// Buat logger
	logger := zap.New(core, zap.AddCaller())

	// SET GLOBAL
	Logger = logger
	zap.ReplaceGlobals(logger)

	return logger, nil
}
