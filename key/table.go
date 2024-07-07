package key

import "fmt"

// Table represents a table of key states.
// It consists of four 64-bit integers, each bit representing a specific key.
type Table [4]uint64

// MakeTable creates a Table from a list of key codes.
func MakeTable(codes ...Code) Table {
	table := Table{}
	for _, code := range codes {
		byteIndex := code / 64
		bitIndex := uint(code % 64)
		mask := uint64(1 << bitIndex)
		table[byteIndex] |= mask
	}
	return table
}

func (kpt Table) String() string {
	str := ""
	for _, code := range kpt.PressedKeys() {
		str += fmt.Sprintf("%s ", code.Name())
	}
	return str
}

// Eq checks if two Tables are equal.
func (kpt Table) Eq(other Table) bool {
	return kpt[0] == other[0] && kpt[1] == other[1] && kpt[2] == other[2] && kpt[3] == other[3]
}

// IsSubsetOf checks if the current Table is a subset of another Table.
func (kpt Table) IsSubsetOf(other Table) bool {
	return (kpt[0]&other[0] == kpt[0]) &&
		(kpt[1]&other[1] == kpt[1]) &&
		(kpt[2]&other[2] == kpt[2]) &&
		(kpt[3]&other[3] == kpt[3])
}

// IsKeyPressed checks if a specific key is pressed in the Table.
func (kpt Table) IsKeyPressed(keycode Code) bool {
	byteIndex := keycode / 64
	bitIndex := uint(keycode % 64)
	mask := uint64(1 << bitIndex)
	return kpt[byteIndex]&mask != 0
}

// CheckKeyCombination checks if all the specified key codes are pressed in the Table.
func (kpt Table) CheckKeyCombination(keyCodes ...Code) bool {
	for _, key := range keyCodes {
		if !kpt.IsKeyPressed(key) {
			return false
		}
	}
	return true
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
