// Code generated by internal/gen/main.go DO NOT EDIT.
package key

const (
	// Alphabet
	A, B, C, D, E, F, G, H, I, J, K, L, M = Code(0x41), Code(0x42), Code(0x43), Code(0x44), Code(0x45), Code(0x46), Code(0x47), Code(0x48), Code(0x49), Code(0x4A), Code(0x4B), Code(0x4C), Code(0x4D)
	N, O, P, Q, R, S, T, U, V, W, X, Y, Z = Code(0x4E), Code(0x4F), Code(0x50), Code(0x51), Code(0x52), Code(0x53), Code(0x54), Code(0x55), Code(0x56), Code(0x57), Code(0x58), Code(0x59), Code(0x5A)

	// Extra Keys
	Grave, Minus, Equal, LeftBracket, RightBracket = Code(0xC0), Code(0xBD), Code(0xBB), Code(0xDB), Code(0xDD)
	Backslash, Semicolon, Quote, Comma, Dot, Slash = Code(0xDC), Code(0xBA), Code(0xDE), Code(0xBC), Code(0xBE), Code(0xBF)

	// Number
	Num0, Num1, Num2, Num3, Num4, Num5, Num6, Num7, Num8, Num9 = Code(0x30), Code(0x31), Code(0x32), Code(0x33), Code(0x34), Code(0x35), Code(0x36), Code(0x37), Code(0x38), Code(0x39)

	// Function
	F1, F2, F3, F4, F5, F6, F7, F8, F9, F10          = Code(0x70), Code(0x71), Code(0x72), Code(0x73), Code(0x74), Code(0x75), Code(0x76), Code(0x77), Code(0x78), Code(0x79)
	F11, F12, F13, F14, F15, F16, F17, F18, F19, F20 = Code(0x7A), Code(0x7B), Code(0x7C), Code(0x7D), Code(0x7E), Code(0x7F), Code(0x80), Code(0x81), Code(0x82), Code(0x83)

	// Modifier
	Shift, Ctrl, Alt, Super, RightShift, RightCtrl, RightAlt, RightSuper, CapsLock = Code(0xA0), Code(0xA2), Code(0xA4), Code(0x5B), Code(0xA1), Code(0xA3), Code(0xA5), Code(0x5C), Code(0x14)

	// Navigation
	ArrowUp, ArrowDown, ArrowLeft, ArrowRight, Home, End, PageUp, PageDown = Code(0x26), Code(0x28), Code(0x25), Code(0x27), Code(0x24), Code(0x23), Code(0x21), Code(0x22)

	// Editing
	Backspace, Delete, Insert, Enter = Code(0x08), Code(0x2E), Code(0x2D), Code(0x0D)

	// Special
	ESC, SpaceBar, Tab = Code(0x1B), Code(0x20), Code(0x09)

	// Keypad
	Pad0, Pad1, Pad2, Pad3, Pad4, Pad5, Pad6, Pad7, Pad8, Pad9 = Code(0x60), Code(0x61), Code(0x62), Code(0x63), Code(0x64), Code(0x65), Code(0x66), Code(0x67), Code(0x68), Code(0x69)

	// Keypad Extra
	PadPlus, PadMinus, PadAsterisk, PadSlash = Code(0x6B), Code(0x6D), Code(0x6A), Code(0x6F)

	// KeyCode Aliases

	QuestionMark = Slash
	Option       = Alt
	Cmd          = Super
	Win          = Super
)

