package key

import (
	"fmt"
	"strconv"
	"strings"
)

type Code int

func (k Code) Name() string {
	return codeToName[k]
}

func Name(code Code) string {
	name, ok := codeToName[code]
	if ok {
		return name
	}
	return fmt.Sprintf("Unk%d", code)
}

// Names converts a slice of key codes to a slice of key names.
// If the key code is not found, it will be named as "Unk" followed by the code.
func Names(codes ...Code) []string {
	names := make([]string, len(codes))
	for i, code := range codes {
		names[i] = Name(code)
	}
	return names
}

// Codes converts a slice of key names to a slice of key codes.
func Codes(names ...string) []Code {
	codes := make([]Code, len(names))
	for i, name := range names {
		codes[i] = toCode(name)
	}
	return codes
}

func toCode(name string) Code {
	name = strings.ToUpper(name)
	if after, found := strings.CutPrefix(name, "UNK"); found {
		code, err := strconv.Atoi(after)
		if err != nil {
			return Code(code)
		}
		return Code(255)
	}
	return nameToCode[strings.ToUpper(name)]
}
