package key

type Code int

const (
	// ANSI 키코드
	A, B, C, D, E, F, G, H, I, J, K, L, M = Code(0x00), Code(0x0B), Code(0x08), Code(0x02), Code(0x0E), Code(0x03), Code(0x05), Code(0x04), Code(0x22), Code(0x26), Code(0x28), Code(0x25), Code(0x2E)
	N, O, P, Q, R, S, T, U, V, W, X, Y, Z = Code(0x2D), Code(0x1F), Code(0x23), Code(0x0C), Code(0x0F), Code(0x01), Code(0x11), Code(0x20), Code(0x09), Code(0x0D), Code(0x07), Code(0x10), Code(0x06)
	Num0                                  = Code(0x1D)
	Num1                                  = Code(0x12)
	Num2                                  = Code(0x13)
	Num3                                  = Code(0x14)
	Num4                                  = Code(0x15)
	Num5                                  = Code(0x17)
	Num6                                  = Code(0x16)
	Num7                                  = Code(0x1A)
	Num8                                  = Code(0x1C)
	Num9                                  = Code(0x19)
	Backslash                             = Code(0x2A)
	Comma                                 = Code(0x2B)
	Equal                                 = Code(0x18)
	Grave                                 = Code(0x32)
	LeftBracket                           = Code(0x21)
	Minus                                 = Code(0x1B)
	Period                                = Code(0x2F)
	Quote                                 = Code(0x27)
	RightBracket                          = Code(0x1E)
	Semicolon                             = Code(0x29)
	Slash                                 = Code(0x2C)
	// Keypad 키코드
	Keypad0        = Code(0x52)
	Keypad1        = Code(0x53)
	Keypad2        = Code(0x54)
	Keypad3        = Code(0x55)
	Keypad4        = Code(0x56)
	Keypad5        = Code(0x57)
	Keypad6        = Code(0x58)
	Keypad7        = Code(0x59)
	Keypad8        = Code(0x5B)
	Keypad9        = Code(0x5C)
	KeypadClear    = Code(0x47)
	KeypadDecimal  = Code(0x41)
	KeypadDivide   = Code(0x4B)
	KeypadEnter    = Code(0x4C)
	KeypadEquals   = Code(0x51)
	KeypadMinus    = Code(0x4E)
	KeypadMultiply = Code(0x43)
	KeypadPlus     = Code(0x45)
	// 레이아웃에 독립적인 키코드
	CapsLock      = Code(0x39)
	Command       = Code(0x37)
	Control       = Code(0x3B)
	Delete        = Code(0x33)
	DownArrow     = Code(0x7D)
	End           = Code(0x77)
	Escape        = Code(0x35)
	ForwardDelete = Code(0x75)
	Function      = Code(0x3F)
	Help          = Code(0x72)
	Home          = Code(0x73)
	LeftArrow     = Code(0x7B)
	Option        = Code(0x3A)
	PageDown      = Code(0x79)
	PageUp        = Code(0x74)
	Return        = Code(0x24)
	RightArrow    = Code(0x7C)
	RightControl  = Code(0x3E)
	RightOption   = Code(0x3D)
	RightShift    = Code(0x3C)
	Shift         = Code(0x38)
	Space         = Code(0x31)
	Tab           = Code(0x30)
	UpArrow       = Code(0x7E)

	// F열 키코드
	F1  = Code(0x7A)
	F2  = Code(0x78)
	F3  = Code(0x63)
	F4  = Code(0x76)
	F5  = Code(0x60)
	F6  = Code(0x61)
	F7  = Code(0x62)
	F8  = Code(0x64)
	F9  = Code(0x65)
	F10 = Code(0x6D)
	F11 = Code(0x67)
	F12 = Code(0x6F)
	F13 = Code(0x69)
	F14 = Code(0x6B)
	F15 = Code(0x71)
	F16 = Code(0x6A)
	F17 = Code(0x40)
	F18 = Code(0x4F)
	F19 = Code(0x50)
	F20 = Code(0x5A)

	// 볼륨 키코드
	VolumeDown = Code(0x49)
	VolumeUp   = Code(0x48)
	Mute       = Code(0x4A)
)

