package types

import (
	"regexp"
	"strings"
)

// ValidateUserHandle validates user handle format and constraints
func ValidateUserHandle(userHandle string) (bool, error) {
	trimmedHandle := strings.TrimSpace(userHandle)
	if trimmedHandle == "" {
		return false, NewInvalidUserHandleError("user handle cannot be empty or consist solely of whitespace")
	}
	if len(trimmedHandle) > 36 {
		return false, NewInvalidUserHandleError("user handle must be 36 characters or less")
	}
	var validHandle = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	if !validHandle.MatchString(trimmedHandle) {
		return false, NewInvalidUserHandleError("user handle must contain only lowercase English letters (a-z), digits (0-9), or underscores (_)")
	}
	return true, nil
}

// ValidateNickname validates nickname format and constraints
func ValidateNickname(nickname string) (bool, error) {
	// Trim leading and trailing spaces to accurately check for empty or all-space nicknames
	trimmedNick := strings.TrimSpace(nickname)
	if trimmedNick == "" {
		return false, NewInvalidNicknameError("nickname cannot be empty or consist only of spaces")
	}
	if len(trimmedNick) > 15 {
		return false, NewInvalidNicknameError("nickname must be 15 characters or less")
	}

	//validNickname := regexp.MustCompile(`^[\p{L}\p{N}_ ]+$`)
	validNickname := regexp.MustCompile(`^[^@#$]+$`)
	if !validNickname.MatchString(trimmedNick) {
		//return false, NewInvalidNicknameError("nickName must contain only letters, numbers, underscores (_), or spaces")
		return false, NewInvalidNicknameError("nickname cannot contain '@', '#', or '$' characters")

	}
	if strings.Contains(trimmedNick, "  ") || strings.Contains(trimmedNick, "__") {
		return false, NewInvalidNicknameError("nickname cannot contain consecutive spaces or underscores")
	}
	return true, nil
}
