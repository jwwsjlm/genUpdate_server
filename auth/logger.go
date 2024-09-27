package auth

import (
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	Logger        *zap.Logger
	SugaredLogger *zap.SugaredLogger
)

// ParseLogLevel 解析日志级别字符串
func ParseLogLevel(level string) (zapcore.Level, error) {
	switch level {
	case "debug":
		return zapcore.DebugLevel, nil
	case "info":
		return zapcore.InfoLevel, nil
	case "warn":
		return zapcore.WarnLevel, nil
	case "error":
		return zapcore.ErrorLevel, nil
	case "dpanic":
		return zapcore.DPanicLevel, nil
	case "panic":
		return zapcore.PanicLevel, nil
	case "fatal":
		return zapcore.FatalLevel, nil
	default:
		return zapcore.InfoLevel, fmt.Errorf("无效的日志级别: %s", level)
	}
}

// InitLogger 初始化日志记录器
func InitLogger(logLevel zapcore.Level) {
	encoder := getEncoder()

	// 定义日志级别
	infoLevel := zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
		return lev >= logLevel && lev < zap.ErrorLevel
	})
	errorLevel := zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
		return lev >= zap.ErrorLevel
	})

	// 获取日志写入器
	infoWriter := getInfoWriterSyncer()
	errorWriter := getErrorWriterSyncer()

	// 创建核心
	infoCore := zapcore.NewCore(encoder, zapcore.NewMultiWriteSyncer(infoWriter, zapcore.AddSync(os.Stdout)), infoLevel)
	errorCore := zapcore.NewCore(encoder, zapcore.NewMultiWriteSyncer(errorWriter, zapcore.AddSync(os.Stdout)), errorLevel)

	// 合并核心
	core := zapcore.NewTee(infoCore, errorCore)

	// 创建日志记录器
	Logger = zap.New(core, zap.AddCallerSkip(0), zap.AddStacktrace(zap.WarnLevel))
	SugaredLogger = Logger.Sugar()
}

// 自定义时间编码器
func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

// 自定义日志级别编码器
func levelEncoder(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	levelMap := map[zapcore.Level]string{
		zapcore.DebugLevel:  "[DEBUG]",
		zapcore.InfoLevel:   "[INFO]",
		zapcore.WarnLevel:   "[WARN]",
		zapcore.ErrorLevel:  "[ERROR]",
		zapcore.DPanicLevel: "[DPANIC]",
		zapcore.PanicLevel:  "[PANIC]",
		zapcore.FatalLevel:  "[FATAL]",
	}
	enc.AppendString(levelMap[l])
}

// NewEncoderConfig 创建编码器配置
func NewEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "auth",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    levelEncoder,
		EncodeTime:     timeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

// getEncoder 获取 JSON 编码器
func getEncoder() zapcore.Encoder {
	return zapcore.NewJSONEncoder(NewEncoderConfig())
}

// getInfoWriterSyncer 获取 info 级别日志的写入器
func getInfoWriterSyncer() zapcore.WriteSyncer {
	return zapcore.AddSync(&lumberjack.Logger{
		Filename:   "./log/info.log",
		MaxSize:    100,
		MaxBackups: 100,
		MaxAge:     28,
		Compress:   false,
	})
}

// getErrorWriterSyncer 获取 error 级别日志的写入器
func getErrorWriterSyncer() zapcore.WriteSyncer {
	return zapcore.AddSync(&lumberjack.Logger{
		Filename:   "./log/error.log",
		MaxSize:    100,
		MaxBackups: 100,
		MaxAge:     28,
		Compress:   false,
	})
}

// 以下是日志记录的便捷方法

func Debugf(format string, v ...interface{}) {
	SugaredLogger.Debugf(format, v...)
}

func Infof(format string, v ...interface{}) {
	SugaredLogger.Infof(format, v...)
}

func Warnf(format string, v ...interface{}) {
	SugaredLogger.Warnf(format, v...)
}

func Errorf(format string, v ...interface{}) {
	SugaredLogger.Errorf(format, v...)
}

func Panicf(format string, v ...interface{}) {
	SugaredLogger.Panicf(format, v...)
}

func Debug(msg string, fields ...zap.Field) {
	Logger.Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	Logger.Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	Logger.Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	Logger.Error(msg, fields...)
}

func Panic(msg string, fields ...zap.Field) {
	Logger.Panic(msg, fields...)
}
