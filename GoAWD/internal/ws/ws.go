package ws

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
)

const magicGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

const (
	OpContinuation = 0x0
	OpText         = 0x1
	OpBinary       = 0x2
	OpClose        = 0x8
	OpPing         = 0x9
	OpPong         = 0xA
)

type Conn struct {
	conn   net.Conn
	reader *bufio.Reader
}

func Upgrade(w http.ResponseWriter, r *http.Request) (*Conn, error) {
	h, ok := w.(http.Hijacker)
	if !ok {
		return nil, errors.New("response writer does not support hijacking")
	}

	key := r.Header.Get("Sec-WebSocket-Key")
	if key == "" {
		return nil, errors.New("missing Sec-WebSocket-Key")
	}

	accept := computeAccept(key)

	conn, bufrw, err := h.Hijack()
	if err != nil {
		return nil, err
	}

	resp := "HTTP/1.1 101 Switching Protocols\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-WebSocket-Accept: " + accept + "\r\n\r\n"
	if _, err := bufrw.WriteString(resp); err != nil {
		conn.Close()
		return nil, err
	}
	if err := bufrw.Flush(); err != nil {
		conn.Close()
		return nil, err
	}

	return &Conn{conn: conn, reader: bufrw.Reader}, nil
}

func computeAccept(key string) string {
	h := sha1.New()
	h.Write([]byte(key + magicGUID))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func (c *Conn) ReadMessage() (opcode byte, data []byte, err error) {
	for {
		op, payload, err := c.readFrame()
		if err != nil {
			return 0, nil, err
		}
		switch op {
		case OpPing:
			c.writeFrame(OpPong, payload)
			continue
		case OpPong:
			continue
		case OpClose:
			return OpClose, nil, io.EOF
		case OpText, OpBinary:
			return op, payload, nil
		case OpContinuation:
			return op, payload, nil
		}
	}
}

func (c *Conn) readFrame() (opcode byte, payload []byte, err error) {
	header := make([]byte, 2)
	if _, err = io.ReadFull(c.reader, header); err != nil {
		return 0, nil, err
	}

	fin := header[0]&0x80 != 0
	opcode = header[0] & 0x0F
	masked := header[1]&0x80 != 0
	length := int(header[1] & 0x7F)

	switch length {
	case 126:
		ext := make([]byte, 2)
		if _, err = io.ReadFull(c.reader, ext); err != nil {
			return 0, nil, err
		}
		length = int(binary.BigEndian.Uint16(ext))
	case 127:
		ext := make([]byte, 8)
		if _, err = io.ReadFull(c.reader, ext); err != nil {
			return 0, nil, err
		}
		length = int(binary.BigEndian.Uint64(ext))
	}

	maskKey := make([]byte, 4)
	if masked {
		if _, err = io.ReadFull(c.reader, maskKey); err != nil {
			return 0, nil, err
		}
	}

	payload = make([]byte, length)
	if length > 0 {
		if _, err = io.ReadFull(c.reader, payload); err != nil {
			return 0, nil, err
		}
		if masked {
			for i := range payload {
				payload[i] ^= maskKey[i%4]
			}
		}
	}

	_ = fin
	return opcode, payload, nil
}

func (c *Conn) WriteMessage(opcode byte, data []byte) error {
	return c.writeFrame(opcode, data)
}

func (c *Conn) WriteText(data string) error {
	return c.writeFrame(OpText, []byte(data))
}

func (c *Conn) writeFrame(opcode byte, data []byte) error {
	header := make([]byte, 2)
	header[0] = 0x80 | opcode

	n := len(data)
	switch {
	case n <= 125:
		header[1] = byte(n)
	case n <= 65535:
		header[1] = 126
		header = append(header, 0, 0)
		binary.BigEndian.PutUint16(header[2:], uint16(n))
	default:
		header[1] = 127
		header = append(header, 0, 0, 0, 0, 0, 0, 0, 0)
		binary.BigEndian.PutUint64(header[2:], uint64(n))
	}

	if _, err := c.conn.Write(header); err != nil {
		return err
	}
	if n > 0 {
		_, err := c.conn.Write(data)
		return err
	}
	return nil
}

func (c *Conn) Close() error {
	c.writeFrame(OpClose, nil)
	return c.conn.Close()
}

func (c *Conn) UnderlyingConn() net.Conn {
	return c.conn
}

func IsWebSocketRequest(r *http.Request) bool {
	return strings.EqualFold(r.Header.Get("Upgrade"), "websocket") &&
		strings.Contains(strings.ToLower(r.Header.Get("Connection")), "upgrade")
}
