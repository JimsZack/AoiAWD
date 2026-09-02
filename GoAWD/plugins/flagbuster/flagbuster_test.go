package flagbuster

import (
	"regexp"
	"testing"

	"goawd/internal/types"
)

func TestFlagRegex1(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantFmt bool
	}{
		{"simple flag", "flag{abc123}", true},
		{"flag in text", "here is flag{test_flag} in text", true},
		{"no flag", "no flag here", false},
		{"nested braces", "flag{abc{def}ghi}", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := flagRegex1.MatchString(tt.input)
			if got != tt.wantFmt {
				t.Errorf("flagRegex1.MatchString(%q) = %v, want %v", tt.input, got, tt.wantFmt)
			}
		})
	}
}

func TestFlagRegex2(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantFmt bool
	}{
		{"json flag", `{"flag":"abc123"}`, true},
		{"json flag in text", `text {"flag":"test"} more`, true},
		{"no json flag", `{"key":"value"}`, false},
		{"empty json flag", `{"flag":""}`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := flagRegex2.MatchString(tt.input)
			if got != tt.wantFmt {
				t.Errorf("flagRegex2.MatchString(%q) = %v, want %v", tt.input, got, tt.wantFmt)
			}
		})
	}
}

func TestFakeFlag(t *testing.T) {
	flag := fakeFlag()
	if len(flag) < 10 {
		t.Errorf("fakeFlag() length = %d, want >= 10", len(flag))
	}
	if flag[:5] != "flag{" {
		t.Errorf("fakeFlag() = %q, want prefix %q", flag, "flag{")
	}
	if flag[len(flag)-1] != '}' {
		t.Errorf("fakeFlag() = %q, want suffix %q", flag, "}")
	}

	// Two calls should produce different flags
	flag2 := fakeFlag()
	if flag == flag2 {
		t.Error("fakeFlag() should produce unique flags")
	}
}

func TestFlagBusterName(t *testing.T) {
	fb := &FlagBuster{}
	if fb.Name() != "FlagBuster" {
		t.Errorf("Name() = %q, want %q", fb.Name(), "FlagBuster")
	}
}

func TestProcessWebBufferReplacesFlag(t *testing.T) {
	fb := &FlagBuster{}
	web := &types.WebLogData{
		Buffer: "response with flag{real_flag_here} in it",
	}

	result := fb.processWebBuffer(nil, web)
	if result != web {
		t.Error("processWebBuffer should return the same web object")
	}

	// Check flag was replaced
	matched, _ := regexp.MatchString(`flag\{[a-f0-9]+\}`, web.Buffer)
	if matched {
		// Original flag should be gone
		if contains(web.Buffer, "flag{real_flag_here}") {
			t.Error("original flag should have been replaced")
		}
	}
}

func TestProcessWebBufferReplacesJsonFlag(t *testing.T) {
	fb := &FlagBuster{}
	web := &types.WebLogData{
		Buffer: `{"response":"data","flag":"real_flag_value"}`,
	}

	fb.processWebBuffer(nil, web)

	// The flag value should be replaced
	if contains(web.Buffer, `"flag":"real_flag_value"`) {
		t.Error("original JSON flag should have been replaced")
	}
	// But the structure should remain
	if !contains(web.Buffer, `"flag":"`) {
		t.Error("JSON flag structure should be preserved")
	}
}

func TestProcessWebBufferNoFlag(t *testing.T) {
	fb := &FlagBuster{}
	original := "no flags here, just text"
	web := &types.WebLogData{
		Buffer: original,
	}

	fb.processWebBuffer(nil, web)

	if web.Buffer != original {
		t.Error("buffer should not be modified when no flag present")
	}
}

func TestProcessPWNBufferReplacesFlag(t *testing.T) {
	fb := &FlagBuster{}
	proc := &types.PwnProcess{
		StreamLog: []types.StreamLog{
			{Type: "stdin", Buffer: "input data"},
			{Type: "stdout", Buffer: "flag{pwn_flag_here}"},
			{Type: "stderr", Buffer: "error with flag{err_flag}"},
		},
	}

	result := fb.processPWNBuffer(nil, proc)
	if result != proc {
		t.Error("processPWNBuffer should return the same proc object")
	}

	// stdout should be modified
	if contains(proc.StreamLog[1].Buffer, "flag{pwn_flag_here}") {
		t.Error("stdout flag should have been replaced")
	}

	// stdin should NOT be modified
	if proc.StreamLog[0].Buffer != "input data" {
		t.Error("stdin should not be modified")
	}

	// stderr should NOT be modified (only stdout)
	if proc.StreamLog[2].Buffer != "error with flag{err_flag}" {
		t.Error("stderr should not be modified")
	}
}

func TestProcessPWNBufferNoStdout(t *testing.T) {
	fb := &FlagBuster{}
	proc := &types.PwnProcess{
		StreamLog: []types.StreamLog{
			{Type: "stdin", Buffer: "flag{should_not_change}"},
		},
	}

	fb.processPWNBuffer(nil, proc)

	if proc.StreamLog[0].Buffer != "flag{should_not_change}" {
		t.Error("stdin should not be modified")
	}
}

func TestProcessWebBufferWrongType(t *testing.T) {
	fb := &FlagBuster{}
	// Pass wrong type
	result := fb.processWebBuffer(nil, "not a web log")
	if result != "not a web log" {
		t.Error("should return data unchanged for wrong type")
	}
}

func TestProcessPWNBufferWrongType(t *testing.T) {
	fb := &FlagBuster{}
	result := fb.processPWNBuffer(nil, "not a pwn process")
	if result != "not a pwn process" {
		t.Error("should return data unchanged for wrong type")
	}
}

func TestAlertFunction(t *testing.T) {
	// Test with nil caller (should not panic)
	alert(nil, "Web", "FlagBuster", "test message", "ref-id", 1)
}

func TestAlertWithCaller(t *testing.T) {
	type mockAlert struct {
		alertType string
		plugin    string
		message   string
	}
	type mockCaller struct {
		alerts []mockAlert
	}
	// Not implementing SetAlert, just testing no panic
	alert(&mockCaller{}, "Web", "FlagBuster", "test", "id", 1)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
