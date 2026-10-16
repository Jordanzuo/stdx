package file

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestRotateFile_writeAndRotateBySize(t *testing.T) {
	dir := t.TempDir()
	rf := NewRotateFile(dir, "app", ".log", RotateByDateOrSize(5)).Data()
	defer rf.Close()

	if err := rf.WriteString("12345"); err != nil {
		t.Fatal(err)
	}
	if err := rf.WriteString("6"); err != nil {
		t.Fatal(err)
	}

	list := rf.GetFileList().Data()
	if len(list) != 1 {
		t.Fatalf("file list: got %d paths, want 1: %v", len(list), list)
	}

	mainPath := filepath.Join(dir, "app.log")
	data, err := os.ReadFile(mainPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "6" {
		t.Fatalf("main file: got %q, want %q", data, "6")
	}

	rotated := list[0]
	rotatedData, err := os.ReadFile(rotated)
	if err != nil {
		t.Fatal(err)
	}
	if string(rotatedData) != "12345" {
		t.Fatalf("rotated file: got %q, want %q", rotatedData, "12345")
	}
	base := filepath.Base(rotated)
	if !strings.HasPrefix(base, "app_") || !strings.HasSuffix(base, ".log") {
		t.Fatalf("rotated name: got %q", base)
	}
	stamp := strings.TrimSuffix(strings.TrimPrefix(base, "app_"), ".log")
	if len(stamp) != 14 {
		t.Fatalf("timestamp length: got %d", len(stamp))
	}
}

func TestRotateFile_rotateSkippedWhenEmpty(t *testing.T) {
	dir := t.TempDir()
	rf := NewRotateFile(dir, "empty", ".log", RotateByDateOrSize(0)).Data()
	defer rf.Close()

	rf.activeDay = "20000101"
	if err := rf.WriteString("first"); err != nil {
		t.Fatal(err)
	}

	list := rf.GetFileList().Data()
	if len(list) != 0 {
		t.Fatalf("rotated files: got %v", list)
	}

	data, err := os.ReadFile(filepath.Join(dir, "empty.log"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "first" {
		t.Fatalf("main file: got %q, want first", data)
	}
}

func TestRotateFile_rotateByDate(t *testing.T) {
	dir := t.TempDir()
	rf := NewRotateFile(dir, "daily", ".txt", RotateByDateOrSize(0)).Data()
	defer rf.Close()

	if err := rf.WriteString("old"); err != nil {
		t.Fatal(err)
	}
	rf.activeDay = "20000101"
	if err := rf.WriteString("x"); err != nil {
		t.Fatal(err)
	}

	list := rf.GetFileList().Data()
	if len(list) != 1 {
		t.Fatalf("rotated files: got %v", list)
	}
}

func TestRotateFile_activeScheduleRotationWithoutWrite(t *testing.T) {
	dir := t.TempDir()
	rf := NewRotateFile(dir, "idle", ".log", RotateByTimeOrSize(0, 1)).Data()

	if err := rf.WriteString("idle"); err != nil {
		t.Fatal(err)
	}

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if list := rf.GetFileList().Data(); len(list) == 1 {
			rotatedData, err := os.ReadFile(list[0])
			if err != nil {
				t.Fatal(err)
			}
			if string(rotatedData) != "idle" {
				t.Fatalf("rotated file: got %q, want idle", rotatedData)
			}
			if err := rf.Close(); err != nil {
				t.Fatal(err)
			}
			return
		}
		time.Sleep(50 * time.Millisecond)
	}

	t.Fatal("timed out waiting for schedule rotation")
}

func TestRotateFile_rotateByTime(t *testing.T) {
	dir := t.TempDir()
	rf := NewRotateFile(dir, "timed", ".log", RotateByTimeOrSize(0, 1)).Data()
	defer rf.Close()

	if err := rf.WriteString("old"); err != nil {
		t.Fatal(err)
	}
	rf.fileEpoch = time.Now().Add(-2 * time.Second)
	if err := rf.WriteString("a"); err != nil {
		t.Fatal(err)
	}

	list := rf.GetFileList().Data()
	if len(list) != 1 {
		t.Fatalf("file list: got %v", list)
	}
}

func TestRotateFile_getFileListExcludesMain(t *testing.T) {
	dir := t.TempDir()
	rf := NewRotateFile(dir, "list", ".log", RotateByDateOrSize(3)).Data()
	defer rf.Close()

	if err := rf.WriteString("abc"); err != nil {
		t.Fatal(err)
	}
	if err := rf.WriteString("d"); err != nil {
		t.Fatal(err)
	}

	list := rf.GetFileList().Data()
	if len(list) != 1 {
		t.Fatalf("rotated files: got %v", list)
	}
	mainPath := filepath.Join(dir, "list.log")
	for _, p := range list {
		if p == mainPath {
			t.Fatalf("main file must not appear in list: %v", list)
		}
	}
}

func TestRotateFile_concurrentWrites(t *testing.T) {
	dir := t.TempDir()
	rf := NewRotateFile(dir, "conc", ".log", RotateByDateOrSize(1<<20)).Data()
	defer rf.Close()

	const n = 50
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			_ = rf.WriteString("x")
		}(i)
	}
	wg.Wait()

	data, err := os.ReadFile(filepath.Join(dir, "conc.log"))
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != n {
		t.Fatalf("bytes written: got %d, want %d", len(data), n)
	}
}

