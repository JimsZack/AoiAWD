package core

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"net"
	"strings"
	"testing"
	"time"

	"goawd/internal/plugin"
	"goawd/internal/storage"
	"goawd/internal/types"
)

// startReceiver runs a Receiver on a free loopback port and returns its address.
func startReceiver(t *testing.T, mgr *plugin.Manager) (*Receiver, string, storage.Storage) {
	t.Helper()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve port: %v", err)
	}
	addr := l.Addr().String()
	l.Close()

	store := storage.NewMemory()
	recv := NewReceiver(addr, store, NewHub(), mgr)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go func() {
		if err := recv.Start(ctx); err != nil && ctx.Err() == nil {
			t.Logf("receiver stopped: %v", err)
		}
	}()
	return recv, addr, store
}

func dialReceiver(t *testing.T, addr string) net.Conn {
	t.Helper()
	var conn net.Conn
	var err error
	for i := 0; i < 100; i++ {
		conn, err = net.DialTimeout("tcp", addr, 500*time.Millisecond)
		if err == nil {
			return conn
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("dial %s: %v", addr, err)
	return nil
}

func TestReceiverPingPong(t *testing.T) {
	_, addr, _ := startReceiver(t, plugin.NewManager())

	conn := dialReceiver(t, addr)
	defer conn.Close()

	if _, err := conn.Write([]byte(`{"type":"ping"}` + "\n")); err != nil {
		t.Fatalf("write ping: %v", err)
	}

	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	var msg types.ProbeMessage
	if err := json.NewDecoder(conn).Decode(&msg); err != nil {
		t.Fatalf("decode pong: %v", err)
	}
	if msg.Type != types.MsgTypePong {
		t.Errorf("type = %q, want %q", msg.Type, types.MsgTypePong)
	}
}

func TestReceiverWebFlowReportsAlertThroughCaller(t *testing.T) {
	mgr := plugin.NewManager()
	var gotCaller interface{}
	mgr.Register("Web", "processLog", func(caller interface{}, data interface{}) interface{} {
		gotCaller = caller
		if setter, ok := caller.(AlertSetter); ok {
			setter.SetAlert("Web", "test", "flag replaced", "", 1)
		}
		return data
	})

	recv, addr, store := startReceiver(t, mgr)

	conn := dialReceiver(t, addr)
	defer conn.Close()

	payload, err := json.Marshal(types.ProbeMessage{
		Type: types.MsgTypeWeb,
		Data: map[string]interface{}{
			"method": "GET",
			"uri":    "/index.php",
			"remote": "1.2.3.4",
			"buffer": "flag{real}",
		},
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if _, err := conn.Write(append(payload, '\n')); err != nil {
		t.Fatalf("write web: %v", err)
	}

	// The server echoes the (possibly rewritten) response buffer as a raw
	// base64 line, which is what the TapeWorm hook decodes.
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	line, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		t.Fatalf("read response: %v", err)
	}
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(line))
	if err != nil {
		t.Fatalf("response is not base64: %v", err)
	}
	if string(decoded) != "flag{real}" {
		t.Errorf("echoed buffer = %q, want %q", decoded, "flag{real}")
	}

	// The hook must have received the receiver itself as caller.
	if gotCaller != recv {
		t.Errorf("hook caller = %v, want the receiver", gotCaller)
	}

	// ...and the alert raised through it must land in storage.
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if store.Count(context.Background(), types.CollAlert) > 0 &&
			store.Count(context.Background(), types.CollWeb) > 0 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Errorf("alert/web not persisted: alert=%d web=%d",
		store.Count(context.Background(), types.CollAlert),
		store.Count(context.Background(), types.CollWeb))
}

func TestReceiverPWNStreamIsCapped(t *testing.T) {
	_, addr, store := startReceiver(t, plugin.NewManager())

	conn := dialReceiver(t, addr)

	init, err := json.Marshal(types.ProbeMessage{
		Type: types.MsgTypePWN,
		Data: types.PWNInitData{File: "/bin/sh", Type: "stdout", PID: 4242, Maps: ""},
	})
	if err != nil {
		t.Fatalf("marshal init: %v", err)
	}
	if _, err := conn.Write(append(init, '\n')); err != nil {
		t.Fatalf("write pwn init: %v", err)
	}
	// Give the receiver a moment to register the socket as a PWN stream.
	time.Sleep(200 * time.Millisecond)

	for i := 0; i < maxPWNStreamLog+200; i++ {
		if _, err := conn.Write([]byte("AAAA\n")); err != nil {
			t.Fatalf("write stream chunk %d: %v", i, err)
		}
	}
	conn.Close()

	// Closing the connection persists the session; it must stay bounded.
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if store.Count(context.Background(), types.CollPWN) > 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	docs, err := store.All(context.Background(), types.CollPWN)
	if err != nil || len(docs) == 0 {
		t.Fatalf("pwn session not persisted: %v (%d docs)", err, len(docs))
	}
	proc, ok := docs[0].(*types.PwnProcess)
	if !ok {
		t.Fatalf("unexpected doc type %T", docs[0])
	}
	if len(proc.StreamLog) != maxPWNStreamLog {
		t.Errorf("stream log len = %d, want %d", len(proc.StreamLog), maxPWNStreamLog)
	}
	if !proc.Truncated {
		t.Error("Truncated should be true when the stream log cap is hit")
	}
	if proc.Stdout.Group != maxPWNStreamLog+200 {
		t.Errorf("stdout groups = %d, want %d (stats must stay accurate)", proc.Stdout.Group, maxPWNStreamLog+200)
	}
}
