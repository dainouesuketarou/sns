package user

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var (
	// ErrUserNotFound は検索して見つからないときのエラー。
	ErrUserNotFound = errors.New("user not found")
	// ErrUsernameAlreadyExists は username の UNIQUE 制約違反の翻訳先。
	ErrUsernameAlreadyExists = errors.New("username already exists")
)

// Repository はユーザー集約を保存・取得する窓口。実装はドメイン層の外に置く。
type Repository interface {
	// Create は新規ユーザーを INSERT し、DB に保存された状態の User を返す。
	Create(ctx context.Context, u *User) (*User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
}
