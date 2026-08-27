//go:build darwin

package help

import "strings"

var compactModifierReplacer = strings.NewReplacer(
	"Ctrl+", "⌃",
	"ctrl+", "⌃",
	"Shift+", "⇧",
	"shift+", "⇧",
	"Alt+", "⌥",
	"alt+", "⌥",
	"Meta+", "⌘",
	"meta+", "⌘",
)
