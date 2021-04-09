package logging

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/dunelang/dune/filesystem"
)

const DATE_LAYOUT = "2006-01-02 15:04:05"
const LAYOUT_LEN = len(DATE_LAYOUT)

type Logger struct {
	Path      string
	mutex     *sync.RWMutex
	fs        filesystem.FS
	file      filesystem.File
	writePath string
}

func New(path string, fs filesystem.FS) *Logger {
	return &Logger{
		Path:  path,
		mutex: &sync.RWMutex{},
		fs:    fs,
	}
}

func (db *Logger) Close() error {
	if db.file == nil {
		return nil
	}
	return db.file.Close()
}

func (db *Logger) Save(table, data string, v ...interface{}) error {
	return db.save(time.Now(), table, data, v...)
}

type DataPoint struct {
	Time time.Time
	Text string
}

func (d DataPoint) String() string {
	return d.Time.Format(DATE_LAYOUT) + " " + d.Text
}

type Scanner struct {
	reader  *reader
	scanner *bufio.Scanner
	Error   error
}

func (s *Scanner) Scan() bool {
	r := s.reader
	sc := s.scanner

LOOP:
	for {
		if r.limit > 0 && r.index >= r.limit {
			return false
		}

		if ok := sc.Scan(); !ok {
			return false
		}

		d := s.Data()
		// advance to start before sending data
		if d.Time.Before(r.start) {
			continue LOOP
		}

		if r.filter != "" {
			if !strings.Contains(sc.Text(), r.filter) {
				continue LOOP
			}
		}

		// advance to Offset before sending data
		for r.index < r.offset {
			r.index++
			continue LOOP
		}

		r.index++
		return true
	}
}

func (s *Scanner) Close() {
	s.reader.Close()
}

func (s *Scanner) SetFilter(v string) {
	s.reader.filter = v
}

func (s *Scanner) Data() DataPoint {
	line := s.scanner.Text()

	err := s.scanner.Err()
	if err != nil {
		s.Error = err
		return DataPoint{}
	}

	d, err := time.Parse(DATE_LAYOUT, line[:LAYOUT_LEN])
	if err != nil {
		s.Error = fmt.Errorf("error parsing time in '%s': %w", line, err)
		return DataPoint{}
	}

	return DataPoint{Time: d, Text: line[LAYOUT_LEN+1:]}
}

func (db *Logger) Query(table string, start, end time.Time, offset, size int) *Scanner {
	r := db.reader(start, end, table, offset, offset+size)
	s := bufio.NewScanner(r)

	// set large capacity (some lines ar very long)
	const maxCapacity = 512 * 1024
	buf := make([]byte, maxCapacity)
	s.Buffer(buf, maxCapacity)

	return &Scanner{
		scanner: s,
		reader:  r,
	}
}

type reader struct {
	db       *Logger
	table    string
	start    time.Time
	end      time.Time
	offset   int
	limit    int
	index    int
	filter   string
	current  time.Time
	file     *os.File
	keepFile bool
	buf      []byte
}

// Read reads up to len(p) bytes through one or many files
func (r *reader) Read(p []byte) (int, error) {
	r.db.mutex.RLock()
	defer r.db.mutex.RUnlock()

	l := len(p)

	for {
		// if there is enough data, send it
		if len(r.buf) >= l {
			copy(p, r.buf[:l])
			r.buf = r.buf[l:]
			return l, nil
		}

		// advance to the next file in necessary
		if !r.keepFile {
			err := r.nextFile()
			if err != nil {
				var n int
				if err == io.EOF {
					// Last file reached. Copy the remaining data
					if len(r.buf) > 0 {
						n = copy(p, r.buf)
					}
				}
				return n, err
			}
		}

		// read the current file and grow the buffer
		b := make([]byte, len(p))
		n, err := r.file.Read(b)
		if err == io.EOF {
			r.keepFile = false
		} else if err != nil {
			return n, err
		} else {
			r.keepFile = true
		}

		r.buf = append(r.buf, b[:n]...)
	}
}

func (r *reader) nextFile() error {
	for {
		// poner al inicio o avanzar un día
		if r.current.Before(r.start) {
			r.current = r.start
		} else {
			r.current = r.current.Add(time.Hour * 24)
		}

		// controlar si nos  hemos pasado de fecha
		if r.current.After(r.end) {
			return io.EOF
		}

		file, err := r.open(r.current)
		if err != nil {
			// si este día no hay datos pasar al siguiente
			if os.IsNotExist(err) {
				continue
			}
			return err
		}

		r.file = file
		return nil
	}
}

func (r *reader) Close() {
	if r.file != nil {
		r.file.Close()
		r.file = nil
	}
}

func (r *reader) open(t time.Time) (*os.File, error) {
	// Close the previous one if exists
	r.Close()

	path := r.db.getTablePath(t, r.table)

	f, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, err
		}
		return nil, fmt.Errorf("logDB.open: error openning file %s: %w", path, err)
	}
	return f, nil
}

func (db *Logger) reader(start, end time.Time, table string, offset, limit int) *reader {
	return &reader{
		db:     db,
		start:  start.Local(),
		end:    end.Local(),
		table:  table,
		offset: offset,
		limit:  limit,
	}
}

func (db *Logger) getDir(t time.Time) string {
	return filepath.Join(db.Path, t.Format("2006-01-02"))
}

func (db *Logger) getTablePath(t time.Time, table string) string {
	return filepath.Join(db.getDir(t), table+".log")
}

func (db *Logger) save(t time.Time, table, data string, v ...interface{}) error {
	if len(v) > 0 {
		data = fmt.Sprintf(data, v...)
	}

	dirName := db.getDir(t)
	fileName := db.getTablePath(t, table)

	db.mutex.Lock()

	if db.file == nil || db.writePath != fileName {
		if db.file != nil {
			db.file.Close()
			db.file = nil
		}

		fs := db.fs

		err := fs.MkdirAll(dirName)
		if err != nil {
			return fmt.Errorf("logDB: error creating dir %s: %w", dirName, err)
		}

		f, err := fs.OpenForAppend(fileName) //.OpenFile(fileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			return fmt.Errorf("logDB.Save: error openning file %s: %w", fileName, err)
		}

		db.file = f
		db.writePath = fileName
	}

	if _, err := fmt.Fprintf(db.file, "%s %s\n", t.Format(DATE_LAYOUT), data); err != nil {
		return fmt.Errorf("logDB: error writing data %w", err)
	}

	db.mutex.Unlock()
	return nil
}
