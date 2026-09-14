package file

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewRotateFile_validation(t *testing.T) {
	cases := []struct {
		name      string
		directory string
		fileName  string
		suffix    string
		maxSize   int64
		interval  int64
	}{
		{name: "empty directory", directory: "", fileName: "app", suffix: "log"},
		{name: "empty file name", directory: t.TempDir(), fileName: "", suffix: "log"},
		{name: "empty suffix", directory: t.TempDir(), fileName: "app", suffix: ""},
		{name: "negative size", directory: t.TempDir(), fileName: "app", suffix: "log", maxSize: -1},
		{name: "negative interval", directory: t.TempDir(), fileName: "app", suffix: "log", interval: -1},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := NewRotateFile(tc.directory, tc.fileName, tc.suffix, tc.maxSize, tc.interval)
			if !result.HasError() {
				t.Fatalf("expected error")
			}
		})
	}
}

func TestRotateFile_rotateOnSize(t *testing.T) {
	dir := t.TempDir()
	result := NewRotateFile(dir, "size", "txt", 5, 0)
	if result.HasError() {
		t.Fatalf("NewRotateFile failed: %v", result.Error())
	}

	rotateFile := result.Data()
	defer rotateFile.Close()

	if err := rotateFile.WriteString("", "12345"); err != nil {
		t.Fatalf("WriteString failed: %v", err)
	}

	rotated, err := rotateFile.RotatedFilePaths()
	if err != nil {
		t.Fatalf("RotatedFilePaths failed: %v", err)
	}
	if len(rotated) != 1 {
		t.Fatalf("expected 1 rotated file, got %d", len(rotated))
	}
	if !strings.Contains(filepath.Base(rotated[0]), "size_") {
		t.Fatalf("unexpected rotated file name: %s", rotated[0])
	}

	content, err := os.ReadFile(rotated[0])
	if err != nil {
		t.Fatalf("read rotated file failed: %v", err)
	}
	if string(content) != "12345" {
		t.Fatalf("unexpected rotated content: %q", content)
	}

	if err := rotateFile.WriteString("", "next"); err != nil {
		t.Fatalf("second WriteString failed: %v", err)
	}

	activeContent, err := os.ReadFile(filepath.Join(dir, "size.txt"))
	if err != nil {
		t.Fatalf("read active file failed: %v", err)
	}
	if string(activeContent) != "next" {
		t.Fatalf("unexpected active content: %q", activeContent)
	}
}

func TestRotateFile_rotateOnDataKey(t *testing.T) {
	dir := t.TempDir()
	result := NewRotateFile(dir, "key", "txt", 0, 0)
	if result.HasError() {
		t.Fatalf("NewRotateFile failed: %v", result.Error())
	}

	rotateFile := result.Data()
	defer rotateFile.Close()

	if err := rotateFile.WriteString("alpha", "first"); err != nil {
		t.Fatalf("first WriteString failed: %v", err)
	}
	if err := rotateFile.WriteString("beta", "second"); err != nil {
		t.Fatalf("second WriteString failed: %v", err)
	}

	rotated, err := rotateFile.RotatedFilePaths()
	if err != nil {
		t.Fatalf("RotatedFilePaths failed: %v", err)
	}
	if len(rotated) != 1 {
		t.Fatalf("expected 1 rotated file, got %d", len(rotated))
	}

	rotatedContent, err := os.ReadFile(rotated[0])
	if err != nil {
		t.Fatalf("read rotated file failed: %v", err)
	}
	if string(rotatedContent) != "first" {
		t.Fatalf("unexpected rotated content: %q", rotatedContent)
	}

	activeContent, err := os.ReadFile(filepath.Join(dir, "key.txt"))
	if err != nil {
		t.Fatalf("read active file failed: %v", err)
	}
	if string(activeContent) != "second" {
		t.Fatalf("unexpected active content: %q", activeContent)
	}
}

func TestRotateFile_rotateOnTime(t *testing.T) {
	dir := t.TempDir()
	result := NewRotateFile(dir, "time", "txt", 0, 1)
	if result.HasError() {
		t.Fatalf("NewRotateFile failed: %v", result.Error())
	}

	rotateFile := result.Data()
	defer rotateFile.Close()

	if err := rotateFile.WriteString("", "timed"); err != nil {
		t.Fatalf("WriteString failed: %v", err)
	}

	time.Sleep(1100 * time.Millisecond)

	if err := rotateFile.WriteString("", "after"); err != nil {
		t.Fatalf("second WriteString failed: %v", err)
	}

	rotated, err := rotateFile.RotatedFilePaths()
	if err != nil {
		t.Fatalf("RotatedFilePaths failed: %v", err)
	}
	if len(rotated) != 1 {
		t.Fatalf("expected 1 rotated file, got %d", len(rotated))
	}

	rotatedContent, err := os.ReadFile(rotated[0])
	if err != nil {
		t.Fatalf("read rotated file failed: %v", err)
	}
	if string(rotatedContent) != "timed" {
		t.Fatalf("unexpected rotated content: %q", rotatedContent)
	}
}

func TestRotateFile_writeAfterClose(t *testing.T) {
	dir := t.TempDir()
	result := NewRotateFile(dir, "closed", "txt", 0, 0)
	if result.HasError() {
		t.Fatalf("NewRotateFile failed: %v", result.Error())
	}

	rotateFile := result.Data()
	if err := rotateFile.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	if err := rotateFile.WriteString("", "data"); err == nil {
		t.Fatalf("expected error after close")
	}
}
