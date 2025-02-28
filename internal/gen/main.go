package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type KeyCodes struct {
	Category string
	Prefix   string
	Codes    []string
	Names    []string
}

func (k KeyCodes) NameWithPrefix() (names []string) {
	names = make([]string, len(k.Names))
	for i, name := range k.Names {
		names[i] = k.Prefix + name
	}
	return
}
func (k KeyCodes) Split() (KeyCodes, KeyCodes) {
	return KeyCodes{
			Category: k.Category,
			Prefix:   k.Prefix,
			Codes:    k.Codes[:len(k.Names)/2],
			Names:    k.Names[:len(k.Names)/2],
		}, KeyCodes{
			Category: k.Category,
			Prefix:   k.Prefix,
			Codes:    k.Codes[len(k.Names)/2:],
			Names:    k.Names[len(k.Names)/2:],
		}
}

type OsSpecificConfig struct {
	Codes []KeyCodes
	Alias map[string][]string
}
type KeyCodeConfig struct {
	Mac     OsSpecificConfig
	Windows OsSpecificConfig
	Alias   map[string][]string
}

func genConst(keyCodes KeyCodes) (constList string) {
	codeNames := make([]string, len(keyCodes.Codes))
	wrappedCodes := make([]string, len(keyCodes.Codes))
	for i, code := range keyCodes.Codes {
		wrappedCodes[i] = "Code(" + code + ")"
		codeNames[i] = keyCodes.Prefix + keyCodes.Names[i]
	}
	constList += "\t" + strings.Join(codeNames, ", ") + " = " + strings.Join(wrappedCodes, ", ") + "\n"
	return
}

func gen(config OsSpecificConfig, alias map[string][]string, filename string) {
	list := config.Codes
	alias = maps.Clone(alias)
	// Merge aliases
	for k, v := range config.Alias {
		alias[k] = v
	}
	constList := "const ("
	for _, keyCodes := range list {
		if 0 < len(keyCodes.Names) {
			// Const
			constList += "\n"
			constList += "\t// " + keyCodes.Category + "\n"
			if len(keyCodes.Names) <= 10 {
				constList += genConst(keyCodes)
			} else {
				a, b := keyCodes.Split()
				constList += genConst(a)
				constList += genConst(b)
			}
		}
	}
	constList += "\n// KeyCode Aliases\n\n"

	// Check a-Z match
	regex := regexp.MustCompile(`^[A-Za-z][A-Za-z0-9]*$`)
	for _, keyCodes := range list {
		for _, keyName := range keyCodes.NameWithPrefix() {
			if aliases, ok := alias[keyName]; ok {
				//constList += "// " + keyName + "\n"
				for _, aliasName := range aliases {
					if !regex.Match([]byte(aliasName)) {
						continue
					}
					constList += fmt.Sprintf("%s = %s\n", aliasName, keyName)
				}
			}
		}
	}
	constList += ")\n"

	varList := "var (\n"
	varList += "\tcodeToName = map[Code]string{\n"
	for _, keyCodes := range list {
		names := keyCodes.NameWithPrefix()
		for i, name := range names {
			names[i] = fmt.Sprintf("%s:\"%s\"", name, strings.ToUpper(name))
		}
		varList += "\t\t" + strings.Join(names, ",") + ","
		varList += "\n"
	}
	varList += "\t}\n"
	varList += "\tnameToCode = map[string]Code{\n"
	for _, keyCodes := range list {
		names := keyCodes.NameWithPrefix()
		for i, name := range names {
			names[i] = fmt.Sprintf("\"%s\":%s", strings.ToUpper(name), name)
		}
		varList += "\t\t" + strings.Join(names, ",") + ","
		varList += "\n"
	}
	varList += "// KeyCode Aliases\n"
	for _, keyCodes := range list {
		for _, keyName := range keyCodes.NameWithPrefix() {
			if aliases, ok := alias[keyName]; ok {
				for _, aliasName := range aliases {
					buf := &bytes.Buffer{}
					enc := json.NewEncoder(buf)
					enc.SetEscapeHTML(false)
					_ = enc.Encode(aliasName)
					data := strings.ToUpper(strings.TrimSpace(buf.String()))
					varList += fmt.Sprintf("%s:%s", data, keyName) + ","
				}
				varList += "\n"
			}
		}
	}

	varList += "\t}\n"
	varList += ")\n"

	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}

	file.WriteString("// Code generated by internal/gen/main.go DO NOT EDIT.\n")
	file.WriteString("package key\n")
	file.WriteString(constList + varList)
	file.Close()

	exec.Command("gofmt", "-w", filename).Run()
}

func main() {
	configFs, err := os.Open("./keycode.json")
	if err != nil {
		panic(err)
	}
	defer configFs.Close()

	config := KeyCodeConfig{}
	err = json.NewDecoder(configFs).Decode(&config)
	if err != nil {
		panic(err)
	}

	gen(config.Mac, config.Alias, "./codes_gen_darwin.go")
	gen(config.Windows, config.Alias, "./codes_gen_windows.go")
}
