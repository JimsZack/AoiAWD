package tcpclient

import (
	"bufio"
	"encoding/json"
	"net"
	"testing"
	"time"
)

func TestNewSender(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	// Accept connection in background
	go func() {
		conn, _ := l.Accept()
		if conn != nil {
			conn.Close()
		}
	}()

	addr := l.Addr().(*net.TCPAddr)
	sender, err := New(addr.IP.String(), addr.Port)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	defer sender.Close()
}

func TestNewSenderInvalidHost(t *testing.T) {
	_, err := New("192.0.2.1", 12345) // TEST-NET, should timeout
	if err == nil {
		t.Error("expected error for invalid host")
	}
}

func TestSend(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	msgCh := make(chan []byte, 1)
	go func() {
		conn, _ := l.Accept()
		if conn == nil {
			return
		}
		defer conn.Close()
		reader := bufio.NewReader(conn)
		line, _ := reader.ReadBytes('\n')
		msgCh <- line
	}()

	addr := l.Addr().(*net.TCPAddr)
	sender, err := New(addr.IP.String(), addr.Port)
	if err != nil {
		t.Fatal(err)
	}
	defer sender.Close()

	testMsg := map[string]string{"type": "test", "data": "hello"}
	err = sender.Send(testMsg)
	if err != nil {
		t.Fatalf("Send() error: %v", err)
	}

	select {
	case line := <-msgCh:
		var received map[string]string
		if err := json.Unmarshal(line, &received); err != nil {
			t.Fatalf("failed to unmarshal received message: %v", err)
		}
		if received["type"] != "test" {
			t.Errorf("received type = %q, want %q", received["type"], "test")
		}
		if received["data"] != "hello" {
			t.Errorf("received data = %q, want %q", received["data"], "hello")
		}
	case <-time.After(time.Second):
		t.Error("timeout waiting for message")
	}
}

func TestSendRaw(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	msgCh := make(chan []byte, 1)
	go func() {
		conn, _ := l.Accept()
		if conn == nil {
			return
		}
		defer conn.Close()
		reader := bufio.NewReader(conn)
		line, _ := reader.ReadBytes('\n')
		msgCh <- line
	}()

	addr := l.Addr().(*net.TCPAddr)
	sender, err := New(addr.IP.String(), addr.Port)
	if err != nil {
		t.Fatal(err)
	}
	defer sender.Close()

	err = sender.SendRaw([]byte(`{"raw":"data"}`))
	if err != nil {
		t.Fatalf("SendRaw() error: %v", err)
	}

	select {
	case line := <-msgCh:
		if string(line) != `{"raw":"data"}`+"\n" {
			t.Errorf("received = %q, want %q", string(line), `{"raw":"data"}`+"\n")
		}
	case <-time.After(time.Second):
		t.Error("timeout waiting for raw message")
	}
}

func TestClose(t *testing.T) {
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

	addr := l.Addr().(*net.TCPAddr)
	sender, err := New(addr.IP.String(), addr.Port)
	if err != nil {
		t.Fatal(err)
	}

	err = sender.Close()
	if err != nil {
		t.Errorf("Close() error: %v", err)
	}
}

func TestConcurrentSend(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	msgCount := 0
	mu := make(chan struct{}, 1)
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				reader := bufio.NewReader(c)
				for {
					_, err := reader.ReadBytes('\n')
					if err != nil {
						return
					}
					mu <- struct{}{}
					msgCount++
					<-mu
				}
			}(conn)
		}
	}()

	addr := l.Addr().(*net.TCPAddr)
	sender, err := New(addr.IP.String(), addr.Port)
	if err != nil {
		t.Fatal(err)
	}
	defer sender.Close()

	// Send multiple messages concurrently
	for i := 0; i < 10; i++ {
		go func() {
			sender.Send(map[string]int{"i": 1})
		}()
	}

	time.Sleep(100 * time.Millisecond)

	mu <- struct{}{}
	if msgCount != 10 {
		t.Errorf("msgCount = %d, want 10", msgCount)
	}
	<-mu
}
