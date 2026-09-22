package user

import (
	"errors"
	"strings"
	"testing"
)

func TestNewDescription(t *testing.T) {
	maxDesc := strings.Repeat("あ", MaxDescriptionLength)

	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{name: "空は許可", input: "", wantErr: nil},
		{name: "通常", input: "hello", wantErr: nil},
		{name: "上限ちょうど（文字数で数える）", input: maxDesc, wantErr: nil},
		{name: "上限超え", input: maxDesc + "あ", wantErr: ErrDescriptionTooLong},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewDescription(tt.input)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestDescriptionIsSet(t *testing.T) {
	empty := mustDescription(t, "")
	if empty.IsSet() {
		t.Error("空なのに IsSet() = true")
	}

	set := mustDescription(t, "hello")
	if !set.IsSet() {
		t.Error("設定済みなのに IsSet() = false")
	}
}

func TestRebuildDescription(t *testing.T) {
	d := RebuildDescription(strings.Repeat("あ", MaxDescriptionLength+1))
	if !d.IsSet() {
		t.Error("IsSet() = false")
	}
}
