package user

import (
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

func mustDescription(t *testing.T, s string) Description {
	t.Helper()
	d, err := NewDescription(s)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func mustAvatarURL(t *testing.T, s string) AvatarURL {
	t.Helper()
	a, err := NewAvatarURL(s)
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func TestNewUser_idが生成されcreatedAtは渡した時刻になる(t *testing.T) {
	u := NewUser(mustUserName(t, "alice"), Description{}, AvatarURL{}, fixedNow)

	if u.ID() == uuid.Nil {
		t.Error("ID() が生成されていない")
	}
	if !u.CreatedAt().Equal(fixedNow) {
		t.Errorf("CreatedAt() = %v, want %v", u.CreatedAt(), fixedNow)
	}
}

func TestUser_UpdateProfile(t *testing.T) {
	u := NewUser(mustUserName(t, "alice"), mustDescription(t, "before"), AvatarURL{}, fixedNow)

	u.UpdateProfile(mustDescription(t, "after"), mustAvatarURL(t, "https://example.com/a.png"))

	if u.Description().String() != "after" {
		t.Errorf("Description() = %q", u.Description().String())
	}
	if u.AvatarURL().String() != "https://example.com/a.png" {
		t.Errorf("AvatarURL() = %q", u.AvatarURL().String())
	}
}
