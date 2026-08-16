//go:build !windows

package startup

import "os"

func isReparsePoint(os.FileInfo) bool {
	return false
}
