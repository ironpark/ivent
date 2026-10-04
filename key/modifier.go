package key

// Side-independent modifiers. Each one matches either its left or its right physical key.
// They are outside the range of a Table and only have meaning in a Combo.
const (
	Ctrl Code = tableSize + iota
	Alt
	Shift
	Super

	Option = Alt
	Cmd    = Super
	Win    = Super
)

type modifier struct {
	name        string
	left, right Code
	keys        Table // left and right keys
}

// modifiers is indexed by Code - tableSize.
var modifiers = [...]modifier{
	{name: "CTRL", left: LeftCtrl, right: RightCtrl},
	{name: "ALT", left: LeftAlt, right: RightAlt},
	{name: "SHIFT", left: LeftShift, right: RightShift},
	{name: "SUPER", left: LeftSuper, right: RightSuper},
}

func init() {
	// Register the side-independent modifiers in the generated name tables,
	// so Name and Parse resolve them like any other key.
	for i := range modifiers {
		m := &modifiers[i]
		m.keys = MakeTable(m.left, m.right)
		code := Code(tableSize + i)
		codeToName[code] = m.name
		nameToCode[m.name] = code
	}
	nameToCode["OPTION"] = Option
	nameToCode["CMD"] = Cmd
	nameToCode["WIN"] = Win
}

// modifierIndex returns the index of a side-independent modifier in modifiers, and false if code is not one.
func modifierIndex(code Code) (int, bool) {
	i := int(code - tableSize)
	return i, i >= 0 && i < len(modifiers)
}
