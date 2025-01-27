package types

import (
	"fmt"
	"regexp"
)

// ValidateUserHandle ensures that the userHandle consists solely of lowercase English letters (a-z).
//func ValidateUserHandle(userHandle string) (bool, error) {
//	if len(userHandle) == 0 {
//		return false, fmt.Errorf("userHandle cannot be empty")
//	}
//	for _, r := range userHandle {
//		if !(unicode.IsLower(r) || unicode.IsDigit(r)) || r > unicode.MaxASCII {
//			return false, fmt.Errorf("userHandle must contain only lowercase English letters (a-z) or digits (0-9)")
//		}
//	}
//	return true, nil
//}

func ValidateUserHandle(userHandle string) (bool, error) {
	var validHandle = regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	if !validHandle.MatchString(userHandle) {
		return false, fmt.Errorf("userHandle must contain only lowercase English letters (a-z) or digits (0-9)")
	}
	return true, nil
}
