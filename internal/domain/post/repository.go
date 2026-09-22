package post

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var (
	ErrPostNotFound      = errors.New("post not found")
	ErrPostAlreadyExists = errors.New("post already exists")
	// ErrUserNotFound は投稿者が存在しないときのエラー。user 集約ができたらそちらへ移す。
	ErrUserNotFound = errors.New("user not found")
)

// Repository は投稿集約を保存・取得する窓口。実装はドメイン層の外に置く。
type Repository interface {
	// Create は新規投稿を INSERT し、DB に保存された状態の Post を返す。更新は別メソッドにする。
	Create(ctx context.Context, p *Post) (*Post, error)
	FindByID(ctx context.Context, id uuid.UUID) (*Post, error)
}
