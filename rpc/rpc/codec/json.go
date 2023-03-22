package codec

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
)

type jsonCodec struct {
	conn io.ReadWriteCloser
	buf  *bufio.Writer
	dec  *json.Decoder
	enc  *json.Encoder
}

func NewJsonCodec(conn io.ReadWriteCloser) Codec {
	buf := bufio.NewWriter(conn)
	return &jsonCodec{
		conn: conn,
		buf:  buf,
		dec:  json.NewDecoder(conn),
		enc:  json.NewEncoder(buf),
	}
}

func (c *jsonCodec) Close() error {
	return c.conn.Close()
}

func (c *jsonCodec) ReadHeader(h *Header) error {
	return c.dec.Decode(h)
}

func (c *jsonCodec) ReadBody(body any) error {
	return c.dec.Decode(body)
}

func (c *jsonCodec) Write(h *Header, body any) error {
	if err := c.enc.Encode(h); err != nil {
		if c.buf.Flush() == nil {
			log.Println("rpc: json error encoding header:", err)
			c.Close()
		}
		return err
	}
	if err := c.enc.Encode(body); err != nil {
		if c.buf.Flush() == nil {
			log.Println("rpc: json error encoding body:", err)
			c.Close()
		}
		return err
	}

	return c.buf.Flush()
}
