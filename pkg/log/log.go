/*
 *
 * Copyright 2020 waterdrop authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package log

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/natefinch/lumberjack"

	"go.uber.org/zap/zapcore"

	"github.com/UnderTreeTech/waterdrop/pkg/trace"

	"go.uber.org/zap"
)

// Field is an alias of zap.Field
type Field = zap.Field

var (
	// String is an alias of zap.String
	String = zap.String
	// Bytes is an alias of zap.Bytes
	Bytes = zap.ByteString
	// Duration is an alias of zap.Duration
	Duration = zap.Duration
	// Int8 is an alias of zap.Int8
	Int8 = zap.Int8
	// Int32 is an alias of zap.Int32
	Int32 = zap.Int32
	// Int is an alias of zap.Int
	Int = zap.Int
	// Int64 is an alias of zap.Int64
	Int64 = zap.Int64
	// Uint8 is an alias of zap.Uint8
	Uint8 = zap.Uint8
	// Uint32 is an alias of zap.Uint32
	Uint32 = zap.Uint32
	// Uint is an alias of zap.Uint
	Uint = zap.Uint
	// Uint64 is an alias of zap.Uint64
	Uint64 = zap.Uint64
	// Float64 is an alias of zap.Float64
	Float64 = zap.Float64
	// Any is an alias of zap.Any
	Any = zap.Any
)

// defaultLogger default logger for internal used
var defaultLogger *Logger

// Logger logger definition
type Logger struct {
	logger *zap.Logger
	level  zap.AtomicLevel
}

// Config log configs
type Config struct {
	// Dir log output directory
	Dir string
	// Name log output file name
	Name string
	// Level log level
	Level string
	// CallerSkip log depth
	CallerSkip int
	// FlushInterval log flush interval
	FlushInterval time.Duration
	// Debug log mode, default true
	Debug bool
	// WatchConfig whether watch config file changes
	WatchConfig bool
	// EnableAsyncLog whether flush log async
	EnableAsyncLog bool
	// DisableStacktrace where log stack details if run into error
	DisableStacktrace bool
	// MaxSize max size of log file, it'll rotate log automatically if exceed the max size
	MaxSize int
	// MaxAge max duration of store logs
	MaxAge int
	// MaxBackup max files of backup logs
	MaxBackup int
}

// newLogger returns a Logger pointer
func newLogger(config *Config) *Logger {
	lv := zap.NewAtomicLevelAt(zapcore.InfoLevel)
	if err := lv.UnmarshalText([]byte(config.Level)); err != nil {
		panic(fmt.Sprintf("unmarshal log level fail, err msg %s", err.Error()))
	}

	opts := make([]zap.Option, 0)
	opts = append(opts, zap.AddCallerSkip(config.CallerSkip))
	if !config.DisableStacktrace {
		opts = append(opts, zap.AddStacktrace(zap.ErrorLevel))
	}

	var ws zapcore.WriteSyncer
	if config.Debug {
		ws = os.Stdout
	} else {
		ws = zapcore.AddSync(rotate(config))
	}

	if config.EnableAsyncLog {
		ws = &zapcore.BufferedWriteSyncer{
			WS:            ws,
			FlushInterval: config.FlushInterval,
		}
	}

	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	core := zapcore.NewCore(encoder, ws, lv)
	logger := zap.New(core, opts...)

	return &Logger{
		logger: logger,
		level:  lv,
	}
}

// defaultConfig default logger config
func defaultConfig() *Config {
	return &Config{
		Name:  "run.log",
		Dir:   ".",
		Level: "debug",

		CallerSkip:    1,
		FlushInterval: time.Second,

		Debug:             true,
		EnableAsyncLog:    true,
		DisableStacktrace: false,

		MaxSize:   10,  // 10M
		MaxAge:    30,  // 30 day
		MaxBackup: 100, // 100 backup
	}
}

// New returns a Logger instance
func New(config *Config) *Logger {
	if config == nil {
		config = defaultConfig()
	}

	defaultLogger = newLogger(config)

	return defaultLogger
}

// SetLevel set log level
func (l *Logger) SetLevel(level string) {
	if err := l.level.UnmarshalText([]byte(level)); err != nil {
		log.Printf("set log level fail, err msg %s", err.Error())
	}
}

// Sync flush log
func (l *Logger) Sync() error {
	return l.logger.Sync()
}

// Debug logs are typically voluminous, and are usually disabled in production
func Debug(ctx context.Context, msg string, fields ...Field) {
	defaultLogger.logger.Debug(msg, assembleFields(ctx, fields...)...)
}

// Info logs Info Level
func Info(ctx context.Context, msg string, fields ...Field) {
	defaultLogger.logger.Info(msg, assembleFields(ctx, fields...)...)
}

// Warn logs are more important than Info, but don't need individual human review
func Warn(ctx context.Context, msg string, fields ...Field) {
	defaultLogger.logger.Warn(msg, assembleFields(ctx, fields...)...)
}

// Error logs are high-priority.
// If an application is running smoothly, it shouldn't generate any error-Level logs
func Error(ctx context.Context, msg string, fields ...Field) {
	defaultLogger.logger.Error(msg, assembleFields(ctx, fields...)...)
}

// Panic logs a message then panic
func Panic(ctx context.Context, msg string, fields ...Field) {
	defaultLogger.logger.Panic(msg, assembleFields(ctx, fields...)...)
}

// Debugf logs are typically voluminous without context
// and are usually disabled in production
func Debugf(msg string, fields ...Field) {
	defaultLogger.logger.Debug(msg, fields...)
}

// Infof logs Info Level without context
func Infof(msg string, fields ...Field) {
	defaultLogger.logger.Info(msg, fields...)
}

// Warnf logs are more important than Info
// but don't need individual human review
func Warnf(msg string, fields ...Field) {
	defaultLogger.logger.Warn(msg, fields...)
}

// Errorf logs are high-priority without context
// If an application is running smoothly, it shouldn't generate any error-Level logs.
func Errorf(msg string, fields ...Field) {
	defaultLogger.logger.Error(msg, fields...)
}

// Panic logs a message then panic without context
func Panicf(msg string, fields ...Field) {
	defaultLogger.logger.Panic(msg, fields...)
}

// assembleFields format log fields
func assembleFields(ctx context.Context, fields ...Field) []Field {
	fs := make([]Field, len(fields)+1)
	fs[0] = String("trace_id", trace.TraceID(ctx))
	copy(fs[1:], fields)

	return fs
}

// rotate rotate log according to the predefined polices
func rotate(config *Config) io.Writer {
	return &lumberjack.Logger{
		Filename:   fmt.Sprintf("%s/%s", config.Dir, config.Name),
		MaxSize:    config.MaxSize, // MB
		MaxAge:     config.MaxAge,  // days
		MaxBackups: config.MaxBackup,
		LocalTime:  true,
		Compress:   false,
	}
}
