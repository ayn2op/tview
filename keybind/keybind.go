package keybind

import (
	"slices"
	"strings"

	"github.com/ayn2op/tview"
	"github.com/gdamore/tcell/v3"
)

type Keybind struct {
	keys []string
	help Help
}

// New returns a keybind for keys, without help.
func New(keys ...string) Keybind {
	return Keybind{keys: normalizeKeys(keys...)}
}

// NewSingleKeybind returns a keybind for key, with key and desc as its help.
func NewSingleKeybind(key, desc string) Keybind {
	return New(key).WithHelp(key, desc)
}

// WithHelp returns k with key and desc as its help.
func (k Keybind) WithHelp(key, desc string) Keybind {
	k.help = Help{Key: key, Desc: desc}
	return k
}

func (k Keybind) Keys() []string {
	return k.keys
}

func (k Keybind) Help() Help {
	return k.help
}

type Help struct {
	Key  string
	Desc string
}

func Matches(msg tview.KeyMsg, keybinds ...Keybind) bool {
	if msg == nil {
		return false
	}

	key := String(msg)
	return slices.ContainsFunc(keybinds, func(k Keybind) bool { return slices.Contains(k.keys, key) })
}

func normalizeKeys(keys ...string) []string {
	normalized := make([]string, 0, len(keys))
	for _, key := range keys {
		key = normalizeKey(key)
		if key == "" {
			continue
		}
		normalized = append(normalized, key)
	}
	return normalized
}

func normalizeKey(key string) string {
	parts := strings.Split(key, "+")
	mods := make([]string, 0, len(parts))
	primary := ""
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		switch strings.ToLower(part) {
		case "ctrl", "control":
			mods = append(mods, "ctrl")
		case "alt":
			mods = append(mods, "alt")
		case "shift":
			mods = append(mods, "shift")
		case "meta":
			mods = append(mods, "meta")
		default:
			primary = normalizePrimaryKey(part)
		}
	}

	if primary == "" {
		return ""
	}

	if primary == "backtab" {
		mods = append(mods, "shift")
		primary = "tab"
	}

	if len(mods) > 0 && len([]rune(primary)) == 1 {
		primary = strings.ToLower(primary)
	}

	if len(mods) == 0 {
		return primary
	}

	return strings.Join(append(uniqueOrdered(mods), primary), "+")
}

func normalizePrimaryKey(key string) string {
	if strings.HasPrefix(key, "Rune[") && strings.HasSuffix(key, "]") && len(key) >= 7 {
		return key[5 : len(key)-1]
	}

	switch strings.ToLower(key) {
	case "esc", "escape":
		return "esc"
	case "return":
		return "enter"
	case "pageup":
		return "pgup"
	case "pagedown":
		return "pgdn"
	}

	if strings.HasPrefix(strings.ToLower(key), "ctrl-") && len(key) > len("ctrl-") {
		return "ctrl+" + strings.ToLower(key[len("ctrl-"):])
	}

	if len([]rune(key)) == 1 {
		return key
	}

	return strings.ToLower(key)
}

func uniqueOrdered(in []string) []string {
	out := make([]string, 0, len(in))
	for _, value := range in {
		if !slices.Contains(out, value) {
			out = append(out, value)
		}
	}
	return out
}

// String returns the name of the key msg is, in the form keybinds are written, such as "ctrl+home".
func String(msg tview.KeyMsg) string {
	key := msg.Key()
	if key >= tcell.KeyCtrlA && key <= tcell.KeyCtrlZ {
		return "ctrl+" + string(rune('a'+(key-tcell.KeyCtrlA)))
	}

	primary := keyName(key)
	if primary == "" && key == tcell.KeyRune {
		primary = msg.Str()
	}
	if primary == "" {
		return normalizeKey(msg.Name())
	}

	mods := make([]string, 0, 4)
	if msg.Modifiers()&tcell.ModCtrl != 0 {
		mods = append(mods, "ctrl")
	}
	if msg.Modifiers()&tcell.ModAlt != 0 {
		mods = append(mods, "alt")
	}
	if msg.Modifiers()&tcell.ModShift != 0 {
		mods = append(mods, "shift")
	}
	if msg.Modifiers()&tcell.ModMeta != 0 {
		mods = append(mods, "meta")
	}
	if len(mods) == 0 {
		return primary
	}
	return strings.Join(append(mods, primary), "+")
}

func keyName(key tcell.Key) string {
	switch key {
	case tcell.KeyBacktab:
		return "shift+tab"
	case tcell.KeyBackspace2:
		return "backspace"
	case tcell.KeyEnter, tcell.KeyEscape, tcell.KeyTab, tcell.KeyHome, tcell.KeyEnd, tcell.KeyUp, tcell.KeyDown, tcell.KeyLeft, tcell.KeyRight, tcell.KeyPgUp, tcell.KeyPgDn, tcell.KeyDelete, tcell.KeyBackspace, tcell.KeyInsert:
		return strings.ToLower(tcell.KeyNames[key])
	}
	return ""
}
