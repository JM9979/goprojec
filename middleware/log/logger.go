package log

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"ginproject/entity/config"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var globalLogger *zap.Logger

type Level string

const (
	DebugLevel Level = "Debug"
	InfoLevel  Level = "Info"
	WarnLevel  Level = "Warn"
	ErrorLevel Level = "Error"
)

// InitLogger 使用AppConfig初始化日志
func InitLogger(cfg *config.LogConfig) error {
	if cfg.Path != "" {
		if err := os.MkdirAll(filepath.Dir(cfg.Path), 0755); err != nil {
			return fmt.Errorf("创建日志目录失败: %w", err)
		}
	}

	// 配置zapcore
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 配置日志输出
	var cores []zapcore.Core

	// 控制台输出
	consoleEncoder := zapcore.NewJSONEncoder(encoderConfig)
	consoleCore := zapcore.NewCore(
		consoleEncoder,
		zapcore.AddSync(os.Stdout),
		getZapLevel(Level(cfg.Level)),
	)
	cores = append(cores, consoleCore)

	// 文件输出
	if cfg.Path != "" {
		fileWriter := &lumberjack.Logger{
			Filename:   cfg.Path,
			MaxSize:    100, // MB
			MaxBackups: 3,
			MaxAge:     7, // days
		}
		fileEncoder := zapcore.NewJSONEncoder(encoderConfig)
		fileCore := zapcore.NewCore(
			fileEncoder,
			zapcore.AddSync(fileWriter),
			getZapLevel(Level(cfg.Level)),
		)
		cores = append(cores, fileCore)
	}

	// 创建logger
	core := zapcore.NewTee(cores...)
	logger := zap.New(core,
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.Fields(zap.String("service", "ginproject")),
	)

	globalLogger = logger
	return nil
}

func getZapLevel(level Level) zapcore.Level {
	switch level {
	case DebugLevel:
		return zapcore.DebugLevel
	case InfoLevel:
		return zapcore.InfoLevel
	case WarnLevel:
		return zapcore.WarnLevel
	case ErrorLevel:
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

// 基础日志接口
func Debug(msg string, fields ...zap.Field) {
	if globalLogger != nil {
		globalLogger.Debug(msg, fields...)
	}
}

func Info(msg string, fields ...zap.Field) {
	if globalLogger != nil {
		globalLogger.Info(msg, fields...)
	}
}

func Warn(msg string, fields ...zap.Field) {
	if globalLogger != nil {
		globalLogger.Warn(msg, fields...)
	}
}

func Error(msg string, fields ...zap.Field) {
	if globalLogger != nil {
		globalLogger.Error(msg, fields...)
	}
}

// WithContext相关的日志方法
func DebugWithContext(ctx context.Context, msg string, fields ...zap.Field) {
	if globalLogger != nil {
		fields = appendTraceFields(ctx, fields...)
		globalLogger.Debug(msg, fields...)
	}
}

func InfoWithContext(ctx context.Context, msg string, fields ...zap.Field) {
	if globalLogger != nil {
		fields = appendTraceFields(ctx, fields...)
		globalLogger.Info(msg, fields...)
	}
}

func WarnWithContext(ctx context.Context, msg string, fields ...zap.Field) {
	if globalLogger != nil {
		fields = appendTraceFields(ctx, fields...)
		globalLogger.Warn(msg, fields...)
	}
}

func ErrorWithContext(ctx context.Context, msg string, fields ...zap.Field) {
	if globalLogger != nil {
		fields = appendTraceFields(ctx, fields...)
		globalLogger.Error(msg, fields...)
	}
}

// 格式化日志方法
func Debugf(format string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Debug(fmt.Sprintf(format, args...))
	}
}

func Infof(format string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Info(fmt.Sprintf(format, args...))
	}
}

func Warnf(format string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Warn(fmt.Sprintf(format, args...))
	}
}

func Errorf(format string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Error(fmt.Sprintf(format, args...))
	}
}

// 带上下文的格式化日志方法
func DebugWithContextf(ctx context.Context, format string, args ...interface{}) {
	if globalLogger != nil {
		msg := fmt.Sprintf(format, args...)
		fields := appendTraceFields(ctx)
		globalLogger.Debug(msg, fields...)
	}
}

func InfoWithContextf(ctx context.Context, format string, args ...interface{}) {
	if globalLogger != nil {
		msg := fmt.Sprintf(format, args...)
		fields := appendTraceFields(ctx)
		globalLogger.Info(msg, fields...)
	}
}

func WarnWithContextf(ctx context.Context, format string, args ...interface{}) {
	if globalLogger != nil {
		msg := fmt.Sprintf(format, args...)
		fields := appendTraceFields(ctx)
		globalLogger.Warn(msg, fields...)
	}
}

func ErrorWithContextf(ctx context.Context, format string, args ...interface{}) {
	if globalLogger != nil {
		msg := fmt.Sprintf(format, args...)
		fields := appendTraceFields(ctx)
		globalLogger.Error(msg, fields...)
	}
}

// appendTraceFields 添加追踪相关字段
func appendTraceFields(ctx context.Context, fields ...zap.Field) []zap.Field {
	if span := trace.SpanFromContext(ctx); span != nil {
		spanContext := span.SpanContext()
		if spanContext.HasTraceID() {
			fields = append(fields, zap.String("trace_id", spanContext.TraceID().String()))
		}
		if spanContext.HasSpanID() {
			fields = append(fields, zap.String("span_id", spanContext.SpanID().String()))
		}
	}
	return fields
}

// Field 创建一个zap.Field，方便使用
func Field(key string, value interface{}) zap.Field {
	return zap.Any(key, value)
}

// String 创建一个字符串类型的zap.Field
func String(key string, value string) zap.Field {
	return zap.String(key, value)
}

// Int 创建一个整数类型的zap.Field
func Int(key string, value int) zap.Field {
	return zap.Int(key, value)
}

// Bool 创建一个布尔类型的zap.Field
func Bool(key string, value bool) zap.Field {
	return zap.Bool(key, value)
}

// ErrorField 创建一个错误类型的zap.Field
func ErrorField(err error) zap.Field {
	return zap.Error(err)
}
