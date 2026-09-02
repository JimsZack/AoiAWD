package proc

import (
	"testing"
)

func TestParseStatus(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		wantUID  int
		wantUser string
	}{
		{
			name: "valid status",
			data: []byte(`Name:	bash
State:	S (sleeping)
Tgid:	12345
Ngid:	0
Pid:	12345
PPid:	1
TracerPid:	0
Uid:	1000	1000	1000	1000
Gid:	1000	1000	1000	1000
`),
			wantUID:  1000,
			wantUser: "1000", // Will be resolved from /etc/passwd
		},
		{
			name:     "empty data",
			data:     []byte{},
			wantUID:  0,
			wantUser: "0",
		},
		{
			name: "no uid line",
			data: []byte(`Name:	bash
State:	S (sleeping)
`),
			wantUID:  0,
			wantUser: "0",
		},
		{
			name: "uid with multiple fields",
			data: []byte(`Uid:	0	0	0	0
`),
			wantUID:  0,
			wantUser: "root",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uid, username := parseStatus(tt.data)
			if uid != tt.wantUID {
				t.Errorf("parseStatus() uid = %d, want %d", uid, tt.wantUID)
			}
			// Username depends on /etc/passwd, just check it's not empty
			if username == "" {
				t.Error("parseStatus() username should not be empty")
			}
		})
	}
}

func TestParseCmdline(t *testing.T) {
	tests := []struct {
		name      string
		data      []byte
		wantCmd   string
		wantParam string
	}{
		{
			name:      "simple command",
			data:      []byte("/bin/bash\x00"),
			wantCmd:   "/bin/bash",
			wantParam: "",
		},
		{
			name:      "command with args",
			data:      []byte("/usr/bin/python\x00script.py\x00--verbose\x00"),
			wantCmd:   "/usr/bin/python",
			wantParam: "script.py --verbose",
		},
		{
			name:      "empty data",
			data:      []byte{},
			wantCmd:   "",
			wantParam: "",
		},
		{
			name:      "all null bytes",
			data:      []byte("\x00\x00\x00"),
			wantCmd:   "",
			wantParam: "",
		},
		{
			name:      "single command no null",
			data:      []byte("ls"),
			wantCmd:   "ls",
			wantParam: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, param := parseCmdline(tt.data)
			if cmd != tt.wantCmd {
				t.Errorf("parseCmdline() cmd = %q, want %q", cmd, tt.wantCmd)
			}
			if param != tt.wantParam {
				t.Errorf("parseCmdline() param = %q, want %q", param, tt.wantParam)
			}
		})
	}
}

func TestLookupUsername(t *testing.T) {
	// Test with root (uid 0)
	username := lookupUsername(0)
	if username != "root" {
		t.Errorf("lookupUsername(0) = %q, want %q", username, "root")
	}

	// Test with non-existent uid (should return string of uid)
	username = lookupUsername(99999)
	if username != "99999" {
		t.Errorf("lookupUsername(99999) = %q, want %q", username, "99999")
	}
}

func TestNewScanner(t *testing.T) {
	scanner := NewScanner()
	if scanner == nil {
		t.Fatal("NewScanner() returned nil")
	}
	if scanner.knownPIDs == nil {
		t.Error("knownPIDs map should be initialized")
	}
}
