package user

import (
	"errors"
	"regexp"
	"strings"
	"unicode/utf8"
)

const MaxUsernameLength = 20

var (
	ErrEmptyUsername        = errors.New("username must not be empty")
	ErrUsernameTooLong      = errors.New("username exceeds max length")
	ErrInvalidUsernameChars = errors.New("username must contain only [a-zA-Z0-9_]")
)

var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

// UserName はユーザー名の値オブジェクト。
// 一意性はここでは検証しない（他の行を見ないと判断できないため DB の UNIQUE に任せる）。
type UserName struct {
	value string
}

func NewUserName(s string) (UserName, error) {
	if strings.TrimSpace(s) == "" {
		return UserName{}, ErrEmptyUsername
	}
	if utf8.RuneCountInString(s) > MaxUsernameLength {
		return UserName{}, ErrUsernameTooLong
	}
	if !usernamePattern.MatchString(s) {
		return UserName{}, ErrInvalidUsernameChars
	}
	return UserName{value: s}, nil
}

// RebuildUserName は永続化済みの値から組み立て直す。保存時に検証済みとみなしルールは再実行しない。
func RebuildUserName(s string) UserName {
	return UserName{value: s}
}

func (n UserName) String() string {
	return n.value
}
