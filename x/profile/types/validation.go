package types

import (
	"fmt"
	"unicode"
)

// ValidateUserHandle ensures that the userHandle consists solely of lowercase English letters (a-z).
func ValidateUserHandle(userHandle string) (bool, error) {
	if len(userHandle) == 0 {
		return false, fmt.Errorf("userHandle cannot be empty")
	}
	for _, r := range userHandle {
		if !unicode.IsLower(r) || !unicode.IsLetter(r) || r > unicode.MaxASCII || r < 'a' || r > 'z' {
			return false, fmt.Errorf("userHandle must contain only lowercase English letters (a-z)")
		}
	}
	return true, nil
}
