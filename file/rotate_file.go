// Package file provides file utilities including size- and time-based log rotation.
package file

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	. "github.com/Jordanzuo/stdx/result"
)

var (
	errRotateFileClosed = errors.New("rotate_file: closed")
)

// RotateModeKind selects whether rotation is driven by calendar day or elapsed time.
type RotateModeKind int

const (
	// RotateModeDateOrSize rotates when the calendar day changes or the file exceeds max size.
	RotateModeDateOrSize RotateModeKind = iota
	// RotateModeTimeOrSize rotates when the file age exceeds an interval or max size.
	RotateModeTimeOrSize
)

// RotateMode configures when the active file is rolled to a timestamped backup.
//
// Use RotateByDateOrSize or RotateByTimeOrSize to construct a mode. A zero maxSizeBytes
// disables size-based rotation; a zero intervalSeconds disables time-based rotation in
// RotateModeTimeOrSize.
type RotateMode struct {
	kind            RotateModeKind
	maxSizeBytes    int64
	intervalSeconds int64
}

// RotateByDateOrSize returns a mode that rotates on calendar-day change or when the active
// file exceeds maxSizeBytes. When maxSizeBytes is zero, only the date triggers rotation.
func RotateByDateOrSize(maxSizeBytes int64) RotateMode {
	return RotateMode{
		kind:         RotateModeDateOrSize,
		maxSizeBytes: maxSizeBytes,
	}
}

// RotateByTimeOrSize returns a mode that rotates when the active file is older than
// intervalSeconds or exceeds maxSizeBytes. When maxSizeBytes is zero, only time triggers
// rotation; when intervalSeconds is zero, only size triggers rotation.
func RotateByTimeOrSize(maxSizeBytes int64, intervalSeconds int64) RotateMode {
	return RotateMode{
		kind:            RotateModeTimeOrSize,
		maxSizeBytes:    maxSizeBytes,
		intervalSeconds: intervalSeconds,
	}
}

// RotateFile appends to a main file and rotates it to a timestamped backup when configured
// limits are reached. Rotation renames the main file to
// {mainFilename}_YYYYMMDDHHmmss{fileSuffix} and starts a new empty main file. When the
// active file has no written bytes, rotation limits may reset without creating a backup.
// Calendar-day and interval limits are also checked on a background schedule between writes.
//
// RotateFile is safe for concurrent use.
type RotateFile struct {
	directoryPath string
	mainFilename  string
	fileSuffix    string
	mode          RotateMode

	mu           sync.Mutex
	file         *os.File
	currentSize  int64
	activeDay    string
	fileEpoch    time.Time
	closed       bool
	stopSchedule chan struct{}
}

// NewRotateFile opens or creates the main file under directoryPath. directoryPath is created
// when it does not exist. mainFilename is the base name without suffix; fileSuffix is
// appended to both the main and rotated names (for example ".log").
func NewRotateFile(directoryPath, mainFilename, fileSuffix string, mode RotateMode) Result[*RotateFile] {
	type ResultDataType = *RotateFile

	if strings.TrimSpace(directoryPath) == "" {
		return NewErrorResult[ResultDataType](fmt.Errorf("rotate_file: empty directory path"))
	}
	if strings.TrimSpace(mainFilename) == "" {
		return NewErrorResult[ResultDataType](fmt.Errorf("rotate_file: empty main filename"))
	}
	if mode.maxSizeBytes < 0 {
		return NewErrorResult[ResultDataType](fmt.Errorf("rotate_file: negative max size"))
	}
	if mode.intervalSeconds < 0 {
		return NewErrorResult[ResultDataType](fmt.Errorf("rotate_file: negative interval"))
	}
	if mode.kind != RotateModeDateOrSize && mode.kind != RotateModeTimeOrSize {
		return NewErrorResult[ResultDataType](fmt.Errorf("rotate_file: invalid rotate mode"))
	}

	if err := os.MkdirAll(directoryPath, 0o755); err != nil {
		return NewErrorResult[ResultDataType](fmt.Errorf("rotate_file: create directory: %w", err))
	}

	rotateFile := &RotateFile{
		directoryPath: directoryPath,
		mainFilename:  mainFilename,
		fileSuffix:    fileSuffix,
		mode:          mode,
	}

	mainPath := rotateFile.mainPath()
	info, statErr := os.Stat(mainPath)
	now := time.Now()
	if statErr == nil {
		rotateFile.currentSize = info.Size()
		rotateFile.activeDay = info.ModTime().In(time.Local).Format("20060102")
		rotateFile.fileEpoch = info.ModTime()
	} else if os.IsNotExist(statErr) {
		rotateFile.activeDay = now.Format("20060102")
		rotateFile.fileEpoch = now
	} else {
		return NewErrorResult[ResultDataType](fmt.Errorf("rotate_file: stat main file: %w", statErr))
	}

	file, err := os.OpenFile(mainPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return NewErrorResult[ResultDataType](fmt.Errorf("rotate_file: open main file: %w", err))
	}
	rotateFile.file = file
	rotateFile.stopSchedule = make(chan struct{})
	if rotateFile.needsScheduleWatch() {
		go rotateFile.watchSchedule()
	}

	return NewDataResult(rotateFile)
}

