// +build !windows

package oci

import (
	"os"
	"syscall"
	"testing"
)

func TestMain(m *testing.M) {
	oldmask := syscall.Umask(0)
	code := m.Run()
	syscall.Umask(oldmask)
	os.Exit(code)
}
