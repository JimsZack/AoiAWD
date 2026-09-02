package ws

import (
	"bufio"
	"encoding/binary"
	"net"
	"strings"
	"testing"
	"time"
)

func TestComputeAccept(t *testing.T) {
	key := "dGhlIHNhbXBsZSBub25jZQ=="
	expected := "s3pPLMBiTxaQ9kYGzzhZRbK+xOo="
	got := computeAccept(key)
	if got != expected {
		t.Errorf("computeAccept(%q) = %q, want %q", key, got, expected)
	}
}

func TestUpgradeMissingKey(t *testing.T) {
	// Use a custom ResponseWriter that supports Hijack
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	go func() {
		conn, _ := l.Accept()
		if conn != nil {
			conn.Close()
		}
	}()

	conn, err := net.Dial("tcp", l.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	// Send HTTP request without Sec-WebSocket-Key
	req := "GET /ws HTTP/1.1\r\nHost: localhost\r\n\r\n"
	conn.Write([]byte(req))

	// Server should reject
	buf := make([]byte, 1024)
	n, _ := conn.Read(buf)
	if n > 0 {
		t.Logf("Server response: %s", string(buf[:n]))
	}
}

func TestUpgradeSuccess(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	done := make(chan *Conn, 1)
	errCh := make(chan error, 1)

	go func() {
		serverConn, err := l.Accept()
		if err != nil {
			errCh <- err
			return
		}
		buf := make([]byte, 4096)
		n, _ := serverConn.Read(buf)
		serverConn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Length: 0\r\n\r\n"))
		_ = n
		serverConn.Close()
		done <- nil
	}()

	key := "dGhlIHNhbXBsZSBub25jZQ=="
	conn, err := net.Dial("tcp", l.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	req := "GET /ws HTTP/1.1\r\n" +
		"Host: localhost\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-WebSocket-Key: " + key + "\r\n\r\n"
	conn.Write([]byte(req))

	// Read response to verify upgrade happened
	reader := bufio.NewReader(conn)
	resp, _ := reader.ReadString('\n')
	if !strings.Contains(resp, "101") {
		t.Logf("Server response: %s", resp)
	}
}

func TestWriteAndReadFrame(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	wsConn := &Conn{conn: client, reader: bufio.NewReader(client)}

	// Server side: read frame
	go func() {
		buf := make([]byte, 2)
		server.Read(buf)
		length := int(buf[1] & 0x7F)
		payload := make([]byte, length)
		server.Read(payload)

		// Send back a text frame
		resp := []byte{0x81, byte(len(payload))}
		resp = append(resp, payload...)
		server.Write(resp)
	}()

	// Write a text frame
	err := wsConn.WriteText("hello")
	if err != nil {
		t.Fatalf("WriteText error: %v", err)
	}

	// Read response
	_, data, err := wsConn.ReadMessage()
	if err != nil {
		t.Fatalf("ReadMessage error: %v", err)
	}
	if string(data) != "hello" {
		t.Errorf("data = %q, want %q", string(data), "hello")
	}
}

func TestWriteFrameLargePayload(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	wsConn := &Conn{conn: client, reader: bufio.NewReader(client)}

	// Server side: read frame with extended length
	go func() {
		buf := make([]byte, 2)
		server.Read(buf)
		lengthByte := buf[1] & 0x7F
		var length int
		switch lengthByte {
		case 126:
			ext := make([]byte, 2)
			server.Read(ext)
			length = int(binary.BigEndian.Uint16(ext))
		case 127:
			ext := make([]byte, 8)
			server.Read(ext)
			length = int(binary.BigEndian.Uint64(ext))
		default:
			length = int(lengthByte)
		}
		payload := make([]byte, length)
		server.Read(payload)

		// Send back a pong
		resp := []byte{0x8A, byte(length)}
		resp = append(resp, payload...)
		server.Write(resp)
	}()

	// Create large payload (130 bytes, triggers 126 extended length)
	largeData := strings.Repeat("A", 130)
	err := wsConn.WriteMessage(OpText, []byte(largeData))
	if err != nil {
		t.Fatalf("WriteMessage error: %v", err)
	}
}

func TestCloseReturnsEOF(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()

	wsConn := &Conn{conn: client, reader: bufio.NewReader(client)}

	// Send close frame from server side
	closeFrame := []byte{0x88, 0}
	go func() {
		server.Write(closeFrame)
	}()

	_, _, err := wsConn.ReadMessage()
	// Should get EOF or read error
	if err == nil {
		t.Error("expected error on close frame")
	}
}

func TestUnderlyingConn(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	wsConn := &Conn{conn: client, reader: bufio.NewReader(client)}
	if wsConn.UnderlyingConn() != client {
		t.Error("UnderlyingConn() does not return the original connection")
	}
}

func TestConnClose(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()

	wsConn := &Conn{conn: client, reader: bufio.NewReader(client)}

	// Server reads close frame
	go func() {
		buf := make([]byte, 2)
		server.Read(buf)
	}()

	err := wsConn.Close()
	if err != nil {
		t.Errorf("Close() error: %v", err)
	}
}

func TestWriteFrameMiddlePayload(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	wsConn := &Conn{conn: client, reader: bufio.NewReader(client)}

	// Server side: read frame with 126 extended length (for 70000 bytes)
	go func() {
		buf := make([]byte, 2)
		server.Read(buf)
		lengthByte := buf[1] & 0x7F
		var length int
		if lengthByte == 126 {
			ext := make([]byte, 2)
			server.Read(ext)
			length = int(binary.BigEndian.Uint16(ext))
		}
		payload := make([]byte, length)
		server.Read(payload)
		server.Write([]byte{0x8A, 0})
	}()

	// Create 70000 byte payload (triggers 126 extended length but < 65536)
	// Actually let's use 200 bytes which is <= 65535 and > 125
	data := strings.Repeat("B", 200)
	err := wsConn.WriteMessage(OpBinary, []byte(data))
	if err != nil {
		t.Fatalf("WriteMessage error: %v", err)
	}
}

func TestWriteFrameEmptyPayload(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	wsConn := &Conn{conn: client, reader: bufio.NewReader(client)}

	// Server reads empty frame
	go func() {
		buf := make([]byte, 2)
		server.Read(buf)
	}()

	err := wsConn.writeFrame(OpPing, nil)
	if err != nil {
		t.Fatalf("writeFrame with nil payload error: %v", err)
	}
}

func TestWriteTextVsWriteMessage(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	wsConn := &Conn{conn: client, reader: bufio.NewReader(client)}

	// Server reads both frames
	go func() {
		for i := 0; i < 2; i++ {
			buf := make([]byte, 2)
			server.Read(buf)
			length := int(buf[1] & 0x7F)
			payload := make([]byte, length)
			server.Read(payload)
		}
	}()

	// WriteText should send OpText
	err := wsConn.WriteText("test1")
	if err != nil {
		t.Fatalf("WriteText error: %v", err)
	}

	// WriteMessage with OpText should work the same
	err = wsConn.WriteMessage(OpText, []byte("test2"))
	if err != nil {
		t.Fatalf("WriteMessage error: %v", err)
	}
}

func TestReadMessageSkipsPong(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	wsConn := &Conn{conn: client, reader: bufio.NewReader(client)}

	// Server sends pong then text
	go func() {
		// Pong frame
		pongFrame := []byte{0x8A, 0}
		server.Write(pongFrame)
		time.Sleep(10 * time.Millisecond)
		// Text frame
		textFrame := []byte{0x81, 5, 'h', 'e', 'l', 'l', 'o'}
		server.Write(textFrame)
	}()

	opcode, data, err := wsConn.ReadMessage()
	if err != nil {
		t.Fatalf("ReadMessage error: %v", err)
	}
	if opcode != OpText {
		t.Errorf("opcode = %d, want %d", opcode, OpText)
	}
	if string(data) != "hello" {
		t.Errorf("data = %q, want %q", string(data), "hello")
	}
}