// WriteBytes appends message to the active file, rotating first when limits require it.
func (this *RotateFile) WriteBytes(message []byte) error {
	this.mu.Lock()
	defer this.mu.Unlock()

	if this.closed {
		return errRotateFileClosed
	}

	now := time.Now()
	if err := this.maybeRotateLocked(now, len(message)); err != nil {
		return err
	}

	n, err := this.file.Write(message)
	this.currentSize += int64(n)
	if err != nil {
		return fmt.Errorf("rotate_file: write: %w", err)
	}
	return nil
}

// WriteString appends message to the active file, rotating first when limits require it.
func (this *RotateFile) WriteString(message string) error {
	return this.WriteBytes([]byte(message))
}

// GetFileList returns paths of rotated backup files in the directory. The active main file
// is omitted. Paths are sorted lexicographically.
func (this *RotateFile) GetFileList() Result[[]string] {
	type ResultDataType = []string

	this.mu.Lock()
	dir := this.directoryPath
	mainName := this.mainFilename
	suffix := this.fileSuffix
	this.mu.Unlock()

	entries, err := os.ReadDir(dir)
	if err != nil {
		return NewErrorResult[ResultDataType](fmt.Errorf("rotate_file: read directory: %w", err))
	}

	prefix := mainName + "_"
	var paths []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if name == mainName+suffix {
			continue
		}
		if !strings.HasPrefix(name, prefix) || !strings.HasSuffix(name, suffix) {
			continue
		}
		stamp := strings.TrimSuffix(strings.TrimPrefix(name, prefix), suffix)
		if len(stamp) != 14 || !isAllDigits(stamp) {
			continue
		}
		paths = append(paths, filepath.Join(dir, name))
	}

	sort.Strings(paths)
	return NewDataResult(paths)
}

// Close rotates the main file when it contains data, then closes the active file handle.
// Further writes return errRotateFileClosed.
func (this *RotateFile) Close() error {
	this.mu.Lock()
	defer this.mu.Unlock()

	if this.closed {
		return nil
	}
	if this.file == nil {
		this.stopScheduleWatchLocked()
		this.closed = true
		return nil
	}

	this.stopScheduleWatchLocked()

	if this.currentSize > 0 {
		if err := this.rotateLocked(time.Now()); err != nil {
			return err
		}
	}

	err := this.file.Close()
	this.file = nil
	this.closed = true
	if err != nil {
		return fmt.Errorf("rotate_file: close: %w", err)
	}
	return nil
}

func (this *RotateFile) mainPath() string {
	return filepath.Join(this.directoryPath, this.mainFilename+this.fileSuffix)
}

func (this *RotateFile) rotatedPath(now time.Time) string {
	stamp := now.Format("20060102150405")
	name := this.mainFilename + "_" + stamp + this.fileSuffix
	return filepath.Join(this.directoryPath, name)
}

func (this *RotateFile) needsScheduleWatch() bool {
	switch this.mode.kind {
	case RotateModeDateOrSize:
		return true
	case RotateModeTimeOrSize:
		return this.mode.intervalSeconds > 0
	default:
		return false
	}
}

func (this *RotateFile) watchSchedule() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-this.stopSchedule:
			return
		case now := <-ticker.C:
			this.mu.Lock()
			if this.closed {
				this.mu.Unlock()
				return
			}
			if this.shouldRotateForScheduleLocked(now) {
				_ = this.rotateLocked(now)
			}
			this.mu.Unlock()
		}
	}
}

func (this *RotateFile) stopScheduleWatchLocked() {
	if this.stopSchedule == nil {
		return
	}
	close(this.stopSchedule)
	this.stopSchedule = nil
}

func (this *RotateFile) maybeRotateLocked(now time.Time, nextWriteLen int) error {
	if !this.shouldRotateLocked(now, nextWriteLen) {
		return nil
	}
	return this.rotateLocked(now)
}

func (this *RotateFile) shouldRotateLocked(now time.Time, nextWriteLen int) bool {
	if this.shouldRotateForSizeLocked(nextWriteLen) {
		return true
	}
	return this.shouldRotateForScheduleLocked(now)
}

func (this *RotateFile) shouldRotateForSizeLocked(nextWriteLen int) bool {
	return this.mode.maxSizeBytes > 0 && this.currentSize+int64(nextWriteLen) > this.mode.maxSizeBytes
}

func (this *RotateFile) shouldRotateForScheduleLocked(now time.Time) bool {
	switch this.mode.kind {
	case RotateModeDateOrSize:
		return now.Format("20060102") != this.activeDay
	case RotateModeTimeOrSize:
		if this.mode.intervalSeconds <= 0 {
			return false
		}
		return now.Sub(this.fileEpoch) >= time.Duration(this.mode.intervalSeconds)*time.Second
	default:
		return false
	}
}

func (this *RotateFile) rotateLocked(now time.Time) error {
	if this.currentSize == 0 {
		this.activeDay = now.Format("20060102")
		this.fileEpoch = now
		return nil
	}

	mainPath := this.mainPath()
	if this.file != nil {
		if err := this.file.Close(); err != nil {
			return fmt.Errorf("rotate_file: close before rotate: %w", err)
		}
	}

	rotatedPath := this.rotatedPath(now)
	if err := os.Rename(mainPath, rotatedPath); err != nil {
		return fmt.Errorf("rotate_file: rename: %w", err)
	}

	file, err := os.OpenFile(mainPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("rotate_file: open after rotate: %w", err)
	}

	this.file = file
	this.currentSize = 0
	this.activeDay = now.Format("20060102")
	this.fileEpoch = now
	return nil
}

func isAllDigits(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}
