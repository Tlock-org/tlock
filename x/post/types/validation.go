package types

import (
	"fmt"
	"unicode/utf8"
)

// ValidatePostContent validates that the content does not exceed the maximum allowed length
func ValidatePostContent(content string) error {
	length := utf8.RuneCountInString(content)
	if length > MaxPostContentLength {
		return fmt.Errorf("post content exceeds maximum length of %d characters", MaxPostContentLength)
	}
	return nil
}

func ValidatePostWithTitleContent(content string) error {
	length := utf8.RuneCountInString(content)
	if length > MaxPostContentLength {
		return fmt.Errorf("post content exceeds maximum length of %d characters", MaxPostWithTitleContentLength)
	}
	return nil
}
