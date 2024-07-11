package key

import (
	"fmt"
	"strconv"
	"strings"
)

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
	name, ok := codeToName[code]
	if ok {
		return name
	}
	return fmt.Sprintf("UNK_%d", code)
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
		if after, found := strings.CutPrefix(name, "UNK_"); found {
			code, err := strconv.Atoi(after)
			if err != nil {
				codes[i] = Code(code)
			} else {
				codes[i] = Code(255)
			}
			continue
		}
		codes[i] = nameToCode[strings.ToUpper(name)]
	}
	return codes
}
