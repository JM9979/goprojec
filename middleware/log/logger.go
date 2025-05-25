package log

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"ginproject/entity/config"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var globalLogger *zap.Logger

type Level string

const (
	DebugLevel Level = "DEBUG"
	InfoLevel  Level = "INFO"
	WarnLevel  Level = "WARN"
	ErrorLevel Level = "ERROR"
)

// 创建异步写入器，带有缓冲区
func getAsyncWriter(ws zapcore.WriteSyncer, bufferSize int) zapcore.WriteSyncer {
	if bufferSize <= 0 {
		// 如果缓冲区大小无效，直接返回原始的WriteSyncer
		return ws
	}

	buffer := make(chan []byte, bufferSize)
	writer := &asyncWriter{
		ws:     ws,
		buffer: buffer,
	}

	// 启动异步写入协程
	go writer.run()

	return writer
}

// asyncWriter 异步日志写入器
type asyncWriter struct {
	ws     zapcore.WriteSyncer
	buffer chan []byte
}

// Write 实现WriteSyncer接口的Write方法
func (w *asyncWriter) Write(p []byte) (int, error) {
	// 创建一个副本，避免数据竞争
	b := make([]byte, len(p))
	copy(b, p)

	// 尝试将数据发送到缓冲通道
	select {
	case w.buffer <- b:
		// 成功加入缓冲区
	default:
		// 缓冲区已满，直接写入，避免丢失日志
		return w.ws.Write(p)
	}

	return len(p), nil
}

// Sync 实现WriteSyncer接口的Sync方法
func (w *asyncWriter) Sync() error {
	return w.ws.Sync()
}

// run 异步写入循环
func (w *asyncWriter) run() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case p := <-w.buffer:
			// 写入日志数据
			_, err := w.ws.Write(p)
			if err != nil {
				fmt.Fprintf(os.Stderr, "异步日志写入失败: %v\n", err)
			}
		case <-ticker.C:
			// 定时刷新，确保数据被写入磁盘
			w.ws.Sync()
		}
	}
}

// InitLogger 使用AppConfig初始化日志
func InitLogger(cfg *config.LogConfig, serverName string) error {
	if cfg.Path != "" {
		if err := os.MkdirAll(filepath.Dir(cfg.Path), 0755); err != nil {
			return fmt.Errorf("创建日志目录失败: %w", err)
		}
	} else {
		return fmt.Errorf("未配置日志路径，日志初始化失败")
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

	// 文件输出
	fileWriter := &lumberjack.Logger{
		Filename:   cfg.Path,
		MaxSize:    100, // MB
		MaxBackups: 3,
		MaxAge:     7, // days
		Compress:   true,
	}

	// 创建异步写入器，缓冲区大小为8192
	asyncFileWriter := getAsyncWriter(zapcore.AddSync(fileWriter), 8192)

	fileEncoder := zapcore.NewJSONEncoder(encoderConfig)
	fileCore := zapcore.NewCore(
		fileEncoder,
		asyncFileWriter,
		getZapLevel(Level(cfg.Level)),
	)
	cores = append(cores, fileCore)

	// 创建logger
	core := zapcore.NewTee(cores...)
	logger := zap.New(core,
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.Fields(zap.String("service", serverName)),
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
func Debug(args ...interface{}) {
	if globalLogger != nil {
		msg := fmt.Sprintln(args...)
		// 移除末尾的换行符
		if len(msg) > 0 && msg[len(msg)-1] == '\n' {
			msg = msg[:len(msg)-1]
		}
		globalLogger.Debug(msg)
	}
}

func Info(args ...interface{}) {
	if globalLogger != nil {
		msg := fmt.Sprintln(args...)
		// 移除末尾的换行符
		if len(msg) > 0 && msg[len(msg)-1] == '\n' {
			msg = msg[:len(msg)-1]
		}
		globalLogger.Info(msg)
	}
}

func Warn(args ...interface{}) {
	if globalLogger != nil {
		msg := fmt.Sprintln(args...)
		// 移除末尾的换行符
		if len(msg) > 0 && msg[len(msg)-1] == '\n' {
			msg = msg[:len(msg)-1]
		}
		globalLogger.Warn(msg)
	}
}

func Error(args ...interface{}) {
	if globalLogger != nil {
		msg := fmt.Sprintln(args...)
		// 移除末尾的换行符
		if len(msg) > 0 && msg[len(msg)-1] == '\n' {
			msg = msg[:len(msg)-1]
		}
		globalLogger.Error(msg)
	}
}

// WithContext相关的日志方法
func DebugWithContext(ctx context.Context, args ...interface{}) {
	if globalLogger != nil {
		msg := fmt.Sprintln(args...)
		// 移除末尾的换行符
		if len(msg) > 0 && msg[len(msg)-1] == '\n' {
			msg = msg[:len(msg)-1]
		}
		fields := appendTraceFields(ctx)
		globalLogger.Debug(msg, fields...)
	}
}

func InfoWithContext(ctx context.Context, args ...interface{}) {
	if globalLogger != nil {
		msg := fmt.Sprintln(args...)
		// 移除末尾的换行符
		if len(msg) > 0 && msg[len(msg)-1] == '\n' {
			msg = msg[:len(msg)-1]
		}
		fields := appendTraceFields(ctx)
		globalLogger.Info(msg, fields...)
	}
}

func WarnWithContext(ctx context.Context, args ...interface{}) {
	if globalLogger != nil {
		msg := fmt.Sprintln(args...)
		// 移除末尾的换行符
		if len(msg) > 0 && msg[len(msg)-1] == '\n' {
			msg = msg[:len(msg)-1]
		}
		fields := appendTraceFields(ctx)
		globalLogger.Warn(msg, fields...)
	}
}

func ErrorWithContext(ctx context.Context, args ...interface{}) {
	if globalLogger != nil {
		msg := fmt.Sprintln(args...)
		// 移除末尾的换行符
		if len(msg) > 0 && msg[len(msg)-1] == '\n' {
			msg = msg[:len(msg)-1]
		}
		fields := appendTraceFields(ctx)
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
	result := make([]zap.Field, 0, len(fields)+2) // 预分配容量为当前fields加上最多两个trace字段
	result = append(result, fields...)            // 添加已有的fields

	if span := trace.SpanFromContext(ctx); span != nil {
		spanContext := span.SpanContext()
		if spanContext.HasTraceID() {
			result = append(result, zap.String("trace_id", spanContext.TraceID().String()))
		}
		if spanContext.HasSpanID() {
			result = append(result, zap.String("span_id", spanContext.SpanID().String()))
		}
	}
	return result
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
