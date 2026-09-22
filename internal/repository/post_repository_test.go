package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"sns/internal/domain/post"
)

func newPost(t *testing.T, userID uuid.UUID, body, imageURL string) *post.Post {
	t.Helper()

	content, err := post.NewPostContent(body, imageURL)
	if err != nil {
		t.Fatal(err)
	}
	p, err := post.NewPost(userID, content, true, fixedNow)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func assertSamePost(t *testing.T, got, want *post.Post) {
	t.Helper()

	if got.ID() != want.ID() {
		t.Errorf("ID() = %v, want %v", got.ID(), want.ID())
	}
	if got.UserID() != want.UserID() {
		t.Errorf("UserID() = %v, want %v", got.UserID(), want.UserID())
	}
	if got.Content() != want.Content() {
		t.Errorf("Content() = %+v, want %+v", got.Content(), want.Content())
	}
	if got.IsPublic() != want.IsPublic() {
		t.Errorf("IsPublic() = %v, want %v", got.IsPublic(), want.IsPublic())
	}
	// timestamptz はマイクロ秒精度なので、Go 側のナノ秒を切り捨ててから比較する。
	if !got.CreatedAt().Equal(want.CreatedAt().Truncate(time.Microsecond)) {
		t.Errorf("CreatedAt() = %v, want %v", got.CreatedAt(), want.CreatedAt())
	}
}

func TestPostRepository_CreateAndFindByID_RoundTrips(t *testing.T) {
	ctx := context.Background()
	tx := newTestTx(t)
	repo := NewPostRepository(tx)
	p := newPost(t, newTestUser(t, tx).ID(), "hello", "")

	created, err := repo.Create(ctx, p)
	if err != nil {
		t.Fatal(err)
	}
	assertSamePost(t, created, p)

	got, err := repo.FindByID(ctx, p.ID())
	if err != nil {
		t.Fatal(err)
	}
	assertSamePost(t, got, p)
}

func TestPostRepository_ImageOnlyPost_StoresBodyAsNull(t *testing.T) {
	ctx := context.Background()
	tx := newTestTx(t)
	repo := NewPostRepository(tx)
	p := newPost(t, newTestUser(t, tx).ID(), "", "https://example.com/a.png")

	if _, err := repo.Create(ctx, p); err != nil {
		t.Fatal(err)
	}

	var bodyIsNull bool
	err := tx.QueryRow(ctx, "select body is null from posts where id = $1", toPgUUID(p.ID())).Scan(&bodyIsNull)
	if err != nil {
		t.Fatal(err)
	}
	if !bodyIsNull {
		t.Error("空文字の body が NULL として保存されていない")
	}

	got, err := repo.FindByID(ctx, p.ID())
	if err != nil {
		t.Fatal(err)
	}
	assertSamePost(t, got, p)
}

func TestPostRepository_Create_DuplicateReturnsErrPostAlreadyExists(t *testing.T) {
	ctx := context.Background()
	tx := newTestTx(t)
	repo := NewPostRepository(tx)
	p := newPost(t, newTestUser(t, tx).ID(), "hello", "")

	if _, err := repo.Create(ctx, p); err != nil {
		t.Fatal(err)
	}
	_, err := repo.Create(ctx, p)
	if !errors.Is(err, post.ErrPostAlreadyExists) {
		t.Fatalf("err = %v, want %v", err, post.ErrPostAlreadyExists)
	}
}

func TestPostRepository_Create_UnknownUserReturnsErrUserNotFound(t *testing.T) {
	tx := newTestTx(t)
	repo := NewPostRepository(tx)
	p := newPost(t, uuid.New(), "hello", "")

	_, err := repo.Create(context.Background(), p)
	if !errors.Is(err, post.ErrUserNotFound) {
		t.Fatalf("err = %v, want %v", err, post.ErrUserNotFound)
	}
}

func TestPostRepository_FindByID_UnknownIDReturnsErrPostNotFound(t *testing.T) {
	tx := newTestTx(t)
	repo := NewPostRepository(tx)

	_, err := repo.FindByID(context.Background(), uuid.New())
	if !errors.Is(err, post.ErrPostNotFound) {
		t.Fatalf("err = %v, want %v", err, post.ErrPostNotFound)
	}
}