var (
	codeToName = map[Code]string{
		A: "A", B: "B", C: "C", D: "D", E: "E", F: "F", G: "G", H: "H", I: "I", J: "J", K: "K", L: "L", M: "M", N: "N", O: "O", P: "P", Q: "Q", R: "R", S: "S", T: "T", U: "U", V: "V", W: "W", X: "X", Y: "Y", Z: "Z",
		Grave: "GRAVE", Minus: "MINUS", Equal: "EQUAL", LeftBracket: "LEFTBRACKET", RightBracket: "RIGHTBRACKET", Backslash: "BACKSLASH", Semicolon: "SEMICOLON", Quote: "QUOTE", Comma: "COMMA", Dot: "DOT", Slash: "SLASH",
		Num0: "NUM0", Num1: "NUM1", Num2: "NUM2", Num3: "NUM3", Num4: "NUM4", Num5: "NUM5", Num6: "NUM6", Num7: "NUM7", Num8: "NUM8", Num9: "NUM9",
		F1: "F1", F2: "F2", F3: "F3", F4: "F4", F5: "F5", F6: "F6", F7: "F7", F8: "F8", F9: "F9", F10: "F10", F11: "F11", F12: "F12", F13: "F13", F14: "F14", F15: "F15", F16: "F16", F17: "F17", F18: "F18", F19: "F19", F20: "F20",
		Shift: "SHIFT", Ctrl: "CTRL", Alt: "ALT", Super: "SUPER", RightShift: "RIGHTSHIFT", RightCtrl: "RIGHTCTRL", RightAlt: "RIGHTALT", RightSuper: "RIGHTSUPER", CapsLock: "CAPSLOCK",
		ArrowUp: "ARROWUP", ArrowDown: "ARROWDOWN", ArrowLeft: "ARROWLEFT", ArrowRight: "ARROWRIGHT", Home: "HOME", End: "END", PageUp: "PAGEUP", PageDown: "PAGEDOWN",
		Backspace: "BACKSPACE", Delete: "DELETE", Insert: "INSERT", Enter: "ENTER",
		ESC: "ESC", SpaceBar: "SPACEBAR", Tab: "TAB",
		Pad0: "PAD0", Pad1: "PAD1", Pad2: "PAD2", Pad3: "PAD3", Pad4: "PAD4", Pad5: "PAD5", Pad6: "PAD6", Pad7: "PAD7", Pad8: "PAD8", Pad9: "PAD9",
		PadPlus: "PADPLUS", PadMinus: "PADMINUS", PadAsterisk: "PADASTERISK", PadSlash: "PADSLASH",
	}
	nameToCode = map[string]Code{
		"A": A, "B": B, "C": C, "D": D, "E": E, "F": F, "G": G, "H": H, "I": I, "J": J, "K": K, "L": L, "M": M, "N": N, "O": O, "P": P, "Q": Q, "R": R, "S": S, "T": T, "U": U, "V": V, "W": W, "X": X, "Y": Y, "Z": Z,
		"GRAVE": Grave, "MINUS": Minus, "EQUAL": Equal, "LEFTBRACKET": LeftBracket, "RIGHTBRACKET": RightBracket, "BACKSLASH": Backslash, "SEMICOLON": Semicolon, "QUOTE": Quote, "COMMA": Comma, "DOT": Dot, "SLASH": Slash,
		"NUM0": Num0, "NUM1": Num1, "NUM2": Num2, "NUM3": Num3, "NUM4": Num4, "NUM5": Num5, "NUM6": Num6, "NUM7": Num7, "NUM8": Num8, "NUM9": Num9,
		"F1": F1, "F2": F2, "F3": F3, "F4": F4, "F5": F5, "F6": F6, "F7": F7, "F8": F8, "F9": F9, "F10": F10, "F11": F11, "F12": F12, "F13": F13, "F14": F14, "F15": F15, "F16": F16, "F17": F17, "F18": F18, "F19": F19, "F20": F20,
		"SHIFT": Shift, "CTRL": Ctrl, "ALT": Alt, "SUPER": Super, "RIGHTSHIFT": RightShift, "RIGHTCTRL": RightCtrl, "RIGHTALT": RightAlt, "RIGHTSUPER": RightSuper, "CAPSLOCK": CapsLock,
		"ARROWUP": ArrowUp, "ARROWDOWN": ArrowDown, "ARROWLEFT": ArrowLeft, "ARROWRIGHT": ArrowRight, "HOME": Home, "END": End, "PAGEUP": PageUp, "PAGEDOWN": PageDown,
		"BACKSPACE": Backspace, "DELETE": Delete, "INSERT": Insert, "ENTER": Enter,
		"ESC": ESC, "SPACEBAR": SpaceBar, "TAB": Tab,
		"PAD0": Pad0, "PAD1": Pad1, "PAD2": Pad2, "PAD3": Pad3, "PAD4": Pad4, "PAD5": Pad5, "PAD6": Pad6, "PAD7": Pad7, "PAD8": Pad8, "PAD9": Pad9,
		"PADPLUS": PadPlus, "PADMINUS": PadMinus, "PADASTERISK": PadAsterisk, "PADSLASH": PadSlash,
		// KeyCode Aliases
		"_": Minus, "-": Minus,
		"+": Equal, "=": Equal,
		"{": LeftBracket, "[": LeftBracket,
		"}": RightBracket, "]": RightBracket,
		"|": Backslash, "\\": Backslash,
		":": Semicolon, ";": Semicolon,
		"\"": Quote, "'": Quote,
		"<": Comma, ",": Comma,
		">": Dot, ".": Dot,
		"QUESTIONMARK": Slash, "?": Slash, "/": Slash,
		"OPTION": Alt,
		"CMD":    Super, "WIN": Super,
	}
)
