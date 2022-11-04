//go:build !windows

package unmask

import "syscall"

func Unmask() {
	syscall.Umask(0)
}
