package user

import (
	"errors"
	"strings"
	"testing"
)

func TestNewAvatarURL(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{name: "空は設定なしとして許可", input: "", wantErr: nil},
		{name: "https", input: "https://example.com/a.png", wantErr: nil},
		{name: "http", input: "http://example.com/a.png", wantErr: nil},
		{name: "URL でない", input: "foo", wantErr: ErrInvalidAvatarURL},
		{name: "http(s) 以外のスキーム", input: "ftp://example.com/a.png", wantErr: ErrInvalidAvatarURL},
		{name: "javascript スキームは弾く", input: "javascript:alert(1)", wantErr: ErrInvalidAvatarURL},
		{name: "ホストが無い", input: "https:///a.png", wantErr: ErrInvalidAvatarURL},
		{name: "上限超え", input: "https://example.com/" + strings.Repeat("a", MaxAvatarURLLength), wantErr: ErrAvatarURLTooLong},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewAvatarURL(tt.input)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestAvatarURLIsSet(t *testing.T) {
	empty, err := NewAvatarURL("")
	if err != nil {
		t.Fatal(err)
	}
	if empty.IsSet() {
		t.Error("空なのに IsSet() = true")
	}

	set := mustAvatarURL(t, "https://example.com/a.png")
	if !set.IsSet() {
		t.Error("設定済みなのに IsSet() = false")
	}
}

func TestRebuildAvatarURL(t *testing.T) {
	a := RebuildAvatarURL("not a url")
	if a.String() != "not a url" {
		t.Errorf("String() = %q", a.String())
	}
}
