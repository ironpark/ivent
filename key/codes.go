package key

import "strings"

type Code int

func init() {
	// Add Aliases
	nameToCode["["] = LeftBracket
	nameToCode["]"] = RightBracket
	nameToCode["\\"] = Backslash
	nameToCode[";"] = Semicolon
	nameToCode[","] = Comma
	nameToCode["/"] = Slash
	nameToCode[" "] = SpaceBar
	nameToCode["`"] = Grave
	nameToCode["-"] = Minus
	nameToCode["="] = Equal
	nameToCode["'"] = Quote
	nameToCode["."] = Dot
	nameToCode["CMD"] = Super
}

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
		codes[i] = nameToCode[strings.ToUpper(name)]
	}
	return codes
}
