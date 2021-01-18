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
	"log"

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
	AddCaller  bool
	CallerSkip int

	Level string

	Debug       bool
	WatchConfig bool

	OutputPath      []string
	ErrorOutputPath []string
}

func newLogger(config *Config) *Logger {
	zapConfig := zap.NewProductionConfig()

	l := &Logger{}
	l.level = zapConfig.Level

	zapConfig.OutputPaths = config.OutputPath
	zapConfig.ErrorOutputPaths = config.ErrorOutputPath
	zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	opts := make([]zap.Option, 0)
	opts = append(opts, zap.AddStacktrace(zap.ErrorLevel))
	if config.AddCaller {
		opts = append(opts, zap.AddCaller(), zap.AddCallerSkip(config.CallerSkip))
	}

	if err := l.level.UnmarshalText([]byte(config.Level)); err != nil {
		panic(fmt.Sprintf("unmarshal log level fail, err msg %s", err.Error()))
	}

	logger, err := zapConfig.Build(opts...)
	if err != nil {
		panic(fmt.Sprintf("build log fail, err msg %s", err.Error()))
	}

	l.logger = logger

	return l
}

func defaultConfig() *Config {
	return &Config{
		AddCaller:       true,
		CallerSkip:      1,
		Level:           "debug",
		Debug:           false,
		OutputPath:      []string{"stdout"},
		ErrorOutputPath: []string{"stderr"},
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
