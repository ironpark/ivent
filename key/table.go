package key

import "strings"

// Table represents a table of key states.
// It consists of four 64-bit integers, each bit representing a specific key.
type Table [4]uint64

// tableSize is the number of key codes a Table can hold.
const tableSize = 64 * 4

// slot returns the word index and bit mask of a key code, and false if it is out of range.
func slot(code Code) (int, uint64, bool) {
	if code < 0 || code >= tableSize {
		return 0, 0, false
	}
	return int(code / 64), 1 << uint(code%64), true
}

// MakeTable creates a Table from a list of key codes.
// Codes outside the range of the Table, including side-independent modifiers such as Ctrl, are ignored.
// Use Combo to describe a combination with side-independent modifiers.
func MakeTable(codes ...Code) Table {
	table := Table{}
	for _, code := range codes {
		table.Set(code, true)
	}
	return table
}

// Set marks a key as pressed or released and reports whether the Table changed.
// Codes outside the range of the Table are ignored.
func (kpt *Table) Set(code Code, down bool) bool {
	i, mask, ok := slot(code)
	if !ok {
		return false
	}
	old := kpt[i]
	if down {
		kpt[i] |= mask
	} else {
		kpt[i] &^= mask
	}
	return kpt[i] != old
}

// String returns the pressed keys joined by "+", such as "LEFTCTRL+A".
func (kpt Table) String() string {
	return strings.Join(Names(kpt.PressedKeys()...), "+")
}

// Combo returns the pressed keys as a Combo of physical keys.
func (kpt Table) Combo() Combo {
	return Combo{keys: kpt}
}

// Eq checks if two Tables are equal.
func (kpt Table) Eq(other Table) bool {
	return kpt == other
}

// IsSubsetOf checks if the current Table is a subset of another Table.
func (kpt Table) IsSubsetOf(other Table) bool {
	return (kpt[0]&other[0] == kpt[0]) &&
		(kpt[1]&other[1] == kpt[1]) &&
		(kpt[2]&other[2] == kpt[2]) &&
		(kpt[3]&other[3] == kpt[3])
}

// intersects reports whether the two Tables have any key in common.
func (kpt Table) intersects(other Table) bool {
	return kpt[0]&other[0] != 0 || kpt[1]&other[1] != 0 || kpt[2]&other[2] != 0 || kpt[3]&other[3] != 0
}

// union returns the Table with the keys of both Tables.
func (kpt Table) union(other Table) Table {
	return Table{kpt[0] | other[0], kpt[1] | other[1], kpt[2] | other[2], kpt[3] | other[3]}
}

// without returns the Table with the keys of other removed.
func (kpt Table) without(other Table) Table {
	return Table{kpt[0] &^ other[0], kpt[1] &^ other[1], kpt[2] &^ other[2], kpt[3] &^ other[3]}
}

// IsKeyPressed checks if a specific key is pressed in the Table.
func (kpt Table) IsKeyPressed(keycode Code) bool {
	i, mask, ok := slot(keycode)
	return ok && kpt[i]&mask != 0
}

// PressedKeys returns a list of all pressed keys in the Table.
func (kpt Table) PressedKeys() []Code {
	var keys []Code
	for i, b := range kpt {
		for j := 0; j < 64; j++ {
			if b&(1<<uint(j)) != 0 {
				keys = append(keys, Code(i*64+j))
			}
		}
	}
	return keys
}
