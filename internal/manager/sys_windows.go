//go:build windows

package manager

import "syscall"

func sysProcAttrWindows(flags uint32) *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		CreationFlags: flags,
	}
}
