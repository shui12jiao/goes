package codec

import (
	"bufio"
	"encoding/gob"
	"io"
	"log"
)

type gobCodec struct {
	conn io.ReadWriteCloser
	buf  *bufio.Writer
	dec  *gob.Decoder
	enc  *gob.Encoder
}

func NewGobCodec(conn io.ReadWriteCloser) Codec {
	buf := bufio.NewWriter(conn)
	return &gobCodec{
		conn: conn,
		buf:  buf,
		dec:  gob.NewDecoder(conn),
		enc:  gob.NewEncoder(buf),
	}
}

func (c *gobCodec) Close() error {
	return c.conn.Close()
}

func (c *gobCodec) ReadHeader(h *Header) error {
	return c.dec.Decode(h)
}

func (c *gobCodec) ReadBody(body any) error {
	return c.dec.Decode(body)
}
func (c *gobCodec) Write(h *Header, body any) error {
	if err := c.enc.Encode(h); err != nil {
		if c.buf.Flush() == nil {
			log.Println("rpc: gob error encoding header:", err)
			c.Close()
		}
		return err
	}
	if err := c.enc.Encode(body); err != nil {
		if c.buf.Flush() == nil {
			log.Println("rpc: gob error encoding body:", err)
			c.Close()
		}
		return err
	}

	return c.buf.Flush()
}

// package codec

// import (
// 	"bufio"
// 	"encoding/gob"
// 	"io"
// 	"log"
// )

// type gobCodec struct {
// 	conn io.ReadWriteCloser
// 	bufW *bufio.Writer
// 	bufR *bufio.Reader
// 	dec  *gob.Decoder
// 	enc  *gob.Encoder
// }

// func NewGobCodec(conn io.ReadWriteCloser) Codec {
// 	bufW := bufio.NewWriter(conn)
// 	bufR := bufio.NewReader(conn)
// 	return &gobCodec{
// 		conn: conn,
// 		bufW: bufW,
// 		bufR: bufR,
// 		dec:  gob.NewDecoder(bufR),
// 		enc:  gob.NewEncoder(bufW),
// 	}
// }

// func (c *gobCodec) Close() error {
// 	return c.conn.Close()
// }

// func (c *gobCodec) ReadHeader(h *Header) error {
// 	// return c.dec.Decode(h)
// 	//read header
// 	_,err := c.bufR.Read(h)
// 	return err

// }

// func (c *gobCodec) ReadBody(body any) error {
// 	// return c.dec.Decode(body)
// }
// func (c *gobCodec) Write(h *Header, body any) error {
// 	if err := c.enc.Encode(h); err != nil {
// 		if c.bufW.Flush() == nil {
// 			log.Println("rpc: gob error encoding header:", err)
// 			c.Close()
// 		}
// 		return err
// 	}
// 	if err := c.enc.Encode(body); err != nil {
// 		if c.bufW.Flush() == nil {
// 			log.Println("rpc: gob error encoding body:", err)
// 			c.Close()
// 		}
// 		return err
// 	}

// 	return c.bufW.Flush()
// }
