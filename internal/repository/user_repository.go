package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"sns/internal/db/sqlcgen"
	"sns/internal/domain/user"
)

// user.Repository を満たしているかコンパイル時に検証する。
var _ user.Repository = (*UserRepository)(nil)

type UserRepository struct {
	q *sqlcgen.Queries
}

func NewUserRepository(db sqlcgen.DBTX) *UserRepository {
	return &UserRepository{q: sqlcgen.New(db)}
}

func (r *UserRepository) Create(ctx context.Context, u *user.User) (*user.User, error) {
	row, err := r.q.CreateUser(ctx, sqlcgen.CreateUserParams{
		ID:          toPgUUID(u.ID()),
		Username:    u.Name().String(),
		Description: toPgText(u.Description()),
		AvatarUrl:   toPgText(u.AvatarURL()),
		CreatedAt:   toPgTimestamptz(u.CreatedAt()),
	})
	if err != nil {
		return nil, translateCreateUserError(u.ID(), err)
	}
	return toDomainUser(row), nil
}

func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	row, err := r.q.GetUserByID(ctx, toPgUUID(id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("find user %s: %w", id, user.ErrUserNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("find user %s: %w", id, err)
	}
	return toDomainUser(row), nil
}

// translateCreateUserError は pgx のエラーをドメインのエラーに翻訳する。
// users の UNIQUE は今 username だけなので 23505 = username 重複と断定できる。
// email などに UNIQUE を足したら pgErr.ConstraintName で分岐する。
func translateCreateUserError(id uuid.UUID, err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
		return fmt.Errorf("create user %s: %w: %w", id, user.ErrUsernameAlreadyExists, err)
	}
	return fmt.Errorf("create user %s: %w", id, err)
}

// toDomainUser は DB の行をドメインの User に変換する。保存済みデータは検証済みとみなす。
func toDomainUser(row sqlcgen.User) *user.User {
	return user.RebuildUser(
		fromPgUUID(row.ID),
		user.RebuildUserName(row.Username),
		fromPgText(row.Description),
		fromPgText(row.AvatarUrl),
		fromPgTimestamptz(row.CreatedAt),
	)
}
