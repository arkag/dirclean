//go:build linux || darwin || freebsd || netbsd || openbsd
// +build linux darwin freebsd netbsd openbsd

package fileutils

import "syscall"

func getDiskSpace(path string) (available, total uint64, err error) {
	var stat syscall.Statfs_t
	err = syscall.Statfs(path, &stat)
	if err != nil {
		return 0, 0, err
	}

	// Convert to GB
	available = (uint64(stat.Bavail) * uint64(stat.Bsize)) / (1024 * 1024 * 1024)
	total = (uint64(stat.Blocks) * uint64(stat.Bsize)) / (1024 * 1024 * 1024)

	return available, total, nil
}
