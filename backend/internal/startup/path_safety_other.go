//go:build !windows

package startup

import (
	"os"
)

func isReparsePoint(os.FileInfo) bool {
	return false
}

func replaceFileAtomically(sourcePath, targetPath string) error {
	return os.Rename(sourcePath, targetPath)
}
