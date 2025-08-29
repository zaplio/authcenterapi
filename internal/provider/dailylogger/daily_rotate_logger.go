package dailylogger

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	backupTimeFormat = "2006-01-02"
	compressSuffix   = ".gz"
	defaultMaxSize   = 100
)

// Modified from https://github.com/natefinch/lumberjack
type DailyRotateLogger struct {
	fileName   string
	maxSize    int
	maxBackups int
	maxAge     int
	localTime  bool
	compress   bool

	size        int64
	backupCount int
	day         time.Time
	isChangeDay bool
	file        *os.File
	mfile       sync.Mutex

	millCh    chan bool
	startMill sync.Once
}

var (
	// currentTime exists so it can be mocked out by tests.
	currentTime = time.Now

	// megabyte is the conversion factor between MaxSize and bytes.  It is a
	// variable so tests can mock it out and not need to write megabytes of data
	// to disk.
	megabyte = 1024 * 1024
)

func NewDailyRotateLogger(fileName string, maxSize, maxBackups, maxAge int, localTime bool, compress bool) *DailyRotateLogger {
	l := DailyRotateLogger{fileName: fileName, maxSize: maxSize, maxBackups: maxBackups, maxAge: maxAge, localTime: localTime, compress: compress}
	l.setDay()
	return &l
}

func (l *DailyRotateLogger) Write(p []byte) (int, error) {
	l.mfile.Lock()
	defer l.mfile.Unlock()

	writeLen := int64(len(p))
	if writeLen > l.max() {
		return 0, fmt.Errorf(
			"write length %d exceeds maximum file size %d", writeLen, l.max(),
		)
	}

	if l.file == nil {
		if err := l.openExistingOrNew(len(p)); err != nil {
			return 0, err
		}
	}

	if l.size+writeLen > l.max() {
		if err := l.rotate(); err != nil {
			return 0, err
		}
	}

	now := currentTime()
	if !l.localTime {
		now = now.UTC()
	}

	if now.After(l.day) {
		l.isChangeDay = true
		if err := l.rotate(); err != nil {
			l.backupCount = 0
			return 0, err
		}
	}

	n, err := l.file.Write(p)
	l.size += int64(n)
	return n, err
}

func (l *DailyRotateLogger) Close() error {
	l.mfile.Lock()
	defer l.mfile.Unlock()
	return l.close()
}

func (l *DailyRotateLogger) close() error {
	if l.file == nil {
		return nil
	}

	err := l.file.Close()
	l.file = nil
	return err
}

func (l *DailyRotateLogger) rotate() error {
	if err := l.close(); err != nil {
		return err
	}
	if err := l.openNew(); err != nil {
		return err
	}
	l.setDay()
	l.mill()
	return nil
}

func (l *DailyRotateLogger) openNew() error {
	err := os.MkdirAll(l.dir(), 0744)
	if err != nil {
		return fmt.Errorf("cannot make directories for new logfile: %s", err)
	}

	name := l.filename()
	mode := os.FileMode(0644)
	info, err := os.Stat(name)
	if err == nil {
		var newname string
		if l.isChangeDay {
			l.backupCount = 0
			files, err := l.oldLogFiles()
			if err != nil {
				return err
			}
			yesterdayCountBackup := 1

			yesterday := currentTime().Add(time.Hour * -24)
			if !l.localTime {
				yesterday = yesterday.UTC()
			}

			prefix, _ := l.prefixAndExt()

			for _, f := range files {
				if strings.HasPrefix(f.Name(), prefix) && strings.Contains(f.Name(), yesterday.Format(backupTimeFormat)) {
					yesterdayCountBackup += 1
				}
			}
			newname = backupName(name, l.localTime, yesterdayCountBackup, l.isChangeDay)
			l.isChangeDay = false
		} else {
			l.backupCount += 1
			newname = backupName(name, l.localTime, l.backupCount, l.isChangeDay)
		}
		mode = info.Mode()
		if err := os.Rename(name, newname); err != nil {
			return fmt.Errorf("cannot rename log file: %s", err)
		}

		if err := chown(name, info); err != nil {
			return err
		}
	}

	f, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("cannot open new logfile: %s", err)
	}

	l.file = f
	l.size = 0
	return nil
}

