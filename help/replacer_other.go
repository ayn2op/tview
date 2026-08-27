//go:build !darwin

package help

import "strings"

var compactModifierReplacer = strings.NewReplacer(
	"Ctrl+", "^",
	"ctrl+", "^",
	"Control+", "^",
	"control+", "^",
	"Shift+", "S-",
	"shift+", "S-",
	"Alt+", "A-",
	"alt+", "A-",
	"Meta+", "M-",
	"meta+", "M-",
)
