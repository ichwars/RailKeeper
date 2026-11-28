//go:build !windows

package main

func openDesktopControl(_ []string, _ bool) (*desktopControl, error) {
	return nil, nil
}
