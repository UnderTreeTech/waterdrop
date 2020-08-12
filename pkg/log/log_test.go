package log

import (
	"context"
	"testing"
)

func TestLog(t *testing.T) {
	defaultLogger = newLogger(defaultConfig())

	Info(context.Background(), "info",
		Int64("age", 10),
		String("hello", "world"),
		Any("any", []string{"shanghai", "xuhui"}),
	)

	Warn(context.Background(), "warn",
		String("john", "sun"),
	)

	Debug(context.Background(), "debug",
		String("shanghai", "xuhui"),
	)

	Error(context.Background(), "division zero") //KVString("shanghai", "xuhui"),

	//Panic(context.Background(), "memory leaky", String("stop", "yes"))
}
