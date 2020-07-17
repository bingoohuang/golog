package port

import (
	"fmt"
	"net"
	"os"
)

// FreeAddr asks the kernel for a free open port that is ready to use.
func FreeAddr() string {
	if v := os.Getenv("ADDR"); v != "" {
		return v
	}

	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return ":10020"
	}

	_ = l.Close()

	return fmt.Sprintf(":%d", l.Addr().(*net.TCPAddr).Port)
}
