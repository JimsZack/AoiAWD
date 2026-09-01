package tcpclient

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"
)

type Sender struct {
	mu   sync.Mutex
	conn net.Conn
	enc  *json.Encoder
}

func New(host string, port int) (*Sender, error) {
	addr := net.JoinHostPort(host, strconv.Itoa(port))
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("connect %s: %w", addr, err)
	}
	return &Sender{
		conn: conn,
		enc:  json.NewEncoder(conn),
	}, nil
}

func (s *Sender) Send(msg interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.enc.Encode(msg)
}

func (s *Sender) SendRaw(data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.conn.Write(append(data, '\n'))
	return err
}

func (s *Sender) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.conn.Close()
}

func (s *Sender) SendWithRetry(msg interface{}, maxRetries int) error {
	var lastErr error
	for i := 0; i <= maxRetries; i++ {
		if err := s.Send(msg); err == nil {
			return nil
		} else {
			lastErr = err
		}
		time.Sleep(time.Duration(i+1) * time.Second)
	}
	return lastErr
}
