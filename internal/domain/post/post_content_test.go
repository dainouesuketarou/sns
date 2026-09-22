package post

import (
	"errors"
	"strings"
	"testing"
)

func TestNewPostContent(t *testing.T) {
	maxBody := strings.Repeat("あ", MaxBodyLength)

	tests := []struct {
		name     string
		body     string
		imageURL string
		wantErr  error
	}{
		{name: "本文も画像も空", body: "", imageURL: "", wantErr: ErrEmptyPostContent},
		{name: "空白のみは無し扱い", body: "   ", imageURL: " ", wantErr: ErrEmptyPostContent},
		{name: "本文のみ", body: "hello", imageURL: "", wantErr: nil},
		{name: "画像のみ", body: "", imageURL: "https://example.com/a.png", wantErr: nil},
		{name: "本文と画像の両方", body: "hello", imageURL: "https://example.com/a.png", wantErr: nil},
		{name: "本文が上限ちょうど(文字数で数える)", body: maxBody, imageURL: "", wantErr: nil},
		{name: "本文が上限超え", body: maxBody + "あ", imageURL: "", wantErr: ErrPostBodyTooLong},
		{name: "画像がURLでない", body: "", imageURL: "foo", wantErr: ErrInvalidImageURL},
		{name: "画像が http(s) 以外", body: "", imageURL: "ftp://example.com/a.png", wantErr: ErrInvalidImageURL},
		{name: "画像にホストが無い", body: "", imageURL: "https:///a.png", wantErr: ErrInvalidImageURL},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewPostContent(tt.body, tt.imageURL)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
		})
	}

	t.Run("本文の前後の空白を保持する", func(t *testing.T) {
		c, err := NewPostContent("  hello\n", "")
		if err != nil {
			t.Fatal(err)
		}
		if c.Body() != "  hello\n" {
			t.Errorf("Body() = %q", c.Body())
		}
	})
}

func TestRebuildPostContent(t *testing.T) {
	c := RebuildPostContent("", "not a url")
	if c.Body() != "" || c.ImageURL() != "not a url" {
		t.Errorf("Body() = %q, ImageURL() = %q", c.Body(), c.ImageURL())
	}
}
