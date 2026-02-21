package dfx

import (
	"fmt"
	"strings"

	"github.com/AllenDang/cimgui-go/imgui"
)

type KeyModifier uint8

const (
	ModNone  KeyModifier = 0
	ModCtrl  KeyModifier = 1 << 0
	ModShift KeyModifier = 1 << 1
	ModAlt   KeyModifier = 1 << 2
	ModSuper KeyModifier = 1 << 3
)

// KeyEvent represents keyboard input for component action checking
type KeyEvent struct {
	Key      imgui.Key
	Pressed  bool
	Modifier KeyModifier
}

// Action represents a keybinding and its associated function
type Action struct {
	Id            string
	Label         string // display name for menu items (if empty, uses Id)
	Keys          string // e.g. "Ctrl+A", "Alt+Shift+F1"
	Handler       func()
	key           imgui.Key
	mods          KeyModifier
	shortcutLabel string // formatted shortcut for menu display
}

// ActionRegistry manages actions (unified for both App and Components)
type ActionRegistry struct {
	actions []*Action
}

type keyCombo struct {
	key  imgui.Key
	mods KeyModifier
}

func NewActionRegistry() *ActionRegistry {
	return &ActionRegistry{}
}

// Register adds an action to the registry
func (r *ActionRegistry) Register(id, keys string, handler func()) error {
	action := &Action{
		Id:      id,
		Keys:    keys,
		Handler: handler,
	}

	// parse the key binding
	if err := action.parse(); err != nil {
		return fmt.Errorf("invalid key binding %q: %w", keys, err)
	}

	return r.RegisterAction(action)
}

func (r *ActionRegistry) MustRegister(id, key string, handler func()) {
	if err := r.Register(id, key, handler); err != nil {
		panic(err)
	}
}

// RegisterAction adds a pre-created action (e.g., menu action) to the registry
func (r *ActionRegistry) RegisterAction(action *Action) error {
	// check for conflicts
	combo := keyCombo{action.key, action.mods}
	for _, existing := range r.actions {
		existingCombo := keyCombo{existing.key, existing.mods}
		if combo == existingCombo {
			return fmt.Errorf("key binding %q conflicts with action %q", action.Keys, existing.Id)
		}
	}

	r.actions = append(r.actions, action)
	return nil
}

// MustRegisterAction adds a pre-created action and panics on error
func (r *ActionRegistry) MustRegisterAction(action *Action) {
	if err := r.RegisterAction(action); err != nil {
		panic(err)
	}
}

// parse converts the key string to imgui key and modifiers
func (a *Action) parse() error {
	parts := strings.Split(a.Keys, "+")
	if len(parts) == 0 {
		return fmt.Errorf("empty key binding")
	}

	// process modifiers
	for i := 0; i < len(parts)-1; i++ {
		switch strings.ToLower(parts[i]) {
		case "ctrl":
			a.mods |= ModCtrl
		case "shift":
			a.mods |= ModShift
		case "alt":
			a.mods |= ModAlt
		case "super", "cmd", "win":
			a.mods |= ModSuper
		default:
			return fmt.Errorf("unknown modifier: %s", parts[i])
		}
	}

	// process the key
	keyName := parts[len(parts)-1]
	key, ok := parseKey(keyName)
	if !ok {
		return fmt.Errorf("unknown key: %s", keyName)
	}
	a.key = key
	return nil
}

// parseKey converts a key name to imgui.Key
func parseKey(name string) (imgui.Key, bool) {
	// single character keys
	if len(name) == 1 {
		ch := strings.ToUpper(name)[0]
		if ch >= 'A' && ch <= 'Z' {
			return imgui.KeyA + imgui.Key(ch-'A'), true
		}
		if ch >= '0' && ch <= '9' {
			return imgui.Key0 + imgui.Key(ch-'0'), true
		}
		// special single character keys
		switch ch {
		case '-':
			return imgui.KeyMinus, true
		case '=':
			return imgui.KeyEqual, true
		case '[':
			return imgui.KeyLeftBracket, true
		case ']':
			return imgui.KeyRightBracket, true
		case ';':
			return imgui.KeySemicolon, true
		case '\'':
			return imgui.KeyApostrophe, true
		case ',':
			return imgui.KeyComma, true
		case '.':
			return imgui.KeyPeriod, true
		case '/':
			return imgui.KeySlash, true
		}
	}

	// special keys
	switch strings.ToLower(name) {
	case "space":
		return imgui.KeySpace, true
	case "enter", "return":
		return imgui.KeyEnter, true
	case "esc", "escape":
		return imgui.KeyEscape, true
	case "tab":
		return imgui.KeyTab, true
	case "backspace":
		return imgui.KeyBackspace, true
	case "delete", "del":
		return imgui.KeyDelete, true
	case "left":
		return imgui.KeyLeftArrow, true
	case "right":
		return imgui.KeyRightArrow, true
	case "up":
		return imgui.KeyUpArrow, true
	case "down":
		return imgui.KeyDownArrow, true
	case "home":
		return imgui.KeyHome, true
	case "end":
		return imgui.KeyEnd, true
	case "pageup", "pgup":
		return imgui.KeyPageUp, true
	case "pagedown", "pgdn":
		return imgui.KeyPageDown, true
	}

	// function keys
	if strings.HasPrefix(strings.ToLower(name), "f") && len(name) <= 3 {
		var num int
		if _, err := fmt.Sscanf(name[1:], "%d", &num); err == nil && num >= 1 && num <= 12 {
			return imgui.KeyF1 + imgui.Key(num-1), true
		}
	}

	return 0, false
}

