package user

import (
	"errors"
	"net/url"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

const (
	MaxDescriptionLength = 500
	MaxAvatarURLLength   = 2000
)

var (
	ErrDescriptionTooLong = errors.New("description exceeds max length")
	ErrInvalidAvatarURL   = errors.New("avatar_url must be an http(s) URL")
	ErrAvatarURLTooLong   = errors.New("avatar_url exceeds max length")
)

type User struct {
	id          uuid.UUID
	name        UserName
	description string
	avatarURL   string
	createdAt   time.Time
}

// NewUser は新規ユーザーを作る。id はここで生成する。
// createdAt は呼び出し側から受け取る（Clock 方針）。
func NewUser(name UserName, description, avatarURL string, now time.Time) (*User, error) {
	if err := validateProfile(description, avatarURL); err != nil {
		return nil, err
	}

	return &User{
		id:          uuid.New(),
		name:        name,
		description: description,
		avatarURL:   avatarURL,
		createdAt:   now,
	}, nil
}

// RebuildUser は永続化済みのデータから User を組み立て直す。新規作成時のルールは再実行しない。
func RebuildUser(id uuid.UUID, name UserName, description, avatarURL string, createdAt time.Time) *User {
	return &User{
		id:          id,
		name:        name,
		description: description,
		avatarURL:   avatarURL,
		createdAt:   createdAt,
	}
}

// UpdateProfile はプロフィールを更新する。作成時と同じルールで検証する。
func (u *User) UpdateProfile(description, avatarURL string) error {
	if err := validateProfile(description, avatarURL); err != nil {
		return err
	}
	u.description = description
	u.avatarURL = avatarURL
	return nil
}

func validateProfile(description, avatarURL string) error {
	if utf8.RuneCountInString(description) > MaxDescriptionLength {
		return ErrDescriptionTooLong
	}
	if avatarURL == "" {
		return nil
	}
	if utf8.RuneCountInString(avatarURL) > MaxAvatarURLLength {
		return ErrAvatarURLTooLong
	}
	if !isHTTPURL(avatarURL) {
		return ErrInvalidAvatarURL
	}
	return nil
}

func isHTTPURL(s string) bool {
	u, err := url.Parse(s)
	if err != nil {
		return false
	}
	return (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}

func (u *User) ID() uuid.UUID {
	return u.id
}

func (u *User) Name() UserName {
	return u.name
}

func (u *User) Description() string {
	return u.description
}

func (u *User) AvatarURL() string {
	return u.avatarURL
}

func (u *User) CreatedAt() time.Time {
	return u.createdAt
}
