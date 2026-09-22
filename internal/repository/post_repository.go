package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"sns/internal/db/sqlcgen"
	"sns/internal/domain/post"
)

// PostgreSQL のエラーコード(SQLSTATE)。
const (
	pgForeignKeyViolation = "23503"
	pgUniqueViolation     = "23505"
)

// post.Repository を満たしているかコンパイル時に検証する。
var _ post.Repository = (*PostRepository)(nil)

type PostRepository struct {
	q *sqlcgen.Queries
}

func NewPostRepository(db sqlcgen.DBTX) *PostRepository {
	return &PostRepository{q: sqlcgen.New(db)}
}

func (r *PostRepository) Create(ctx context.Context, p *post.Post) (*post.Post, error) {
	row, err := r.q.CreatePost(ctx, sqlcgen.CreatePostParams{
		ID:        toPgUUID(p.ID()),
		UserID:    toPgUUID(p.UserID()),
		Body:      toPgText(p.Content().Body()),
		ImageUrl:  toPgText(p.Content().ImageURL()),
		IsPublic:  p.IsPublic(),
		CreatedAt: toPgTimestamptz(p.CreatedAt()),
	})
	if err != nil {
		return nil, translateCreateError(p.ID(), err)
	}
	return toDomain(row), nil
}

func (r *PostRepository) FindByID(ctx context.Context, id uuid.UUID) (*post.Post, error) {
	row, err := r.q.GetPostByID(ctx, toPgUUID(id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("find post %s: %w", id, post.ErrPostNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("find post %s: %w", id, err)
	}
	return toDomain(row), nil
}

// translateCreateError は pgx のエラーをドメインのエラーに翻訳する。
// 元のエラーもチェーンに残すので、ログでは DB の全文が追える。
func translateCreateError(id uuid.UUID, err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgForeignKeyViolation:
			return fmt.Errorf("create post %s: %w: %w", id, post.ErrUserNotFound, err)
		case pgUniqueViolation:
			return fmt.Errorf("create post %s: %w: %w", id, post.ErrPostAlreadyExists, err)
		}
	}
	return fmt.Errorf("create post %s: %w", id, err)
}

// toDomain は DB の行をドメインの Post に変換する。保存済みデータは検証済みとみなし、ルールは再実行しない。
func toDomain(row sqlcgen.Post) *post.Post {
	return post.RebuildPost(
		fromPgUUID(row.ID),
		fromPgUUID(row.UserID),
		post.RebuildPostContent(fromPgText(row.Body), fromPgText(row.ImageUrl)),
		row.IsPublic,
		fromPgTimestamptz(row.CreatedAt),
	)
}
