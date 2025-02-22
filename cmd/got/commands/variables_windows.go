//go:build windows
// +build windows

package commands

import "fmt"

// Windows doesn't handle the block-style very well
var (
	progressStyle = "double-"
	r, l          = "[", "]"
)

func color(content ...interface{}) string {
	return fmt.Sprint(content...)
}
