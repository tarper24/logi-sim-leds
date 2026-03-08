package logging

import (
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func resetLogger() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	// Force GC to finalize the old handler's file so Windows releases the handle
	runtime.GC()
}

func TestSetup_DebugToFile(t *testing.T) {
	dir, err := os.MkdirTemp("", "TestSetup_DebugToFile")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		resetLogger()
		_ = os.RemoveAll(dir)
	}()

	logPath := filepath.Join(dir, "debug.log")

	if err := Setup(true, logPath); err != nil {
		t.Fatal(err)
	}

	slog.Debug("debug message", "key", "val")

	// Reset logger to release the file before reading
	resetLogger()

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "debug message") {
		t.Error("expected debug message in log file")
	}
}

func TestSetup_WarnToFile(t *testing.T) {
	dir, err := os.MkdirTemp("", "TestSetup_WarnToFile")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		resetLogger()
		_ = os.RemoveAll(dir)
	}()

	logPath := filepath.Join(dir, "warn.log")

	if err := Setup(false, logPath); err != nil {
		t.Fatal(err)
	}

	slog.Debug("should not appear")
	slog.Info("should not appear either")
	slog.Warn("warning message")

	// Reset logger to release the file before reading
	resetLogger()

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if strings.Contains(content, "should not appear") {
		t.Error("debug/info messages should be suppressed at warn level")
	}
	if !strings.Contains(content, "warning message") {
		t.Error("expected warn message in log file")
	}
}

func TestSetup_Stderr(t *testing.T) {
	if err := Setup(false, ""); err != nil {
		t.Fatal(err)
	}
}

func TestSetup_InvalidPath(t *testing.T) {
	err := Setup(true, "/nonexistent/dir/file.log")
	if err == nil {
		t.Error("expected error for invalid path")
	}
}
