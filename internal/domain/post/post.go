package post

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrMissingUserID = errors.New("post must have a user id")

type Post struct {
	id        uuid.UUID
	userID    uuid.UUID
	content   PostContent
	isPublic  bool
	createdAt time.Time
}

// NewPost は新規投稿を作る。id はここで生成する。
// createdAt は呼び出し側から受け取る。テストで時刻を固定できるようにするため。
func NewPost(userID uuid.UUID, content PostContent, isPublic bool, now time.Time) (*Post, error) {
	if userID == uuid.Nil {
		return nil, ErrMissingUserID
	}

	return &Post{
		id:        uuid.New(),
		userID:    userID,
		content:   content,
		isPublic:  isPublic,
		createdAt: now,
	}, nil
}

// RebuildPost は永続化済みのデータから Post を組み立て直す。新規作成時のルールは再実行しない。
func RebuildPost(id, userID uuid.UUID, content PostContent, isPublic bool, createdAt time.Time) *Post {
	return &Post{
		id:        id,
		userID:    userID,
		content:   content,
		isPublic:  isPublic,
		createdAt: createdAt,
	}
}

func (p *Post) ID() uuid.UUID {
	return p.id
}

func (p *Post) UserID() uuid.UUID {
	return p.userID
}

func (p *Post) Content() PostContent {
	return p.content
}

func (p *Post) IsPublic() bool {
	return p.isPublic
}

func (p *Post) CreatedAt() time.Time {
	return p.createdAt
}
