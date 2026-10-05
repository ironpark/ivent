package key

import (
	"runtime"
	"strings"
)

// Style selects how Combo.Format writes a combination for people to read.
type Style int

const (
	// StyleText writes key names joined by "+", such as "Ctrl+Shift+A", using the platform's
	// modifier names (Option and Cmd on macOS, Win on Windows).
	StyleText Style = iota
	// StyleSymbols writes modifiers as macOS menu symbols, such as "⌃⇧A".
	StyleSymbols
)

// Keys missing on a platform are Invalid, so the names are listed as pairs and put in maps by nameMap.
var displayNames = nameMap([]codeName{
	{Grave, "`"}, {Minus, "-"}, {Equal, "="}, {LeftBracket, "["}, {RightBracket, "]"}, {Backslash, `\`},
	{Semicolon, ";"}, {Quote, "'"}, {Comma, ","}, {Dot, "."}, {Slash, "/"},
	{SpaceBar, "Space"}, {ESC, "Esc"}, {CapsLock, "Caps Lock"},
	{ArrowUp, "Up"}, {ArrowDown, "Down"}, {ArrowLeft, "Left"}, {ArrowRight, "Right"},
	{PageUp, "Page Up"}, {PageDown, "Page Down"},
	{PadPlus, "Numpad +"}, {PadMinus, "Numpad -"}, {PadAsterisk, "Numpad *"}, {PadSlash, "Numpad /"},
	{PadEnter, "Numpad Enter"}, {PadDecimal, "Numpad ."}, {PadEquals, "Numpad ="}, {PadClear, "Clear"},
	{PrintScreen, "Print Screen"}, {ScrollLock, "Scroll Lock"}, {NumLock, "Num Lock"},
	{VolumeUp, "Volume Up"}, {VolumeDown, "Volume Down"}, {MediaNext, "Next Track"}, {MediaPrev, "Previous Track"},
	{MediaStop, "Stop"}, {MediaPlayPause, "Play/Pause"},
	{MouseLeft, "Left Click"}, {MouseRight, "Right Click"}, {MouseMiddle, "Middle Click"}, {MouseX1, "Mouse 4"}, {MouseX2, "Mouse 5"},
})

var symbolNames = nameMap([]codeName{
	{Enter, "↩"}, {Backspace, "⌫"}, {Delete, "⌦"}, {ESC, "⎋"}, {Tab, "⇥"},
	{ArrowUp, "↑"}, {ArrowDown, "↓"}, {ArrowLeft, "←"}, {ArrowRight, "→"},
	{PageUp, "⇞"}, {PageDown, "⇟"}, {Home, "↖"}, {End, "↘"}, {CapsLock, "⇪"}, {Fn, "fn"},
})

type codeName struct {
	code Code
	name string
}

func nameMap(pairs []codeName) map[Code]string {
	m := make(map[Code]string, len(pairs))
	for _, p := range pairs {
		if p.code != Invalid {
			m[p.code] = p.name
		}
	}
	return m
}

// modifierText returns the platform name of a side-independent modifier.
func modifierText(i int) string {
	switch modifiers[i].name {
	case "ALT":
		if runtime.GOOS == "darwin" {
			return "Option"
		}
		return "Alt"
	case "SUPER":
		switch runtime.GOOS {
		case "darwin":
			return "Cmd"
		case "windows":
			return "Win"
		}
		return "Super"
	case "CTRL":
		return "Ctrl"
	}
	return "Shift"
}

var modifierSymbols = [...]string{"⌃", "⌥", "⇧", "⌘"}

// modifierInfo returns the modifier index of a side-independent modifier or of one of its
// physical keys, and the side ("Left", "Right" or "" for side-independent).
func modifierInfo(code Code) (index int, side string, ok bool) {
	if i, ok := modifierIndex(code); ok {
		return i, "", true
	}
	for i, m := range modifiers {
		switch code {
		case m.left:
			return i, "Left", true
		case m.right:
			return i, "Right", true
		}
	}
	return 0, "", false
}

// DisplayName returns the name of a key for people to read, such as "Ctrl", "Page Up" or "[".
func DisplayName(code Code) string {
	if i, side, ok := modifierInfo(code); ok {
		return strings.TrimSpace(side + " " + modifierText(i))
	}
	if name, ok := displayNames[code]; ok {
		return name
	}
	if name, ok := codeToDisplay[code]; ok {
		if digit, ok := strings.CutPrefix(name, "Pad"); ok {
			return "Numpad " + digit
		}
		// Num0-Num9 are the digit keys.
		return strings.TrimPrefix(name, "Num")
	}
	return Name(code)
}

// Display returns the combination for people to read, in the style of the platform:
// symbols on macOS ("⌃⇧A") and text elsewhere ("Ctrl+Shift+A").
func (c Combo) Display() string {
	if runtime.GOOS == "darwin" {
		return c.Format(StyleSymbols)
	}
	return c.Format(StyleText)
}

// Format returns the combination for people to read, modifiers first.
// Unlike String, the result is not meant to be parsed.
func (c Combo) Format(style Style) string {
	var mods, keys []string
	for _, code := range c.Codes() {
		i, side, isMod := modifierInfo(code)
		switch {
		case isMod && style == StyleSymbols:
			mods = append(mods, side+modifierSymbols[i])
		case isMod:
			mods = append(mods, DisplayName(code))
		case style == StyleSymbols && symbolNames[code] != "":
			keys = append(keys, symbolNames[code])
		default:
			keys = append(keys, DisplayName(code))
		}
	}
	if style == StyleSymbols {
		return strings.Join(mods, "") + strings.Join(keys, "+")
	}
	return strings.Join(append(mods, keys...), "+")
}