func backupName(name string, local bool, backup int, isChangeDay bool) string {
	dir := filepath.Dir(name)
	filename := filepath.Base(name)
	ext := filepath.Ext(filename)
	prefix := filename[:len(filename)-len(ext)]

	t := currentTime()
	if !local {
		t = t.UTC()
	}

	if isChangeDay {
		t = t.Add(time.Hour * -24)
	}

	timestamp := t.Format(backupTimeFormat)
	return filepath.Join(dir, fmt.Sprintf("%s.%s.%d%s", prefix, timestamp, backup, ext))
}

func (l *DailyRotateLogger) openExistingOrNew(writeLen int) error {
	err := os.MkdirAll(l.dir(), 0744)
	if err != nil {
		return fmt.Errorf("cannot make directories for new logfile: %s", err)
	}

	files, err := l.oldLogFiles()
	if err != nil {
		return err
	}

	now := currentTime()
	if !l.localTime {
		now = now.UTC()
	}

	prefix, _ := l.prefixAndExt()

	for _, f := range files {
		if strings.HasPrefix(f.Name(), prefix) && strings.Contains(f.Name(), now.Format(backupTimeFormat)) {
			l.backupCount += 1
		}
	}

	l.mill()

	filename := l.filename()
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return l.openNew()
	}

	if err != nil {
		return fmt.Errorf("error getting log file info: %s", err)
	}

	if getDate(now, l.localTime).After(getDate(info.ModTime(), l.localTime)) {
		l.isChangeDay = true
		l.backupCount = 0
		return l.rotate()
	}

	if now.After(l.day) {
		l.backupCount = 0
		return l.rotate()
	}

	if info.Size()+int64(writeLen) >= l.max() {
		return l.rotate()
	}

	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return l.openNew()
	}
	l.file = file
	l.size = info.Size()
	return nil
}

func (l *DailyRotateLogger) filename() string {
	if l.fileName != "" {
		return l.fileName
	}
	name := filepath.Base(os.Args[0]) + ".log"
	return filepath.Join(os.TempDir(), name)
}

// millRunOnce performs compression and removal of stale log files.
// Log files are compressed if enabled via configuration and old log
// files are removed, keeping at most l.maxBackups files, as long as
// none of them are older than MaxAge.
func (l *DailyRotateLogger) millRunOnce() error {
	if l.maxBackups == 0 && l.maxAge == 0 && !l.compress {
		return nil
	}

	files, err := l.oldLogFiles()
	if err != nil {
		return err
	}

	var compress, remove []logInfo

	if l.maxBackups > 0 && l.maxBackups < len(files) {
		preserved := make(map[string]bool)
		var remaining []logInfo
		for _, f := range files {
			// Only count the uncompressed log file or the
			// compressed log file, not both.
			fn := f.Name()
			fn = strings.TrimSuffix(fn, compressSuffix)
			// if strings.HasSuffix(fn, compressSuffix) {
			// 	fn = fn[:len(fn)-len(compressSuffix)]
			// }
			preserved[fn] = true

			if len(preserved) > l.maxBackups {
				remove = append(remove, f)
			} else {
				remaining = append(remaining, f)
			}
		}
		files = remaining
	}

	if l.maxAge > 0 {
		diff := time.Duration(int64(24*time.Hour) * int64(l.maxAge))
		cutoff := currentTime().Add(-1 * diff)

		var remaining []logInfo
		for _, f := range files {
			if f.timestamp.Before(cutoff) {
				remove = append(remove, f)
			} else {
				remaining = append(remaining, f)
			}
		}
		files = remaining
	}

	if l.compress {
		for _, f := range files {
			if !strings.HasSuffix(f.Name(), compressSuffix) {
				compress = append(compress, f)
			}
		}
	}

	for _, f := range remove {
		errRemove := os.Remove(filepath.Join(l.dir(), f.Name()))
		if err == nil && errRemove != nil {
			err = errRemove
		}
	}

	for _, f := range compress {
		fn := filepath.Join(l.dir(), f.Name())
		errCompress := compressLogFile(fn, fmt.Sprintf("%s%s", fn, compressSuffix))
		if err == nil && errCompress != nil {
			err = errCompress
		}
	}

	return err
}

