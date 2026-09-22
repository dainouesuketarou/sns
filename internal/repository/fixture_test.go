package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"sns/internal/domain/user"
)

var fixedNow = time.Date(2026, 9, 23, 10, 0, 0, 0, time.UTC)

// newTestTx は実DBへのトランザクションを開き、テスト終了時にロールバックする。
// DATABASE_URL が未設定ならテストをスキップする。
func newTestTx(t *testing.T) pgx.Tx {
	t.Helper()

	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL が未設定のためスキップ")
	}

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = tx.Rollback(ctx) })

	return tx
}

// newTestUser は外部キーを満たすためのユーザーを1件作って返す。
// username は毎回ユニークにする（同一トランザクション内で複数作っても衝突しないため）。
func newTestUser(t *testing.T, tx pgx.Tx) *user.User {
	t.Helper()

	name, err := user.NewUserName("u_" + uuid.New().String()[:8])
	if err != nil {
		t.Fatal(err)
	}
	u := user.NewUser(name, user.Description{}, user.AvatarURL{}, fixedNow)

	created, err := NewUserRepository(tx).Create(context.Background(), u)
	if err != nil {
		t.Fatal(err)
	}
	return created
}
