package session

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestFormatLogFilename(t *testing.T) {
	now := time.Date(2026, 9, 10, 8, 7, 0, 0, time.Local)
	tests := []struct {
		name        string
		template    string
		sessionName string
		host        string
		want        string
	}{
		{"default", "", "prod", "server.example.com", "prod_server.example.com_0910_0807.log"},
		{"duplicate session and host", "", "server.example.com", "server.example.com", "server.example.com_0910_0807.log"},
		{"empty host", "", "local shell", "", "local shell_0910_0807.log"},
		{"duplicate tokens", "%S_%H_%H.log", "host", "host", "host.log"},
		{"IPv6 colon", "%H.log", "ipv6", "2001:db8::1", "2001-db8--1.log"},
		{"custom order", "%M-%D %h:%m %S@%H.log", "prod", "host", "09-10 08-07 prod@host.log"},
		{"extension omitted", "%H", "prod", "server.example.com", "server.example.com.log"},
		{"Windows reserved name", "%S.log", "CON", "host", "_CON_.log"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatLogFilename(tt.template, tt.sessionName, tt.host, now); got != tt.want {
				t.Fatalf("formatLogFilename() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestOutputLoggerEnableUsesCollisionSuffix(t *testing.T) {
	dir := t.TempDir()
	first := &OutputLogger{}
	first.SetBuffered(false)
	path1, err := first.Enable(dir, "fixed.log", "session", "host", "ssh")
	if err != nil {
		t.Fatal(err)
	}
	first.Disable()

	second := &OutputLogger{}
	second.SetBuffered(false)
	path2, err := second.Enable(dir, "fixed.log", "session", "host", "ssh")
	if err != nil {
		t.Fatal(err)
	}
	second.Disable()

	if filepath.Base(path1) != "fixed.log" || filepath.Base(path2) != "fixed_2.log" {
		t.Fatalf("paths = %q, %q", path1, path2)
	}
}

func TestOutputLoggerEnableCreatesPrivateLogDirectoryAndFile(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "logs")
	logger := &OutputLogger{}
	path, err := logger.Enable(dir, "private.log", "session", "host", "ssh")
	if err != nil {
		t.Fatal(err)
	}
	logger.Disable()

	if runtime.GOOS == "windows" {
		return
	}

	dirInfo, err := os.Stat(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got := dirInfo.Mode().Perm(); got != 0o700 {
		t.Errorf("log directory permissions = %#o, want %#o", got, 0o700)
	}

	fileInfo, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := fileInfo.Mode().Perm(); got != 0o600 {
		t.Errorf("log file permissions = %#o, want %#o", got, 0o600)
	}
}
