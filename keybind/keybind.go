package keybind

import (
	"slices"
	"strings"

	"github.com/gdamore/tcell/v3"
)

type Keybind struct {
	keys []string
	help Help
}

type Option func(*Keybind)

func NewKeybind(options ...Option) Keybind {
	k := &Keybind{}
	for _, option := range options {
		option(k)
	}
	return *k
}

func NewSingleKeybind(key, desc string) Keybind {
	return NewKeybind(WithKeys(key), WithHelp(key, desc))
}

func WithKeys(keys ...string) Option {
	return func(k *Keybind) {
		k.keys = normalizeKeys(keys...)
	}
}

func WithHelp(key, desc string) Option {
	return func(k *Keybind) {
		k.help = Help{Key: key, Desc: desc}
	}
}

func (k Keybind) Keys() []string {
	return k.keys
}

func (k *Keybind) SetKeys(keys ...string) {
	k.keys = normalizeKeys(keys...)
}

func (k Keybind) Help() Help {
	return k.help
}

func (k *Keybind) SetHelp(key, desc string) {
	k.help = Help{Key: key, Desc: desc}
}

type Help struct {
	Key  string
	Desc string
}

type KeyMsg = *tcell.EventKey

func Matches(msg KeyMsg, keybinds ...Keybind) bool {
	if msg == nil {
		return false
	}

	key := keyMsgString(msg)
	for _, keybind := range keybinds {
		if slices.Contains(keybind.keys, key) {
			return true
		}
	}
	return false
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
	key = strings.TrimSpace(key)
	if key == "" {
		return ""
	}

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
	case "ctrl-c":
		return "ctrl+c"
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
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, value := range in {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func keyMsgString(msg KeyMsg) string {
	if msg == nil {
		return ""
	}

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
	return strings.Join(append(uniqueOrdered(mods), primary), "+")
}

func keyName(key tcell.Key) string {
	switch key {
	case tcell.KeyEnter:
		return "enter"
	case tcell.KeyEscape:
		return "esc"
	case tcell.KeyTab:
		return "tab"
	case tcell.KeyBacktab:
		return "shift+tab"
	case tcell.KeyHome:
		return "home"
	case tcell.KeyEnd:
		return "end"
	case tcell.KeyUp:
		return "up"
	case tcell.KeyDown:
		return "down"
	case tcell.KeyLeft:
		return "left"
	case tcell.KeyRight:
		return "right"
	case tcell.KeyPgUp:
		return "pgup"
	case tcell.KeyPgDn:
		return "pgdn"
	case tcell.KeyDelete:
		return "delete"
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		return "backspace"
	case tcell.KeyInsert:
		return "insert"
	default:
		return ""
	}
}
