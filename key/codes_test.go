package key

import (
	"errors"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name string
		want Code
		err  bool
	}{
		{"A", A, false},
		{" ctrl ", Ctrl, false},
		{"Cmd", Super, false},
		{"Unk200", Code(200), false},
		{"Unk300", Invalid, true},
		{"UnkX", Invalid, true},
		{"Shitf", Invalid, true},
		{"", Invalid, true},
	}
	for _, tt := range tests {
		got, err := Parse(tt.name)
		if (err != nil) != tt.err {
			t.Errorf("Parse(%q) error = %v, want error %v", tt.name, err, tt.err)
		}
		if err != nil && !errors.Is(err, ErrUnknownKey) {
			t.Errorf("Parse(%q) error = %v, want ErrUnknownKey", tt.name, err)
		}
		if got != tt.want {
			t.Errorf("Parse(%q) = %d, want %d", tt.name, got, tt.want)
		}
	}
}

func TestTableOutOfRange(t *testing.T) {
	table := MakeTable(Code(-1), Code(256), A)
	if !table.Eq(MakeTable(A)) {
		t.Fatalf("out of range codes must be ignored: %v", table)
	}
	if table.IsKeyPressed(Code(1000)) {
		t.Fatal("out of range code must not be pressed")
	}
}
