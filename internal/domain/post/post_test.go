package post

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewPost(t *testing.T) {
	content, err := NewPostContent("hello", "")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 9, 22, 10, 0, 0, 0, time.UTC)

	t.Run("userID が空ならエラー", func(t *testing.T) {
		_, err := NewPost(uuid.Nil, content, true, now)
		if !errors.Is(err, ErrMissingUserID) {
			t.Fatalf("err = %v, want %v", err, ErrMissingUserID)
		}
	})

	t.Run("id が生成され、createdAt は渡した時刻になる", func(t *testing.T) {
		p, err := NewPost(uuid.New(), content, true, now)
		if err != nil {
			t.Fatal(err)
		}
		if p.ID() == uuid.Nil {
			t.Error("ID() が生成されていない")
		}
		if !p.CreatedAt().Equal(now) {
			t.Errorf("CreatedAt() = %v, want %v", p.CreatedAt(), now)
		}
	})
}
