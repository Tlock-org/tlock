package types

import (
	"fmt"
	"regexp"
	"strings"
)

func ValidateUserHandle(userHandle string) (bool, error) {
	trimmedHandle := strings.TrimSpace(userHandle)
	if trimmedHandle == "" {
		return false, fmt.Errorf("userHandle cannot be empty or consist solely of whitespace")
	}
	if len(trimmedHandle) > 36 {
		return false, fmt.Errorf("userHandle must be 36 characters or less")
	}
	var validHandle = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	if !validHandle.MatchString(trimmedHandle) {
		return false, fmt.Errorf("userHandle must contain only lowercase English letters (a-z), digits (0-9), or underscores (_)")
	}
	return true, nil
}

func ValidateNickName(nickName string) (bool, error) {
	// Trim leading and trailing spaces to accurately check for empty or all-space nicknames
	trimmedNick := strings.TrimSpace(nickName)
	if trimmedNick == "" {
		return false, fmt.Errorf("nickname cannot be empty or consist only of spaces")
	}
	if len(trimmedNick) > 36 {
		return false, fmt.Errorf("nickname must be 36 characters or less")
	}

	//validNickname := regexp.MustCompile(`^[\p{L}\p{N}_ ]+$`)
	validNickname := regexp.MustCompile(`^[^@#$]+$`)
	if !validNickname.MatchString(trimmedNick) {
		//return false, fmt.Errorf("nickName must contain only letters, numbers, underscores (_), or spaces")
		return false, fmt.Errorf("nickname cannot contain '@', '#', or '$' characters")

	}
	if strings.Contains(trimmedNick, "  ") || strings.Contains(trimmedNick, "__") {
		return false, fmt.Errorf("nickname cannot contain consecutive spaces or underscores")
	}
	return true, nil
}
