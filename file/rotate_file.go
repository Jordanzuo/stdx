// Package file provides utilities for writing data to disk with file rotation.
//
// RotateFile writes to a single active file until a rotation condition is met:
// maximum file size, elapsed time interval, or a change in the data key supplied
// with each write. When rotation happens, the current file is synced, closed, and
// renamed with a timestamp suffix; a new active file is then opened.
//
// Create RotateFile values with NewRotateFile. A zero MaxFileSize or
// TimeIntervalSec disables that rotation trigger. Pass an empty data key to
// disable data-key rotation.
package file

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	. "github.com/Jordanzuo/stdx/result"
)

const rotateTimeLayout = "20060102_150405"
const rotateCheckInterval = 5 * time.Second

// RotateFile writes data to disk and rotates the active file when size, time, or
// data-key limits are reached.
type RotateFile struct {
	directoryName   string
	fileName        string
	fileSuffix      string
	timeIntervalSec int64
	maxTimeSec      int64
	currFileSize    int64
	maxFileSize     int64
	currDataKey     string
	hasDataKey      bool
	file                 *os.File
	mutex                sync.Mutex
	closingChannel       chan bool
	closedChannel        chan bool
	periodicCheckStarted bool
}

// NewRotateFile opens or creates the active file in directoryName.
//
// fileName and fileSuffix form the active file name as fileName.fileSuffix.
// maxFileSize is the size in bytes that triggers rotation; 0 disables size rotation.
// timeIntervalSec is the seconds after which rotation is triggered; 0 disables time rotation.
func NewRotateFile(directoryName, fileName, fileSuffix string, maxFileSize, timeIntervalSec int64) Result[*RotateFile] {
	type ResultDataType = *RotateFile

	if directoryName == "" {
		return NewErrorResult[ResultDataType](fmt.Errorf("file: empty directory name"))
	}
	if fileName == "" {
		return NewErrorResult[ResultDataType](fmt.Errorf("file: empty file name"))
	}
	if fileSuffix == "" {
		return NewErrorResult[ResultDataType](fmt.Errorf("file: empty file suffix"))
	}
	if maxFileSize < 0 {
		return NewErrorResult[ResultDataType](fmt.Errorf("file: negative max file size"))
	}
	if timeIntervalSec < 0 {
		return NewErrorResult[ResultDataType](fmt.Errorf("file: negative time interval"))
	}

	if err := os.MkdirAll(directoryName, 0o755); err != nil {
		return NewErrorResult[ResultDataType](fmt.Errorf("file: create directory %q: %w", directoryName, err))
	}

	rotateFile := &RotateFile{
		directoryName:   directoryName,
		fileName:        fileName,
		fileSuffix:      fileSuffix,
		timeIntervalSec: timeIntervalSec,
		maxFileSize:     maxFileSize,
		closingChannel:  make(chan bool, 1),
		closedChannel:   make(chan bool, 1),
	}

	if err := rotateFile.initFile(); err != nil {
		rotateFile.closeFile()
		return NewErrorResult[ResultDataType](err)
	}

	if timeIntervalSec > 0 {
		rotateFile.periodicCheckStarted = true
		go rotateFile.runPeriodicCheck()
	}

	return NewDataResult(rotateFile)
}

// WriteString appends message to the active file using dataKey for data-change rotation.
// When dataKey is empty, rotation happens only for size or time limits.
func (this *RotateFile) WriteString(dataKey, message string) error {
	return this.write(dataKey, []byte(message))
}

// WriteBytes appends data to the active file using dataKey for data-change rotation.
// When dataKey is empty, rotation happens only for size or time limits.
func (this *RotateFile) WriteBytes(dataKey string, data []byte) error {
	return this.write(dataKey, data)
}

// RotatedFilePaths returns paths of rotated files for this RotateFile, excluding the active file.
func (this *RotateFile) RotatedFilePaths() ([]string, error) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	return listRotatedFiles(this.directoryName, this.filePrefix(), this.fileSuffix)
}

// Close syncs and closes the active file and stops periodic rotation checks.
func (this *RotateFile) Close() error {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	var err error
	if this.file != nil {
		err = this.file.Sync()
		closeErr := this.file.Close()
		if err == nil {
			err = closeErr
		}
		this.file = nil
	}

	if this.periodicCheckStarted {
		this.closingChannel <- true
		<-this.closedChannel
		this.periodicCheckStarted = false
		this.closingChannel = nil
		this.closedChannel = nil
	}

	return err
}

