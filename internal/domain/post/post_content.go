package post

import (
	"errors"
	"net/url"
	"strings"
	"unicode/utf8"
)

const MaxBodyLength = 1000

var (
	ErrEmptyPostContent = errors.New("post content must have body or image_url")
	ErrPostBodyTooLong  = errors.New("post body exceeds max length")
	ErrInvalidImageURL  = errors.New("image_url must be an http(s) URL")
)

type PostContent struct {
	body     string
	imageURL string
}

// NewPostContent は入力を検証して PostContent を作る。
// 空白のみは「無し」とみなして弾くが、保存する値はトリムせず原文のまま。
func NewPostContent(body, imageURL string) (PostContent, error) {
	hasBody := strings.TrimSpace(body) != ""
	hasImage := strings.TrimSpace(imageURL) != ""

	if !hasBody && !hasImage {
		return PostContent{}, ErrEmptyPostContent
	}
	if utf8.RuneCountInString(body) > MaxBodyLength {
		return PostContent{}, ErrPostBodyTooLong
	}
	if hasImage && !isHTTPURL(imageURL) {
		return PostContent{}, ErrInvalidImageURL
	}

	return PostContent{body: body, imageURL: imageURL}, nil
}

// RebuildPostContent は永続化済みの値から PostContent を組み立て直す。
// 保存済みデータは保存時点で検証済みとみなし、ルールは再実行しない。
func RebuildPostContent(body, imageURL string) PostContent {
	return PostContent{body: body, imageURL: imageURL}
}

func (c PostContent) Body() string {
	return c.body
}

func (c PostContent) ImageURL() string {
	return c.imageURL
}

func isHTTPURL(s string) bool {
	u, err := url.Parse(s)
	if err != nil {
		return false
	}
	return (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}
