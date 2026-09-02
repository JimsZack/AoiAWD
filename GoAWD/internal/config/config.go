package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
)

type Config struct {
	HTTPAddr     string
	TCPAddr      string
	MongoDBURI   string
	Database     string
	Token        string
	AllowedOrigins string
	PluginDir    string
	PublicDir    string
	Storage      string
	FilePath     string
}

func Default() *Config {
	return &Config{
		HTTPAddr:     "0.0.0.0:1337",
		TCPAddr:      "0.0.0.0:8023",
		MongoDBURI:   "mongodb://127.0.0.1:27017",
		Database:     "goawd",
		AllowedOrigins: "http://localhost:1337,http://127.0.0.1:1337",
		PluginDir:    "./plugins",
		PublicDir:    "./public",
		Storage:      "memory",
		FilePath:     "./goawd.json",
	}
}

func (c *Config) Validate() error {
	if c.HTTPAddr == "" {
		return fmt.Errorf("http address is required")
	}
	if c.TCPAddr == "" {
		return fmt.Errorf("tcp address is required")
	}
	if c.Database == "" {
		c.Database = "goawd"
	}
	if c.PluginDir == "" {
		c.PluginDir = "./plugins"
	}
	if c.PublicDir == "" {
		c.PublicDir = "./public"
	}
	if c.Storage == "" {
		c.Storage = "memory"
	}
	c.Storage = strings.ToLower(c.Storage)
	if c.Storage != "memory" && c.Storage != "file" && c.Storage != "json" {
		return fmt.Errorf("unsupported storage backend: %s (use memory or file)", c.Storage)
	}
	if c.Token == "" {
		c.Token = randomToken()
	}
	return nil
}

func randomToken() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		b2 := make([]byte, 8)
		for i := range b2 {
			b2[i] = byte(i + 1)
		}
		return hex.EncodeToString(b2)
	}
	return hex.EncodeToString(b)
}

func (c *Config) HTTPPort() int {
	return portFromAddr(c.HTTPAddr)
}

func (c *Config) TCPPort() int {
	return portFromAddr(c.TCPAddr)
}

func portFromAddr(addr string) int {
	idx := strings.LastIndex(addr, ":")
	if idx < 0 {
		return 0
	}
	var p int
	fmt.Sscanf(addr[idx+1:], "%d", &p)
	return p
}