var (
	codeToName = map[Code]string{
		A: "A", B: "B", C: "C", D: "D", E: "E", F: "F", G: "G", H: "H", I: "I", J: "J", K: "K", L: "L", M: "M",
		N: "N", O: "O", P: "P", Q: "Q", R: "R", S: "S", T: "T", U: "U", V: "V", W: "W", X: "X", Y: "Y", Z: "Z",
		Num1: "1", Num2: "2", Num3: "3", Num4: "4", Num5: "5", Num6: "6", Num7: "7", Num8: "8", Num9: "9", Num0: "0",
		Backslash:    "Backslash",
		Comma:        "Comma",
		Equal:        "Equal",
		Grave:        "Grave",
		LeftBracket:  "LeftBracket",
		Minus:        "Minus",
		Period:       "Period",
		Quote:        "Quote",
		RightBracket: "RightBracket",
		Semicolon:    "Semicolon",
		Slash:        "Slash",

		//Keypad키코드
		Keypad0:        "Keypad0",
		Keypad1:        "Keypad1",
		Keypad2:        "Keypad2",
		Keypad3:        "Keypad3",
		Keypad4:        "Keypad4",
		Keypad5:        "Keypad5",
		Keypad6:        "Keypad6",
		Keypad7:        "Keypad7",
		Keypad8:        "Keypad8",
		Keypad9:        "Keypad9",
		KeypadClear:    "KeypadClear",
		KeypadDecimal:  "KeypadDecimal",
		KeypadDivide:   "KeypadDivide",
		KeypadEnter:    "KeypadEnter",
		KeypadEquals:   "KeypadEquals",
		KeypadMinus:    "KeypadMinus",
		KeypadMultiply: "KeypadMultiply",
		KeypadPlus:     "KeypadPlus",

		//레이아웃에독립적인키코드
		CapsLock:      "CapsLock",
		Command:       "Command",
		Control:       "Control",
		Delete:        "Delete",
		DownArrow:     "DownArrow",
		End:           "End",
		Escape:        "Escape",
		ForwardDelete: "ForwardDelete",
		Function:      "Function",
		Help:          "Help",
		Home:          "Home",
		LeftArrow:     "LeftArrow",
		Option:        "Option",
		PageDown:      "PageDown",
		PageUp:        "PageUp",
		Return:        "Return",
		RightArrow:    "RightArrow",
		RightControl:  "RightControl",
		RightOption:   "RightOption",
		RightShift:    "RightShift",
		Shift:         "Shift",
		Space:         "Space",
		Tab:           "Tab",
		UpArrow:       "UpArrow",

		//F키코드
		F1:  "F1",
		F2:  "F2",
		F3:  "F3",
		F4:  "F4",
		F5:  "F5",
		F6:  "F6",
		F7:  "F7",
		F8:  "F8",
		F9:  "F9",
		F10: "F10",
		F11: "F11",
		F12: "F12",
		F13: "F13",
		F14: "F14",
		F15: "F15",
		F16: "F16",
		F17: "F17",
		F18: "F18",
		F19: "F19",
		F20: "F20",

		//볼륨키코드
		VolumeDown: "VolumeDown",
		VolumeUp:   "VolumeUp",
		Mute:       "Mute",
	}

	nameToCode = map[string]Code{
		"A": A, "B": B, "C": C, "D": D, "E": E, "F": F, "G": G, "H": H, "I": I, "J": J, "K": K, "L": L, "M": M,
		"N": N, "O": O, "P": P, "Q": Q, "R": R, "S": S, "T": T, "U": U, "V": V, "W": W, "X": X, "Y": Y, "Z": Z,
		"1": Num1, "2": Num2, "3": Num3, "4": Num4, "5": Num5, "6": Num6, "7": Num7, "8": Num8, "9": Num9, "0": Num0,
		"Backslash":    Backslash,
		"Comma":        Comma,
		"Equal":        Equal,
		"Grave":        Grave,
		"LeftBracket":  LeftBracket,
		"Minus":        Minus,
		"Period":       Period,
		"Quote":        Quote,
		"RightBracket": RightBracket,
		"Semicolon":    Semicolon,
		"Slash":        Slash,

		//Keypad키코드
		"Keypad0":        Keypad0,
		"Keypad1":        Keypad1,
		"Keypad2":        Keypad2,
		"Keypad3":        Keypad3,
		"Keypad4":        Keypad4,
		"Keypad5":        Keypad5,
		"Keypad6":        Keypad6,
		"Keypad7":        Keypad7,
		"Keypad8":        Keypad8,
		"Keypad9":        Keypad9,
		"KeypadClear":    KeypadClear,
		"KeypadDecimal":  KeypadDecimal,
		"KeypadDivide":   KeypadDivide,
		"KeypadEnter":    KeypadEnter,
		"KeypadEquals":   KeypadEquals,
		"KeypadMinus":    KeypadMinus,
		"KeypadMultiply": KeypadMultiply,
		"KeypadPlus":     KeypadPlus,

		//레이아웃에독립적인키코드
		"CapsLock":      CapsLock,
		"Command":       Command,
		"Control":       Control,
		"Delete":        Delete,
		"DownArrow":     DownArrow,
		"End":           End,
		"Escape":        Escape,
		"ForwardDelete": ForwardDelete,
		"Function":      Function,
		"Help":          Help,
		"Home":          Home,
		"LeftArrow":     LeftArrow,
		"Option":        Option,
		"PageDown":      PageDown,
		"PageUp":        PageUp,
		"Return":        Return,
		"RightArrow":    RightArrow,
		"RightControl":  RightControl,
		"RightOption":   RightOption,
		"RightShift":    RightShift,
		"Shift":         Shift,
		"Space":         Space,
		"Tab":           Tab,
		"UpArrow":       UpArrow,

		//F키코드
		"F1":  F1,
		"F2":  F2,
		"F3":  F3,
		"F4":  F4,
		"F5":  F5,
		"F6":  F6,
		"F7":  F7,
		"F8":  F8,
		"F9":  F9,
		"F10": F10,
		"F11": F11,
		"F12": F12,
		"F13": F13,
		"F14": F14,
		"F15": F15,
		"F16": F16,
		"F17": F17,
		"F18": F18,
		"F19": F19,
		"F20": F20,

		//볼륨키코드
		"VolumeDown": VolumeDown,
		"VolumeUp":   VolumeUp,
		"Mute":       Mute,
	}
)

func (k Code) Name() string {
	return codeToName[k]
}

func Name(code Code) string {
	return codeToName[code]
}

func Names(codes ...Code) []string {
	names := make([]string, len(codes))
	for i, code := range codes {
		names[i] = Name(code)
	}
	return names
}

func Codes(names ...string) []Code {
	codes := make([]Code, len(names))
	for i, name := range names {
		codes[i] = nameToCode[name]
	}
	return codes
}