// DrawMenuItem renders the action as a menu item
// returns true if the menu item was clicked
func (a *Action) DrawMenuItem() bool {
	label := a.Label
	if label == "" {
		label = a.Id
	}

	if imgui.MenuItemBoolV(label, a.shortcutLabel, false, true) {
		if a.Handler != nil {
			a.Handler()
		}
		return true
	}
	return false
}

// NewMenuAction creates an action suitable for both menus and keyboard shortcuts
// label: display name for menu items (e.g., "Save As...")
// keys: keyboard shortcut (e.g., "Ctrl+Shift+S")
// handler: function to execute
func NewMenuAction(label, keys string, handler func()) *Action {
	action := &Action{
		Id:      label,
		Label:   label,
		Keys:    keys,
		Handler: handler,
	}

	if err := action.parse(); err != nil {
		panic(fmt.Errorf("invalid action %q: %w", label, err))
	}

	action.shortcutLabel = formatShortcutLabel(action.mods, action.key)
	return action
}

// formatShortcutLabel converts modifiers and key to menu display format
// e.g., ModCtrl|ModShift + KeyS → "Ctrl+Shift+S"
func formatShortcutLabel(mods KeyModifier, key imgui.Key) string {
	var parts []string

	if mods&ModCtrl != 0 {
		parts = append(parts, "Ctrl")
	}
	if mods&ModShift != 0 {
		parts = append(parts, "Shift")
	}
	if mods&ModAlt != 0 {
		parts = append(parts, "Alt")
	}
	if mods&ModSuper != 0 {
		parts = append(parts, "Super")
	}

	keyLabel := keyToLabel(key)
	if keyLabel != "" {
		parts = append(parts, keyLabel)
	}

	return strings.Join(parts, "+")
}

// keyToLabel converts imgui.Key to human-readable label
func keyToLabel(key imgui.Key) string {
	// alphabetic keys
	if key >= imgui.KeyA && key <= imgui.KeyZ {
		return string(rune('A' + (key - imgui.KeyA)))
	}

	// numeric keys
	if key >= imgui.Key0 && key <= imgui.Key9 {
		return string(rune('0' + (key - imgui.Key0)))
	}

	// function keys
	if key >= imgui.KeyF1 && key <= imgui.KeyF12 {
		return fmt.Sprintf("F%d", (key-imgui.KeyF1)+1)
	}

	// special keys
	switch key {
	case imgui.KeySpace:
		return "Space"
	case imgui.KeyEnter:
		return "Enter"
	case imgui.KeyEscape:
		return "Esc"
	case imgui.KeyTab:
		return "Tab"
	case imgui.KeyBackspace:
		return "Backspace"
	case imgui.KeyDelete:
		return "Delete"
	case imgui.KeyLeftArrow:
		return "Left"
	case imgui.KeyRightArrow:
		return "Right"
	case imgui.KeyUpArrow:
		return "Up"
	case imgui.KeyDownArrow:
		return "Down"
	case imgui.KeyHome:
		return "Home"
	case imgui.KeyEnd:
		return "End"
	case imgui.KeyPageUp:
		return "PageUp"
	case imgui.KeyPageDown:
		return "PageDown"
	case imgui.KeyMinus:
		return "-"
	case imgui.KeyEqual:
		return "="
	case imgui.KeyLeftBracket:
		return "["
	case imgui.KeyRightBracket:
		return "]"
	case imgui.KeySemicolon:
		return ";"
	case imgui.KeyApostrophe:
		return "'"
	case imgui.KeyComma:
		return ","
	case imgui.KeyPeriod:
		return "."
	case imgui.KeySlash:
		return "/"
	case imgui.KeyBackslash:
		return "\\"
	case imgui.KeyGraveAccent:
		return "`"
	}

	return ""
}
