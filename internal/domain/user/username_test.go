package user

import (
	"errors"
	"strings"
	"testing"
)

func TestNewUserName(t *testing.T) {
	maxName := strings.Repeat("a", MaxUsernameLength)

	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{name: "空", input: "", wantErr: ErrEmptyUsername},
		{name: "空白のみ", input: "   ", wantErr: ErrEmptyUsername},
		{name: "通常", input: "alice_01", wantErr: nil},
		{name: "上限ちょうど", input: maxName, wantErr: nil},
		{name: "上限超え", input: maxName + "a", wantErr: ErrUsernameTooLong},
		{name: "日本語は不可", input: "たろう", wantErr: ErrInvalidUsernameChars},
		{name: "空白を含むと不可", input: "al ice", wantErr: ErrInvalidUsernameChars},
		{name: "記号は不可", input: "alice!", wantErr: ErrInvalidUsernameChars},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewUserName(tt.input)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestRebuildUserName_SkipsValidation(t *testing.T) {
	n := RebuildUserName("")
	if n.String() != "" {
		t.Errorf("String() = %q", n.String())
	}
}
