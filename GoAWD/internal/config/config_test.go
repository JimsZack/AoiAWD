package config

import (
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()
	if cfg.HTTPAddr != "0.0.0.0:1337" {
		t.Errorf("HTTPAddr = %s", cfg.HTTPAddr)
	}
	if cfg.TCPAddr != "0.0.0.0:8023" {
		t.Errorf("TCPAddr = %s", cfg.TCPAddr)
	}
	if cfg.Storage != "memory" {
		t.Errorf("Storage = %s", cfg.Storage)
	}
}

func TestValidate(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		cfg := Default()
		if err := cfg.Validate(); err != nil {
			t.Errorf("Validate: %v", err)
		}
		if cfg.Token == "" {
			t.Error("Token should be auto-generated")
		}
	})

	t.Run("empty http", func(t *testing.T) {
		cfg := Default()
		cfg.HTTPAddr = ""
		if err := cfg.Validate(); err == nil {
			t.Error("should fail with empty HTTPAddr")
		}
	})

	t.Run("empty tcp", func(t *testing.T) {
		cfg := Default()
		cfg.TCPAddr = ""
		if err := cfg.Validate(); err == nil {
			t.Error("should fail with empty TCPAddr")
		}
	})

	t.Run("invalid storage", func(t *testing.T) {
		cfg := Default()
		cfg.Storage = "redis"
		if err := cfg.Validate(); err == nil {
			t.Error("should fail with invalid storage")
		}
	})

	t.Run("file storage", func(t *testing.T) {
		cfg := Default()
		cfg.Storage = "file"
		if err := cfg.Validate(); err != nil {
			t.Errorf("file storage should be valid: %v", err)
		}
	})
}

func TestPortFromAddr(t *testing.T) {
	cfg := &Config{HTTPAddr: "0.0.0.0:1337", TCPAddr: "0.0.0.0:8023"}
	if cfg.HTTPPort() != 1337 {
		t.Errorf("HTTPPort = %d, want 1337", cfg.HTTPPort())
	}
	if cfg.TCPPort() != 8023 {
		t.Errorf("TCPPort = %d, want 8023", cfg.TCPPort())
	}
}
