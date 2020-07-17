package compress

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
)

// Gzip compresses the given file, removing the source log file if successful.
func Gzip(src string) (err error) {
	f, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat log file: %v", err)
	}

	// If this file already exists, we presume it was created by
	// a previous attempt to compress the log file.
	gzf, err := os.OpenFile(src+".gz", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, fi.Mode())
	if err != nil {
		return fmt.Errorf("failed to open compressed log file: %v", err)
	}

	defer gzf.Close()

	gz := gzip.NewWriter(gzf)
	defer gz.Close()

	if _, err := io.Copy(gz, f); err != nil {
		return err
	}

	if err := os.Remove(src); err != nil {
		return err
	}

	return nil
}
