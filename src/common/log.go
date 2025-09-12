package common

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	LogDir   = "./log/"
	ChartDir = "./chart/"
)

var Logger *zap.Logger

type LogData struct {
	Level     string  `json:"level"`
	Timestamp float64 `json:"ts"`
	Message   string  `json:"msg"`
}

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
			Filename: LogDir + "app.log",
			MaxSize:  100, // MB
			MaxAge:   7,   // days
			Compress: true,
		})
		core = zapcore.NewCore(
			zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
			w,
			zapcore.InfoLevel,
		)
	}
	return zap.New(core)
}

func NewChart(name string, age time.Duration) *zap.Logger {
	var core zapcore.Core
	w := zapcore.AddSync(&lumberjack.Logger{
		Filename: ChartDir + name + "/app.log",
		MaxSize:  100,                   // MB
		MaxAge:   int(age.Hours() / 24), // days
		Compress: true,
	})
	core = zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		w,
		zapcore.InfoLevel,
	)
	return zap.New(core)
}
