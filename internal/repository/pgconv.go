package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// ドメインの型と pgtype の相互変換。pgtype をリポジトリの外に漏らさないための置き場。
// toPg* は Go → DB(書き込み)、fromPg* は DB → Go(読み込み)。NULL の扱いはここで決める。

func toPgUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

// NULL は uuid.Nil になる。nullable な uuid 列を扱うときは (uuid.UUID, bool) を返す関数を別途足す。
func fromPgUUID(v pgtype.UUID) uuid.UUID {
	if !v.Valid {
		return uuid.Nil
	}
	return uuid.UUID(v.Bytes)
}

// 空文字は「値なし」として NULL に落とす。
func toPgText(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: s != ""}
}

// NULL は空文字になる(toPgText の逆)。
func fromPgText(v pgtype.Text) string {
	if !v.Valid {
		return ""
	}
	return v.String
}

func toPgTimestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

func fromPgTimestamptz(v pgtype.Timestamptz) time.Time {
	return v.Time
}
