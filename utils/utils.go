package utils

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/arkag/dirclean/logging"
)

func FileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsPermission(err) {
			logging.LogMessage("ERROR", fmt.Sprintf("Permission denied accessing path: %s", path))
			return false
		}
		if os.IsNotExist(err) {
			logging.LogMessage("ERROR", fmt.Sprintf("Path does not exist: %s", path))
			return false
		}
		logging.LogMessage("ERROR", fmt.Sprintf("Error accessing path %s: %v", path, err))
		return false
	}

	// If it's a symlink, verify the target exists
	if info.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(path)
		if err != nil {
			logging.LogMessage("ERROR", fmt.Sprintf("Error reading symlink %s: %v", path, err))
			return false
		}
		// If the target path is relative, make it absolute
		if !filepath.IsAbs(target) {
			target = filepath.Join(filepath.Dir(path), target)
		}
		return FileExists(target)
	}

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
