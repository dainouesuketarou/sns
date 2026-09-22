package user

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	id          uuid.UUID
	name        UserName
	description Description
	avatarURL   AvatarURL
	createdAt   time.Time
}

// NewUser は新規ユーザーを作る。id はここで生成する。
// 各フィールドのルールは値オブジェクト側で検証済みなので、ここでは検証しない。
// createdAt は呼び出し側から受け取る（Clock 方針）。
func NewUser(name UserName, description Description, avatarURL AvatarURL, now time.Time) *User {
	return &User{
		id:          uuid.New(),
		name:        name,
		description: description,
		avatarURL:   avatarURL,
		createdAt:   now,
	}
}

// RebuildUser は永続化済みのデータから User を組み立て直す。新規作成時のルールは再実行しない。
func RebuildUser(id uuid.UUID, name UserName, description Description, avatarURL AvatarURL, createdAt time.Time) *User {
	return &User{
		id:          id,
		name:        name,
		description: description,
		avatarURL:   avatarURL,
		createdAt:   createdAt,
	}
}

// UpdateProfile はプロフィールを更新する。
func (u *User) UpdateProfile(description Description, avatarURL AvatarURL) {
	u.description = description
	u.avatarURL = avatarURL
}

func (u *User) ID() uuid.UUID {
	return u.id
}

func (u *User) Name() UserName {
	return u.name
}

func (u *User) Description() Description {
	return u.description
}

func (u *User) AvatarURL() AvatarURL {
	return u.avatarURL
}

func (u *User) CreatedAt() time.Time {
	return u.createdAt
}
