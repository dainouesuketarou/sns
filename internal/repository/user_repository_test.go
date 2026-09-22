package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"sns/internal/domain/user"
)

func newUser(t *testing.T, username, description, avatarURL string) *user.User {
	t.Helper()

	name, err := user.NewUserName(username)
	if err != nil {
		t.Fatal(err)
	}
	desc, err := user.NewDescription(description)
	if err != nil {
		t.Fatal(err)
	}
	avatar, err := user.NewAvatarURL(avatarURL)
	if err != nil {
		t.Fatal(err)
	}
	return user.NewUser(name, desc, avatar, fixedNow)
}

func assertSameUser(t *testing.T, got, want *user.User) {
	t.Helper()

	if got.ID() != want.ID() {
		t.Errorf("ID() = %v, want %v", got.ID(), want.ID())
	}
	if got.Name() != want.Name() {
		t.Errorf("Name() = %v, want %v", got.Name(), want.Name())
	}
	if got.Description() != want.Description() {
		t.Errorf("Description() = %q, want %q", got.Description().String(), want.Description().String())
	}
	if got.AvatarURL() != want.AvatarURL() {
		t.Errorf("AvatarURL() = %q, want %q", got.AvatarURL().String(), want.AvatarURL().String())
	}
	// timestamptz はマイクロ秒精度なので、Go 側のナノ秒を切り捨ててから比較する。
	if !got.CreatedAt().Equal(want.CreatedAt().Truncate(time.Microsecond)) {
		t.Errorf("CreatedAt() = %v, want %v", got.CreatedAt(), want.CreatedAt())
	}
}

func TestUserRepository_CreateしてFindByIDで往復できる(t *testing.T) {
	ctx := context.Background()
	repo := NewUserRepository(newTestTx(t))
	u := newUser(t, "alice_01", "hello", "https://example.com/a.png")

	created, err := repo.Create(ctx, u)
	if err != nil {
		t.Fatal(err)
	}
	assertSameUser(t, created, u)

	got, err := repo.FindByID(ctx, u.ID())
	if err != nil {
		t.Fatal(err)
	}
	assertSameUser(t, got, u)
}

func TestUserRepository_プロフィール空ならNULLで往復する(t *testing.T) {
	ctx := context.Background()
	tx := newTestTx(t)
	repo := NewUserRepository(tx)
	u := newUser(t, "bob_02", "", "")

	if _, err := repo.Create(ctx, u); err != nil {
		t.Fatal(err)
	}

	var bothNull bool
	err := tx.QueryRow(ctx,
		"select description is null and avatar_url is null from users where id = $1",
		toPgUUID(u.ID()),
	).Scan(&bothNull)
	if err != nil {
		t.Fatal(err)
	}
	if !bothNull {
		t.Error("空文字が NULL として保存されていない")
	}

	got, err := repo.FindByID(ctx, u.ID())
	if err != nil {
		t.Fatal(err)
	}
	assertSameUser(t, got, u)
}

func TestUserRepository_username重複はErrUsernameAlreadyExists(t *testing.T) {
	ctx := context.Background()
	repo := NewUserRepository(newTestTx(t))

	if _, err := repo.Create(ctx, newUser(t, "carol_03", "", "")); err != nil {
		t.Fatal(err)
	}
	_, err := repo.Create(ctx, newUser(t, "carol_03", "", ""))
	if !errors.Is(err, user.ErrUsernameAlreadyExists) {
		t.Fatalf("err = %v, want %v", err, user.ErrUsernameAlreadyExists)
	}
}

func TestUserRepository_存在しないIDはErrUserNotFound(t *testing.T) {
	repo := NewUserRepository(newTestTx(t))

	_, err := repo.FindByID(context.Background(), uuid.New())
	if !errors.Is(err, user.ErrUserNotFound) {
		t.Fatalf("err = %v, want %v", err, user.ErrUserNotFound)
	}
}
