package key

type Code int

const (
	/* A-Z Key Codes */

	A, B, C, D, E, F, G, H, I, J, K, L, M = Code(0x00), Code(0x0B), Code(0x08), Code(0x02), Code(0x0E), Code(0x03), Code(0x05), Code(0x04), Code(0x22), Code(0x26), Code(0x28), Code(0x25), Code(0x2E)
	N, O, P, Q, R, S, T, U, V, W, X, Y, Z = Code(0x2D), Code(0x1F), Code(0x23), Code(0x0C), Code(0x0F), Code(0x01), Code(0x11), Code(0x20), Code(0x09), Code(0x0D), Code(0x07), Code(0x10), Code(0x06)

	/* 0-9 Key Codes */

	Num0, Num1, Num2, Num3, Num4, Num5, Num6, Num7, Num8, Num9 = Code(0x1D), Code(0x12), Code(0x13), Code(0x14), Code(0x15), Code(0x17), Code(0x16), Code(0x1A), Code(0x1C), Code(0x19)
	Backslash                                                  = Code(0x2A)
	Comma                                                      = Code(0x2B)
	Equal                                                      = Code(0x18)
	Grave                                                      = Code(0x32)
	LeftBracket                                                = Code(0x21)
	Minus                                                      = Code(0x1B)
	Period                                                     = Code(0x2F)
	Quote                                                      = Code(0x27)
	RightBracket                                               = Code(0x1E)
	Semicolon                                                  = Code(0x29)
	Slash                                                      = Code(0x2C)

	/* Pad Key Codes */

	Pad0, Pad1, Pad2, Pad3, Pad4, Pad5, Pad6, Pad7, Pad8, Pad9 = Code(0x52), Code(0x53), Code(0x54), Code(0x55), Code(0x56), Code(0x57), Code(0x58), Code(0x59), Code(0x5B), Code(0x5C)
	PadClear                                                   = Code(0x47)
	PadDecimal                                                 = Code(0x41)
	PadDivide                                                  = Code(0x4B)
	PadEnter                                                   = Code(0x4C)
	PadEquals                                                  = Code(0x51)
	PadMinus                                                   = Code(0x4E)
	PadMultiply                                                = Code(0x43)
	PadPlus                                                    = Code(0x45)

	/* Layout-Independent Key Codes */

	UpArrow, DownArrow, LeftArrow, RightArrow = Code(0x7E), Code(0x7D), Code(0x7B), Code(0x7C)
	CapsLock, Command, Delete                 = Code(0x39), Code(0x37), Code(0x33)
	Control, Option, Shift                    = Code(0x3B), Code(0x3A), Code(0x38)
	RightControl, RightOption, RightShift     = Code(0x3E), Code(0x3D), Code(0x3C)
	End                                       = Code(0x77)
	Escape                                    = Code(0x35)
	ForwardDelete                             = Code(0x75)
	Function                                  = Code(0x3F)
	Help                                      = Code(0x72)
	Home                                      = Code(0x73)
	PageDown                                  = Code(0x79)
	PageUp                                    = Code(0x74)
	Return                                    = Code(0x24)

	Space = Code(0x31)
	Tab   = Code(0x30)

	/* Function Key Codes */

	F1, F2, F3, F4, F5, F6, F7, F8, F9, F10, F11, F12, F13, F14, F15, F16, F17, F18, F19, F20 = Code(0x7A), Code(0x78), Code(0x63), Code(0x76), Code(0x60), Code(0x61), Code(0x62), Code(0x64), Code(0x65), Code(0x6D), Code(0x67), Code(0x6F), Code(0x69), Code(0x6B), Code(0x71), Code(0x6A), Code(0x40), Code(0x4F), Code(0x50), Code(0x5A)

	/* Volume Key Codes */

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

		//Pad키코드
		Pad0: "Pad0", Pad1: "Pad1", Pad2: "Pad2", Pad3: "Pad3", Pad4: "Pad4",
		Pad5: "Pad5", Pad6: "Pad6", Pad7: "Pad7", Pad8: "Pad8", Pad9: "Pad9",
		PadClear:    "PadClear",
		PadDecimal:  "PadDecimal",
		PadDivide:   "PadDivide",
		PadEnter:    "PadEnter",
		PadEquals:   "PadEquals",
		PadMinus:    "PadMinus",
		PadMultiply: "PadMultiply",
		PadPlus:     "PadPlus",

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
		F1: "F1", F2: "F2", F3: "F3", F4: "F4", F5: "F5", F6: "F6", F7: "F7", F8: "F8", F9: "F9",
		F10: "F10", F11: "F11", F12: "F12", F13: "F13", F14: "F14", F15: "F15", F16: "F16", F17: "F17", F18: "F18", F19: "F19", F20: "F20",

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
		"RightBracket": RightBracket,
		"Minus":        Minus,
		"Period":       Period,
		"Quote":        Quote,
		"Semicolon":    Semicolon,
		"Slash":        Slash,

		//Pad키코드
		"Pad0": Pad0, "Pad1": Pad1, "Pad2": Pad2, "Pad3": Pad3, "Pad4": Pad4,
		"Pad5": Pad5, "Pad6": Pad6, "Pad7": Pad7, "Pad8": Pad8, "Pad9": Pad9,
		"PadClear":    PadClear,
		"PadDecimal":  PadDecimal,
		"PadDivide":   PadDivide,
		"PadEnter":    PadEnter,
		"PadEquals":   PadEquals,
		"PadMinus":    PadMinus,
		"PadMultiply": PadMultiply,
		"PadPlus":     PadPlus,

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
		"F1": F1, "F2": F2, "F3": F3, "F4": F4, "F5": F5, "F6": F6, "F7": F7, "F8": F8, "F9": F9,
		"F10": F10, "F11": F11, "F12": F12, "F13": F13, "F14": F14, "F15": F15, "F16": F16, "F17": F17, "F18": F18, "F19": F19, "F20": F20,

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
