package codec

import "io"

type Header struct {
	ServiceMethod string
	Seq           uint64
	Error         string
}

type Codec interface {
	io.Closer
	ReadHeader(*Header) error
	ReadBody(any) error
	Write(*Header, any) error
}

type CodecFunc func(io.ReadWriteCloser) Codec

type Type = string

const (
	GobType Type = "application/gob"
	// JsonType Type = "application/json"
)

var CodecFuncMap map[Type]CodecFunc

func init() {
	CodecFuncMap = make(map[Type]CodecFunc)
	CodecFuncMap[GobType] = NewGobCodec
}
