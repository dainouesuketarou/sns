package user

import (
	"errors"
	"unicode/utf8"
)

const MaxDescriptionLength = 500

var ErrDescriptionTooLong = errors.New("description exceeds max length")

// Description は自己紹介文の値オブジェクト。空（未設定）を許す。
type Description struct {
	value string
}

func NewDescription(s string) (Description, error) {
	if utf8.RuneCountInString(s) > MaxDescriptionLength {
		return Description{}, ErrDescriptionTooLong
	}
	return Description{value: s}, nil
}

// RebuildDescription は永続化済みの値から組み立て直す。保存時に検証済みとみなしルールは再実行しない。
func RebuildDescription(s string) Description {
	return Description{value: s}
}

func (d Description) String() string {
	return d.value
}

// IsSet は自己紹介文が設定されているかを返す。
func (d Description) IsSet() bool {
	return d.value != ""
}