func (this *RotateFile) write(dataKey string, data []byte) error {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	if this.file == nil {
		return fmt.Errorf("file: rotate file is closed")
	}

	if err := this.rotateIfNeeded(dataKey, int64(len(data)), false); err != nil {
		return err
	}

	n, err := this.file.Write(data)
	if err != nil {
		return fmt.Errorf("file: write to %s failed: %w", this.activeFilePath(), err)
	}

	this.currFileSize += int64(n)
	return this.rotateIfNeeded(dataKey, 0, true)
}

func (this *RotateFile) rotateIfNeeded(dataKey string, pendingSize int64, afterWrite bool) error {
	if !this.shouldRotate(dataKey, pendingSize, afterWrite) {
		if dataKey != "" {
			this.currDataKey = dataKey
			this.hasDataKey = true
		}
		return nil
	}

	return this.rotate(dataKey)
}

func (this *RotateFile) shouldRotate(dataKey string, pendingSize int64, afterWrite bool) bool {
	if this.maxFileSize > 0 {
		size := this.currFileSize
		if !afterWrite {
			size += pendingSize
		}
		if size >= this.maxFileSize {
			return true
		}
	}

	if this.timeIntervalSec > 0 && time.Now().Unix() >= this.maxTimeSec {
		return true
	}

	if dataKey != "" && this.hasDataKey && dataKey != this.currDataKey && this.currFileSize > 0 {
		return true
	}

	return false
}

func (this *RotateFile) rotate(nextDataKey string) error {
	if this.file == nil {
		return nil
	}

	if err := this.file.Sync(); err != nil {
		return fmt.Errorf("file: sync %s failed: %w", this.activeFilePath(), err)
	}

	if err := this.file.Close(); err != nil {
		return fmt.Errorf("file: close %s failed: %w", this.activeFilePath(), err)
	}
	this.file = nil

	if this.currFileSize > 0 {
		oldPath := this.activeFilePath()
		newPath := this.rotatedFilePath(time.Now())
		if err := os.Rename(oldPath, newPath); err != nil {
			return fmt.Errorf("file: rename %s to %s failed: %w", oldPath, newPath, err)
		}
	} else if err := os.Remove(this.activeFilePath()); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("file: remove empty active file %s failed: %w", this.activeFilePath(), err)
	}

	if err := this.initFile(); err != nil {
		return err
	}

	if nextDataKey != "" {
		this.currDataKey = nextDataKey
		this.hasDataKey = true
	} else {
		this.currDataKey = ""
		this.hasDataKey = false
	}

	return nil
}

func (this *RotateFile) initFile() error {
	this.maxTimeSec = time.Now().Unix() + this.timeIntervalSec
	this.currFileSize = 0

	file, err := os.OpenFile(this.activeFilePath(), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("file: open %s failed: %w", this.activeFilePath(), err)
	}

	this.file = file
	return nil
}

func (this *RotateFile) closeFile() {
	if this.file != nil {
		this.file.Close()
		this.file = nil
	}
}

func (this *RotateFile) runPeriodicCheck() {
	tick := time.NewTicker(rotateCheckInterval)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			this.mutex.Lock()
			err := this.rotateIfNeeded(this.currDataKey, 0, true)
			this.mutex.Unlock()
			if err != nil {
				return
			}
		case <-this.closingChannel:
			this.closedChannel <- true
			return
		}
	}
}

func (this *RotateFile) filePrefix() string {
	return this.fileName + "_"
}

func (this *RotateFile) activeFilePath() string {
	return filepath.Join(this.directoryName, fmt.Sprintf("%s.%s", this.fileName, this.fileSuffix))
}

func (this *RotateFile) rotatedFilePath(at time.Time) string {
	return filepath.Join(
		this.directoryName,
		fmt.Sprintf("%s_%s.%s", this.fileName, at.Format(rotateTimeLayout), this.fileSuffix),
	)
}

func listRotatedFiles(directoryName, prefix, suffix string) ([]string, error) {
	entries, err := os.ReadDir(directoryName)
	if err != nil {
		return nil, err
	}

	suffixWithDot := "." + suffix
	fileList := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		baseName := entry.Name()
		if prefix != "" && !strings.HasPrefix(baseName, prefix) {
			continue
		}
		if suffix != "" && !strings.HasSuffix(baseName, suffixWithDot) {
			continue
		}

		fileList = append(fileList, filepath.Join(directoryName, baseName))
	}

	return fileList, nil
}