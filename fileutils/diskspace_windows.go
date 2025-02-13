//go:build windows
// +build windows

package fileutils

import (
	"syscall"
	"unsafe"
)

func getDiskSpace(path string) (available, total uint64, err error) {
	kernel32, err := syscall.LoadDLL("kernel32.dll")
	if err != nil {
		return 0, 0, err
	}
	GetDiskFreeSpaceExW, err := kernel32.FindProc("GetDiskFreeSpaceExW")
	if err != nil {
		return 0, 0, err
	}

	var freeBytesAvailable, totalBytes uint64
	pathPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return 0, 0, err
	}

	r1, _, err := GetDiskFreeSpaceExW.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		uintptr(unsafe.Pointer(&freeBytesAvailable)),
		uintptr(unsafe.Pointer(&totalBytes)),
		uintptr(0))

	if r1 == 0 {
		return 0, 0, err
	}

	// Convert to GB
	available = freeBytesAvailable / (1024 * 1024 * 1024)
	total = totalBytes / (1024 * 1024 * 1024)

	return available, total, nil
}
