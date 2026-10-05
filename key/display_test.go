package key

import (
	"runtime"
	"testing"
)

func TestFormatText(t *testing.T) {
	super, alt := "Super", "Alt"
	switch runtime.GOOS {
	case "darwin":
		super, alt = "Cmd", "Option"
	case "windows":
		super = "Win"
	}
	tests := []struct {
		combo Combo
		want  string
	}{
		{MustParseCombo("Shift+A+Ctrl"), "Ctrl+Shift+A"},
		{MustParseCombo("Cmd+Alt+LeftBracket"), alt + "+" + super + "+["},
		{MustParseCombo("RightCtrl+Num1"), "Right Ctrl+1"},
		{MustParseCombo("Pad1"), "Numpad 1"},
		{MustParseCombo("Ctrl+PageUp"), "Ctrl+Page Up"},
		{MustParseCombo("Ctrl+MouseLeft"), "Ctrl+Left Click"},
	}
	for _, tt := range tests {
		if got := tt.combo.Format(StyleText); got != tt.want {
			t.Errorf("Format(%s) = %q, want %q", tt.combo, got, tt.want)
		}
	}
}

func TestFormatSymbols(t *testing.T) {
	if got := MustParseCombo("Cmd+Shift+Ctrl+Alt+Enter").Format(StyleSymbols); got != "⌃⌥⇧⌘↩" {
		t.Fatalf("got %q", got)
	}
	if got := MustParseCombo("Cmd+A").Format(StyleSymbols); got != "⌘A" {
		t.Fatalf("got %q", got)
	}
}

func TestSideIndependent(t *testing.T) {
	c := NewCombo(LeftCtrl, RightShift, A).SideIndependent()
	if c != NewCombo(Ctrl, Shift, A) {
		t.Fatalf("SideIndependent = %s", c)
	}
	if MakeTable(LeftCtrl, A).Combo() != NewCombo(LeftCtrl, A) {
		t.Fatal("Table.Combo")
	}
	if got := MakeTable(A).String(); got != "A" {
		t.Fatalf("Table.String = %q", got)
	}
}

func TestParseComboWith(t *testing.T) {
	// A layout where the key at the US "W" position types "z" (AZERTY).
	azerty := func(r rune) (Code, bool) {
		if r == 'z' {
			return W, true
		}
		return 0, false
	}
	c, err := ParseComboWith("Ctrl+Z", azerty)
	if err != nil || c != NewCombo(Ctrl, W) {
		t.Fatalf("ParseComboWith = %s, %v", c, err)
	}
	// Names that are not single characters are not resolved.
	if c, _ := ParseComboWith("Ctrl+Enter", azerty); c != NewCombo(Ctrl, Enter) {
		t.Fatalf("got %s", c)
	}
}

func TestIsModifierAgain(t *testing.T) {
	for _, c := range []Code{Ctrl, LeftCtrl, RightSuper} {
		if !IsModifier(c) {
			t.Errorf("IsModifier(%s) = false", Name(c))
		}
	}
	if IsModifier(A) || IsModifier(MouseLeft) {
		t.Error("IsModifier")
	}
}
