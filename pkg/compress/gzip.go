package compress

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"sync/atomic"
)

// Gzip compresses the given file, removing the source log file if successful.
func Gzip(src string) (n int, err error) {
	f, err := os.Open(src)
	if err != nil {
		return n, fmt.Errorf("failed to open log file: %v", err)
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return n, fmt.Errorf("failed to stat log file: %v", err)
	}

	// If this file already exists, we presume it was created by
	// a previous attempt to compress the log file.
	gzf, err := os.OpenFile(src+".gz", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, fi.Mode())
	if err != nil {
		return n, fmt.Errorf("failed to open compressed log file: %v", err)
	}

	defer gzf.Close()

	counter := NewWriterCounter(gzf)
	gz := gzip.NewWriter(counter)
	defer gz.Close()

	if _, err := io.Copy(gz, f); err != nil {
		return n, err
	}

	_ = os.Remove(src)
	return int(counter.Count()), nil
}

// WriterCounter is counter for io.Writer
type WriterCounter struct {
	io.Writer
	count uint64
}

// NewWriterCounter function create new WriterCounter
func NewWriterCounter(w io.Writer) *WriterCounter {
	return &WriterCounter{
		Writer: w,
	}
}

func (counter *WriterCounter) Write(buf []byte) (int, error) {
	n, err := counter.Writer.Write(buf)

	// Write() should always return a non-negative `n`.
	// But since `n` is a signed integer, some custom
	// implementation of an io.Writer may return negative
	// values.
	//
	// Excluding such invalid values from counting,
	// thus `if n >= 0`:
	if n >= 0 {
		atomic.AddUint64(&counter.count, uint64(n))
	}

	return n, err
}

// Count function return counted bytes
func (counter *WriterCounter) Count() uint64 {
	return atomic.LoadUint64(&counter.count)
}
