package user

import (
	"errors"
	"net/url"
	"unicode/utf8"
)

const MaxAvatarURLLength = 2000

var (
	ErrInvalidAvatarURL = errors.New("avatar_url must be an http(s) URL")
	ErrAvatarURLTooLong = errors.New("avatar_url exceeds max length")
)

// AvatarURL はアイコン画像の URL の値オブジェクト。
// 未設定を許すため、ゼロ値（空）を「設定なし」として扱う。
type AvatarURL struct {
	value string
}

// NewAvatarURL は URL を検証して AvatarURL を作る。空文字は「設定なし」として許可する。
// http(s) 以外を弾くのは、javascript: のようなスキームがそのまま img/a タグに入るのを防ぐため。
func NewAvatarURL(s string) (AvatarURL, error) {
	if s == "" {
		return AvatarURL{}, nil
	}
	if utf8.RuneCountInString(s) > MaxAvatarURLLength {
		return AvatarURL{}, ErrAvatarURLTooLong
	}

	u, err := url.Parse(s)
	if err != nil {
		return AvatarURL{}, ErrInvalidAvatarURL
	}
	if (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return AvatarURL{}, ErrInvalidAvatarURL
	}

	return AvatarURL{value: s}, nil
}

// RebuildAvatarURL は永続化済みの値から組み立て直す。保存時に検証済みとみなしルールは再実行しない。
func RebuildAvatarURL(s string) AvatarURL {
	return AvatarURL{value: s}
}

func (a AvatarURL) String() string {
	return a.value
}

// IsSet はアイコンが設定されているかを返す。表示側で「デフォルト画像にするか」の判断に使う。
func (a AvatarURL) IsSet() bool {
	return a.value != ""
}