func (l *DailyRotateLogger) millRun() {
	for range l.millCh {
		_ = l.millRunOnce()
	}
}

func (l *DailyRotateLogger) mill() {
	l.startMill.Do(func() {
		l.millCh = make(chan bool, 1)
		go l.millRun()
	})

	select {
	case l.millCh <- true:
	default:
	}
}

func (l *DailyRotateLogger) oldLogFiles() ([]logInfo, error) {
	files, err := ioutil.ReadDir(l.dir())
	if err != nil {
		return nil, fmt.Errorf("cannot read log file directory: %s", err)
	}

	var logFiles []logInfo

	prefix, ext := l.prefixAndExt()

	for _, f := range files {
		if f.IsDir() {
			continue
		}
		if t, err := l.timeFromName(f.Name(), prefix, ext); err == nil {
			logFiles = append(logFiles, logInfo{t, f})
			continue
		}
		if t, err := l.timeFromName(f.Name(), prefix, ext+compressSuffix); err == nil {
			logFiles = append(logFiles, logInfo{t, f})
			continue
		}
	}

	sort.Sort(byFormatTime(logFiles))

	return logFiles, nil
}

func (l *DailyRotateLogger) timeFromName(filename, prefix, ext string) (time.Time, error) {
	if !strings.HasPrefix(filename, prefix) {
		return time.Time{}, errors.New("mismatched prefix")
	}

	if !strings.HasSuffix(filename, ext) {
		return time.Time{}, errors.New("mismatched extension")
	}

	if len(filename)-len(ext) <= len(prefix) {
		return time.Time{}, errors.New("this is active log file")
	}
	ts := filename[len(prefix) : len(filename)-len(ext)]
	if strings.Contains(ts, ".") {
		ts = strings.Split(ts, ".")[0]
	}
	return time.Parse(backupTimeFormat, ts)
}

// max returns the maximum size in bytes of log files before rolling.
func (l *DailyRotateLogger) max() int64 {
	if l.maxSize == 0 {
		return int64(defaultMaxSize * megabyte)
	}
	return int64(l.maxSize) * int64(megabyte)
}

func (l *DailyRotateLogger) dir() string { return filepath.Dir(l.filename()) }

func (l *DailyRotateLogger) prefixAndExt() (prefix, ext string) {
	filename := filepath.Base(l.filename())
	ext = filepath.Ext(filename)
	prefix = filename[:len(filename)-len(ext)] + "."
	return prefix, ext
}

func (l *DailyRotateLogger) setDay() {
	now := currentTime().Add(time.Hour * 24)
	if l.localTime {
		l.day = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	} else {
		l.day = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	}
}

func getDate(t time.Time, localTime bool) time.Time {
	if localTime {
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
	} else {
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	}
}

func compressLogFile(src, dst string) (err error) {
	f, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}
	defer f.Close()

	fi, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to stat log file: %v", err)
	}

	if err := chown(dst, fi); err != nil {
		return fmt.Errorf("failed to chown compressed log file: %v", err)
	}

	gzf, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, fi.Mode())
	if err != nil {
		return fmt.Errorf("failed to open compressed log file: %v", err)
	}
	defer gzf.Close()

	gz := gzip.NewWriter(gzf)

	defer func() {
		if err != nil {
			if e := os.Remove(dst); e != nil {
				err = e
			} else {
				err = fmt.Errorf("failed to compress log file: %v", err)
			}
		}
	}()

	if _, err := io.Copy(gz, f); err != nil {
		return err
	}
	if err := gz.Close(); err != nil {
		return err
	}
	if err := gzf.Close(); err != nil {
		return err
	}

	if err := f.Close(); err != nil {
		return err
	}
	if err := os.Remove(src); err != nil {
		return err
	}

	return nil
}

type logInfo struct {
	timestamp time.Time
	os.FileInfo
}

// byFormatTime sorts by newest time formatted in the name.
type byFormatTime []logInfo

func (b byFormatTime) Less(i, j int) bool { return b[i].timestamp.After(b[j].timestamp) }

func (b byFormatTime) Swap(i, j int) { b[i], b[j] = b[j], b[i] }

func (b byFormatTime) Len() int { return len(b) }

// ensure we always implement io.WriteCloser
var _ io.WriteCloser = (*DailyRotateLogger)(nil)
