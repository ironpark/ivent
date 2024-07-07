package key

import "strings"

type Code int

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
