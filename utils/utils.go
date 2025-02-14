package utils

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/arkag/dirclean/logging"
)

func FileExists(path string) bool {
	logging.LogMessage("DEBUG", fmt.Sprintf("Checking existence of path: %s", path))

	// First try Lstat to handle symlinks properly
	info, err := os.Lstat(path)
	if err != nil {
		logging.LogMessage("DEBUG", fmt.Sprintf("Lstat error for %s: %v", path, err))
		return false
	}

	logging.LogMessage("DEBUG", fmt.Sprintf("File mode: %v", info.Mode()))
	logging.LogMessage("DEBUG", fmt.Sprintf("File size: %d", info.Size()))
	logging.LogMessage("DEBUG", fmt.Sprintf("File mod time: %v", info.ModTime()))

	// If it's a symlink, follow it
	if info.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(path)
		if err != nil {
			logging.LogMessage("DEBUG", fmt.Sprintf("Readlink error for %s: %v", path, err))
			return false
		}
		logging.LogMessage("DEBUG", fmt.Sprintf("Symlink target: %s", target))

		// If the target path is relative, make it absolute
		if !filepath.IsAbs(target) {
			target = filepath.Join(filepath.Dir(path), target)
			logging.LogMessage("DEBUG", fmt.Sprintf("Absolute symlink target: %s", target))
		}

		// Try to stat the target
		targetInfo, err := os.Stat(target)
		if err != nil {
			logging.LogMessage("DEBUG", fmt.Sprintf("Target stat error for %s: %v", target, err))
			return false
		}
		logging.LogMessage("DEBUG", fmt.Sprintf("Target exists and is accessible"))
		return true
	}

	// Try direct Stat as well
	_, err = os.Stat(path)
	if err != nil {
		logging.LogMessage("DEBUG", fmt.Sprintf("Direct stat error for %s: %v", path, err))
		return false
	}

	logging.LogMessage("DEBUG", fmt.Sprintf("File exists and is accessible: %s", path))
	return true
}

func IsDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func GetAbsPath(path string) string {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	return absPath
}