func TestRotateFile_close(t *testing.T) {
	dir := t.TempDir()
	rf := NewRotateFile(dir, "closed", ".log", RotateByDateOrSize(100)).Data()

	if err := rf.Close(); err != nil {
		t.Fatal(err)
	}
	if err := rf.WriteBytes([]byte("x")); err != errRotateFileClosed {
		t.Fatalf("WriteBytes after close: got %v, want %v", err, errRotateFileClosed)
	}
	if err := rf.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestRotateFile_closeRotatesNonEmptyMain(t *testing.T) {
	dir := t.TempDir()
	rf := NewRotateFile(dir, "shutdown", ".log", RotateByDateOrSize(1<<20)).Data()

	if err := rf.WriteString("payload"); err != nil {
		t.Fatal(err)
	}
	if err := rf.Close(); err != nil {
		t.Fatal(err)
	}

	list := rf.GetFileList().Data()
	if len(list) != 1 {
		t.Fatalf("rotated files: got %v", list)
	}
	rotatedData, err := os.ReadFile(list[0])
	if err != nil {
		t.Fatal(err)
	}
	if string(rotatedData) != "payload" {
		t.Fatalf("rotated file: got %q, want payload", rotatedData)
	}

	mainData, err := os.ReadFile(filepath.Join(dir, "shutdown.log"))
	if err != nil {
		t.Fatal(err)
	}
	if len(mainData) != 0 {
		t.Fatalf("main file after close: got %q, want empty", mainData)
	}
}

func TestRotateFile_closeSkipsRotateWhenEmpty(t *testing.T) {
	dir := t.TempDir()
	rf := NewRotateFile(dir, "emptyclose", ".log", RotateByDateOrSize(1)).Data()

	if err := rf.Close(); err != nil {
		t.Fatal(err)
	}
	if list := rf.GetFileList().Data(); len(list) != 0 {
		t.Fatalf("rotated files: got %v", list)
	}
}

func TestNewRotateFile_validation(t *testing.T) {
	dir := t.TempDir()
	if r := NewRotateFile("", "a", ".log", RotateByDateOrSize(1)); !r.HasError() {
		t.Fatal("expected error for empty directory")
	}
	if r := NewRotateFile(dir, "", ".log", RotateByDateOrSize(1)); !r.HasError() {
		t.Fatal("expected error for empty filename")
	}
	if r := NewRotateFile(dir, "a", ".log", RotateByDateOrSize(-1)); !r.HasError() {
		t.Fatal("expected error for negative size")
	}
}
