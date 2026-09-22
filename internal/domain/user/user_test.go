package user

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

var fixedNow = time.Date(2026, 9, 23, 10, 0, 0, 0, time.UTC)

func mustUserName(t *testing.T, s string) UserName {
	t.Helper()
	n, err := NewUserName(s)
	if err != nil {
		t.Fatal(err)
	}
	return n
}

func TestNewUser(t *testing.T) {
	name := mustUserName(t, "alice")

	tests := []struct {
		label       string
		description string
		avatarURL   string
		wantErr     error
	}{
		{label: "プロフィール空", description: "", avatarURL: "", wantErr: nil},
		{label: "通常", description: "hello", avatarURL: "https://example.com/a.png", wantErr: nil},
		{label: "description 上限超え", description: strings.Repeat("あ", MaxDescriptionLength+1), avatarURL: "", wantErr: ErrDescriptionTooLong},
		{label: "avatar_url が URL でない", description: "", avatarURL: "foo", wantErr: ErrInvalidAvatarURL},
		{label: "avatar_url が http(s) 以外", description: "", avatarURL: "ftp://example.com/a.png", wantErr: ErrInvalidAvatarURL},
		{label: "avatar_url 上限超え", description: "", avatarURL: "https://example.com/" + strings.Repeat("a", MaxAvatarURLLength), wantErr: ErrAvatarURLTooLong},
	}

	for _, tt := range tests {
		t.Run(tt.label, func(t *testing.T) {
			_, err := NewUser(name, tt.description, tt.avatarURL, fixedNow)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewUser_idが生成されcreatedAtは渡した時刻になる(t *testing.T) {
	u, err := NewUser(mustUserName(t, "alice"), "", "", fixedNow)
	if err != nil {
		t.Fatal(err)
	}
	if u.ID() == uuid.Nil {
		t.Error("ID() が生成されていない")
	}
	if !u.CreatedAt().Equal(fixedNow) {
		t.Errorf("CreatedAt() = %v, want %v", u.CreatedAt(), fixedNow)
	}
}

func TestUser_UpdateProfile(t *testing.T) {
	u, err := NewUser(mustUserName(t, "alice"), "before", "", fixedNow)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("不正な値なら変更されない", func(t *testing.T) {
		if err := u.UpdateProfile("after", "foo"); !errors.Is(err, ErrInvalidAvatarURL) {
			t.Fatalf("err = %v, want %v", err, ErrInvalidAvatarURL)
		}
		if u.Description() != "before" {
			t.Errorf("Description() = %q, want %q", u.Description(), "before")
		}
	})

	t.Run("正しい値なら変更される", func(t *testing.T) {
		if err := u.UpdateProfile("after", "https://example.com/a.png"); err != nil {
			t.Fatal(err)
		}
		if u.Description() != "after" {
			t.Errorf("Description() = %q", u.Description())
		}
		if u.AvatarURL() != "https://example.com/a.png" {
			t.Errorf("AvatarURL() = %q", u.AvatarURL())
		}
	})
}
