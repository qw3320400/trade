package common

import (
	"os"
	"runtime/debug"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Logger *zap.Logger

func init() {
	Logger = NewLogger(true)
}

func InitLogger(testEnv bool) {
	Logger = NewLogger(testEnv)
}

func NewLogger(testEnv bool) *zap.Logger {
	var core zapcore.Core
	if testEnv {
		encoderCfg := zap.NewDevelopmentEncoderConfig()
		consoleEncoder := zapcore.NewConsoleEncoder(encoderCfg)
		core = zapcore.NewCore(consoleEncoder, zapcore.AddSync(zapcore.Lock(os.Stdout)), zapcore.DebugLevel)
	} else {
		w := zapcore.AddSync(&lumberjack.Logger{
			Filename:   "./log/app.log",
			MaxSize:    10, // MB
			MaxBackups: 5,
			MaxAge:     30, // days
			Compress:   true,
		})
		core = zapcore.NewCore(
			zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
			w,
			zapcore.InfoLevel,
		)
	}
	return zap.New(core)
}

func HandlePanic() {
	if r := recover(); r != nil {
		Logger.Sugar().Errorf("catch panic: %v \n stack: %s", r, string(debug.Stack()))
	}
}
