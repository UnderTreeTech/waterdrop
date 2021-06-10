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

//default logger
var defaultLogger *Logger

type Logger struct {
	logger *zap.Logger
	level  zap.AtomicLevel
}

type Config struct {
	Dir   string
	Name  string
	Level string

	CallerSkip    int
	FlushInterval time.Duration

	Debug             bool
	WatchConfig       bool
	EnableAsyncLog    bool
	DisableStacktrace bool

	// 日志输出文件最大长度，超过改值则截断
	MaxSize   int
	MaxAge    int
	MaxBackup int
}

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

func New(config *Config) *Logger {
	if config == nil {
		config = defaultConfig()
	}

	defaultLogger = newLogger(config)

	return defaultLogger
}

func (l *Logger) SetLevel(level string) {
	if err := l.level.UnmarshalText([]byte(level)); err != nil {
		log.Printf("set log level fail, err msg %s", err.Error())
	}
}

func (l *Logger) Sync() error {
	return l.logger.Sync()
}

func Debug(ctx context.Context, msg string, fields ...Field) {
	defaultLogger.logger.Debug(msg, assembleFields(ctx, fields...)...)
}

func Info(ctx context.Context, msg string, fields ...Field) {
	defaultLogger.logger.Info(msg, assembleFields(ctx, fields...)...)
}

func Warn(ctx context.Context, msg string, fields ...Field) {
	defaultLogger.logger.Warn(msg, assembleFields(ctx, fields...)...)
}

func Error(ctx context.Context, msg string, fields ...Field) {
	defaultLogger.logger.Error(msg, assembleFields(ctx, fields...)...)
}

func Panic(ctx context.Context, msg string, fields ...Field) {
	defaultLogger.logger.Panic(msg, assembleFields(ctx, fields...)...)
}

func Debugf(msg string, fields ...Field) {
	defaultLogger.logger.Debug(msg, fields...)
}

func Infof(msg string, fields ...Field) {
	defaultLogger.logger.Info(msg, fields...)
}

func Warnf(msg string, fields ...Field) {
	defaultLogger.logger.Warn(msg, fields...)
}

func Errorf(msg string, fields ...Field) {
	defaultLogger.logger.Error(msg, fields...)
}

func Panicf(msg string, fields ...Field) {
	defaultLogger.logger.Panic(msg, fields...)
}

func assembleFields(ctx context.Context, fields ...Field) []Field {
	fs := make([]Field, len(fields)+1)
	fs[0] = String("trace_id", trace.TraceID(ctx))
	copy(fs[1:], fields)

	return fs
}

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
