package log

import "go.uber.org/zap"

type Field = zap.Field

var (
	String   = zap.String
	Bytes    = zap.ByteString
	Duration = zap.Duration

	Int8  = zap.Int8
	Int32 = zap.Int32
	Int   = zap.Int
	Int64 = zap.Int64

	Uint8  = zap.Uint8
	Uint32 = zap.Uint32
	Uint   = zap.Uint
	Uint64 = zap.Uint64

	Float64 = zap.Float64

	Any = zap.Any
)
