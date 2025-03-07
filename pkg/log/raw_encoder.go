package log

import (
	"time"

	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

var rawPool = buffer.NewPool()

type rawEncoder struct{}

func NewRawEncoder() zapcore.Encoder {
	return &rawEncoder{}
}

func (r *rawEncoder) Clone() zapcore.Encoder {
	return &rawEncoder{}
}

func (r *rawEncoder) EncodeEntry(ent zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	buf := rawPool.Get()
	buf.AppendString(ent.Message)
	return buf, nil
}

func (r *rawEncoder) AddArray(key string, marshaler zapcore.ArrayMarshaler) error {
	return nil
}
func (r *rawEncoder) AddObject(key string, marshaler zapcore.ObjectMarshaler) error {
	return nil
}
func (r *rawEncoder) AddBinary(key string, value []byte)     {}
func (r *rawEncoder) AddByteString(key string, value []byte) {}
func (r *rawEncoder) AddBool(key string, value bool) {
}
func (r *rawEncoder) AddComplex128(key string, value complex128) {
}
func (r *rawEncoder) AddComplex64(key string, value complex64) {
}
func (r *rawEncoder) AddDuration(key string, value time.Duration) {
}
func (r *rawEncoder) AddFloat64(key string, value float64) {
}
func (r *rawEncoder) AddFloat32(key string, value float32) {
}
func (r *rawEncoder) AddInt(key string, value int) {
}
func (r *rawEncoder) AddInt64(key string, value int64) {
}
func (r *rawEncoder) AddInt32(key string, value int32) {
}
func (r *rawEncoder) AddInt16(key string, value int16) {
}
func (r *rawEncoder) AddInt8(key string, value int8) {
}
func (r *rawEncoder) AddString(key, value string) {
}
func (r *rawEncoder) AddTime(key string, value time.Time) {
}
func (r *rawEncoder) AddUint(key string, value uint) {
}
func (r *rawEncoder) AddUint64(key string, value uint64) {
}
func (r *rawEncoder) AddUint32(key string, value uint32) {
}
func (r *rawEncoder) AddUint16(key string, value uint16) {
}
func (r *rawEncoder) AddUint8(key string, value uint8) {
}
func (r *rawEncoder) AddUintptr(key string, value uintptr) {
}
func (r *rawEncoder) AddReflected(key string, value interface{}) error {
	return nil
}
func (r *rawEncoder) OpenNamespace(key string) {
}
