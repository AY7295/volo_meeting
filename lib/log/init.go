package log

import (
	"github.com/natefinch/lumberjack"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"log"
	"os"
	"volo_meeting/consts"
)

func Init() {
	zap.ReplaceGlobals(getZapLogger(consts.DefaultLogFilePath, zap.AddCaller()))
}

func getZapLogger(logfile string, options ...zap.Option) *zap.Logger {
	return zap.New(
		zapcore.NewCore(
			getEncoder(),
			getLogWriteSyncer(logfile),
			getLogLevel(),
		),
		options...,
	)
}

func CreateOsLogger(logfile string, level zapcore.Level, options ...zap.Option) (logger *log.Logger) {
	if level < zapcore.DebugLevel || level > zapcore.FatalLevel {
		level = zapcore.ErrorLevel
	}

	// descp there won't be any error because the level is an exact zapcore.Level
	logger = GetOsLogger(getZapLogger(logfile, options...), level)
	return
}

func GetOsLogger(zapLogger *zap.Logger, level zapcore.Level) (logger *log.Logger) {
	if level < zapcore.DebugLevel || level > zapcore.FatalLevel {
		level = zapcore.ErrorLevel
	}

	// descp there won't be any error because the level is an exact zapcore.Level
	logger, _ = zap.NewStdLogAt(zapLogger, level)
	return
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	return zapcore.NewJSONEncoder(encoderConfig)
}

func getLogLevel() zapcore.Level {
	if viper.GetBool("DEBUG") {
		return zapcore.DebugLevel
	} else {
		return zapcore.ErrorLevel
	}
}

func getLogWriteSyncer(logFilePath string) zapcore.WriteSyncer {
	if len(logFilePath) == 0 {
		return os.Stdout
	}

	lumberJackLogger := &lumberjack.Logger{
		Filename:   logFilePath,
		MaxSize:    5,
		MaxBackups: 5,
		MaxAge:     30,
		Compress:   true,
	}

	if viper.GetBool("DEBUG") {
		return zapcore.AddSync(io.MultiWriter(os.Stdout, lumberJackLogger))
	}
	return zapcore.AddSync(lumberJackLogger)
}
