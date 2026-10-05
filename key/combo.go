package key

import (
	"errors"
	"regexp"
	"strings"
	"unicode"
)

// Combo is a set of keys that are held together, such as Ctrl+Shift+A.
// Side-independent modifiers (Ctrl, Alt, Shift, Super) match either the left or the right key.
type Combo struct {
	keys    Table
	mods    uint8 // bit i is set for modifiers[i]
	modKeys Table // left and right keys of every modifier in mods
}

// NewCombo creates a Combo from key codes. Codes that are not valid keys are ignored.
func NewCombo(codes ...Code) Combo {
	var c Combo
	for _, code := range codes {
		if i, ok := modifierIndex(code); ok {
			c.mods |= 1 << i
			c.modKeys = c.modKeys.union(modifiers[i].keys)
			continue
		}
		c.keys.Set(code, true)
	}
	return c
}

// ParseCombo parses a combination of key names joined by "+", such as "Ctrl+Shift+A".
// The "+" key itself can be written at the end or the start, as in "Ctrl++".
// It returns an error wrapping ErrUnknownKey if a key name cannot be resolved.
func ParseCombo(s string) (Combo, error) {
	return ParseComboWith(s, nil)
}

// ParseComboWith is like ParseCombo, but single characters such as "z" or "/" are first passed to
// resolve, which can map them to the key that produces the character on the current keyboard layout.
// Names that resolve does not handle are parsed as usual.
func ParseComboWith(s string, resolve func(r rune) (Code, bool)) (Combo, error) {
	parts := strings.Split(s, "+")
	codes := make([]Code, 0, len(parts))
	for i := 0; i < len(parts); i++ {
		name := strings.TrimSpace(parts[i])
		// Two empty parts in a row come from a literal "+" key.
		if name == "" && i+1 < len(parts) && strings.TrimSpace(parts[i+1]) == "" {
			name = "+"
			i++
		}
		code, err := parseName(name, resolve)
		if err != nil {
			return Combo{}, err
		}
		codes = append(codes, code)
	}
	combo := NewCombo(codes...)
	if combo.IsEmpty() {
		return Combo{}, errors.New("empty key combination")
	}
	return combo, nil
}

// ParseSequence parses key combinations separated by whitespace, such as "Ctrl+K Ctrl+C".
// Spaces around "+" are ignored ("Ctrl + K" is one combination); write the + key as "Plus" in a sequence.
func ParseSequence(s string) ([]Combo, error) {
	return ParseSequenceWith(s, nil)
}

// ParseSequenceWith is like ParseSequence, resolving single characters with resolve (see ParseComboWith).
func ParseSequenceWith(s string, resolve func(r rune) (Code, bool)) ([]Combo, error) {
	fields := strings.Fields(spaceAroundPlus.ReplaceAllString(s, "+"))
	if len(fields) == 0 {
		return nil, errors.New("empty key sequence")
	}
	steps := make([]Combo, len(fields))
	for i, field := range fields {
		c, err := ParseComboWith(field, resolve)
		if err != nil {
			return nil, err
		}
		steps[i] = c
	}
	return steps, nil
}

var spaceAroundPlus = regexp.MustCompile(`\s*\+\s*`)

func parseName(name string, resolve func(r rune) (Code, bool)) (Code, error) {
	if resolve != nil {
		if r := []rune(strings.TrimSpace(name)); len(r) == 1 {
			if code, ok := resolve(unicode.ToLower(r[0])); ok {
				return code, nil
			}
		}
	}
	return Parse(name)
}

// MustParseCombo is like ParseCombo but panics on error.
func MustParseCombo(s string) Combo {
	c, err := ParseCombo(s)
	if err != nil {
		panic(err)
	}
	return c
}

// SideIndependent returns the Combo with physical modifier keys (such as LeftCtrl) replaced by
// side-independent modifiers (such as Ctrl), so it matches either side.
func (c Combo) SideIndependent() Combo {
	for i, m := range modifiers {
		if c.keys.intersects(m.keys) {
			c.keys = c.keys.without(m.keys)
			c.mods |= 1 << i
			c.modKeys = c.modKeys.union(m.keys)
		}
	}
	return c
}

// IsEmpty reports whether the Combo contains no keys.
func (c Combo) IsEmpty() bool {
	return c == Combo{}
}

// Codes returns the keys of the Combo, side-independent modifiers first.
func (c Combo) Codes() []Code {
	var codes []Code
	for i := range modifiers {
		if c.mods&(1<<i) != 0 {
			codes = append(codes, Code(tableSize+i))
		}
	}
	return append(codes, c.keys.PressedKeys()...)
}

// String returns the Combo in the form accepted by ParseCombo, such as "CTRL+SHIFT+A".
func (c Combo) String() string {
	return strings.Join(Names(c.Codes()...), "+")
}

// Match reports whether state holds exactly the keys of the Combo.
// An empty Combo never matches.
func (c Combo) Match(state Table) bool {
	rest, ok := c.consumeModifiers(state)
	return ok && rest == c.keys
}

// MatchSubset reports whether state holds the keys of the Combo, possibly with other keys.
// An empty Combo never matches.
func (c Combo) MatchSubset(state Table) bool {
	rest, ok := c.consumeModifiers(state)
	return ok && c.keys.IsSubsetOf(rest)
}

// Covers reports whether every key held in state belongs to the Combo, so that state can still
// become the Combo by pressing more keys. For example, LeftCtrl alone is covered by Ctrl+C.
func (c Combo) Covers(state Table) bool {
	return state.without(c.keys.union(c.modKeys)) == Table{}
}

// consumeModifiers checks that a left or right key is held for each side-independent modifier,
// and returns state without those keys so the remaining keys can be compared.
func (c Combo) consumeModifiers(state Table) (Table, bool) {
	if c.IsEmpty() {
		return state, false
	}
	if c.mods == 0 {
		return state, true
	}
	for i := range modifiers {
		if c.mods&(1<<i) != 0 && !state.intersects(modifiers[i].keys) {
			return state, false
		}
	}
	return state.without(c.modKeys), true
}
