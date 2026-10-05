package key

import (
	"errors"
	"testing"
)

func TestParseCombo(t *testing.T) {
	tests := []struct {
		in   string
		want Combo
	}{
		{"Ctrl+Shift+A", NewCombo(Ctrl, Shift, A)},
		{" cmd + a ", NewCombo(Super, A)},
		{"Ctrl++", NewCombo(Ctrl, Equal)},
		{"+", NewCombo(Equal)},
		{"RightCtrl+C", NewCombo(RightCtrl, C)},
	}
	for _, tt := range tests {
		got, err := ParseCombo(tt.in)
		if err != nil {
			t.Errorf("ParseCombo(%q) error: %v", tt.in, err)
			continue
		}
		if got != tt.want {
			t.Errorf("ParseCombo(%q) = %s, want %s", tt.in, got, tt.want)
		}
	}

	for _, in := range []string{"", "Ctrl+", "Ctrl+Shitf", "A+ +B"} {
		if _, err := ParseCombo(in); err == nil {
			t.Errorf("ParseCombo(%q) expected error", in)
		}
	}
	if _, err := ParseCombo("Ctrl+Shitf"); !errors.Is(err, ErrUnknownKey) {
		t.Errorf("expected ErrUnknownKey, got %v", err)
	}
}

func TestComboString(t *testing.T) {
	c := NewCombo(A, Shift, Ctrl)
	if got := c.String(); got != "CTRL+SHIFT+A" {
		t.Fatalf("String() = %q", got)
	}
	back, err := ParseCombo(c.String())
	if err != nil || back != c {
		t.Fatalf("round trip failed: %v %v", back, err)
	}
}

func TestComboMatch(t *testing.T) {
	c := NewCombo(Ctrl, C)
	cases := []struct {
		state         Table
		exact, subset bool
	}{
		{MakeTable(LeftCtrl, C), true, true},
		{MakeTable(RightCtrl, C), true, true},
		{MakeTable(LeftCtrl, RightCtrl, C), true, true},
		{MakeTable(LeftCtrl, LeftShift, C), false, true},
		{MakeTable(C), false, false},
		{MakeTable(LeftCtrl), false, false},
	}
	for _, tt := range cases {
		if got := c.Match(tt.state); got != tt.exact {
			t.Errorf("Match(%v) = %v, want %v", tt.state, got, tt.exact)
		}
		if got := c.MatchSubset(tt.state); got != tt.subset {
			t.Errorf("MatchSubset(%v) = %v, want %v", tt.state, got, tt.subset)
		}
	}
}

func TestEmptyComboNeverMatches(t *testing.T) {
	var c Combo
	if c.Match(MakeTable(A)) || c.MatchSubset(MakeTable(A)) {
		t.Fatal("empty Combo must not match")
	}
}

func TestModifierNames(t *testing.T) {
	if Name(Ctrl) != "CTRL" || Name(LeftCtrl) != "LEFTCTRL" {
		t.Fatalf("names: %s %s", Name(Ctrl), Name(LeftCtrl))
	}
	for _, name := range []string{"option", "CMD", "Win"} {
		if _, err := Parse(name); err != nil {
			t.Errorf("Parse(%q): %v", name, err)
		}
	}
}

func TestParseSequence(t *testing.T) {
	steps, err := ParseSequence("Ctrl+K  Ctrl+C")
	if err != nil {
		t.Fatal(err)
	}
	if len(steps) != 2 || steps[0] != NewCombo(Ctrl, K) || steps[1] != NewCombo(Ctrl, C) {
		t.Fatalf("ParseSequence = %v", steps)
	}
	for _, in := range []string{"", "  ", "Ctrl+K Nope"} {
		if _, err := ParseSequence(in); err == nil {
			t.Errorf("ParseSequence(%q) expected error", in)
		}
	}
}

func TestCovers(t *testing.T) {
	c := NewCombo(Ctrl, C)
	for _, state := range []Table{{}, MakeTable(LeftCtrl), MakeTable(RightCtrl, C)} {
		if !c.Covers(state) {
			t.Errorf("Covers(%v) = false", state)
		}
	}
	if c.Covers(MakeTable(LeftShift)) || c.Covers(MakeTable(LeftCtrl, A)) {
		t.Error("keys outside the combination must not be covered")
	}
	if Physical(Ctrl) != LeftCtrl || Physical(A) != A {
		t.Fatal("Physical")
	}
}
