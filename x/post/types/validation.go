package types

import (
	"strings"
	"unicode/utf8"
)

// ValidatePostContent validates that the content does not exceed the maximum allowed length
func ValidatePostContent(content string) error {
	// Check if content is empty
	if strings.TrimSpace(content) == "" {
		return ErrEmptyContent
	}

	length := utf8.RuneCountInString(content)
	if length > MaxPostContentLength {
		return NewContentTooLongError(length, MaxPostContentLength)
	}
	return nil
}

// ValidatePostWithTitleContent validates that the content with title does not exceed the maximum allowed length
func ValidatePostWithTitleContent(content string) error {
	// Check if content is empty
	if strings.TrimSpace(content) == "" {
		return ErrEmptyContent
	}

	length := utf8.RuneCountInString(content)
	if length > MaxPostWithTitleContentLength {
		return NewContentTooLongError(length, MaxPostWithTitleContentLength)
	}
	return nil
}

// ValidateTitle validates post title
func ValidateTitle(title string) error {
	if strings.TrimSpace(title) == "" {
		return NewInvalidRequestError("title cannot be empty")
	}

	length := utf8.RuneCountInString(title)
	if length > MaxPostTitleLength {
		return NewContentTooLongError(length, MaxPostTitleLength)
	}
	return nil
}

// ValidateMentions validates the number of mentions
func ValidateMentions(mentions []string) error {
	if len(mentions) > MaxMentionUsers {
		return NewTooManyMentionsError(len(mentions), MaxMentionUsers)
	}
	return nil
}

// ValidateTopics validates the number of topics
func ValidateTopics(topics []string) error {
	if len(topics) > MaxTopicsPerPost {
		return NewTooManyTopicsError(len(topics), MaxTopicsPerPost)
	}
	return nil
}

// ValidateImages validates the number of images
func ValidateImages(images []string) error {
	if len(images) > MaxImagesPerPost {
		return NewTooManyImagesError(len(images), MaxImagesPerPost)
	}
	return nil
}

// ValidatePostID validates post ID format
func ValidatePostID(postID string) error {
	if strings.TrimSpace(postID) == "" {
		return NewInvalidRequestError("post ID cannot be empty")
	}

	// Check minimum length (SHA256 hash should be 64 characters)
	if len(postID) < 32 {
		return NewInvalidRequestError("post ID too short")
	}

	// Check maximum length to prevent excessively long IDs
	if len(postID) > 128 {
		return NewInvalidRequestError("post ID too long")
	}

	// Check for valid characters (alphanumeric)
	for _, char := range postID {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9')) {
			return NewInvalidRequestError("post ID contains invalid characters")
		}
	}

	return nil
}

// ValidateUserAddress validates user address format
func ValidateUserAddress(address string) error {
	if strings.TrimSpace(address) == "" {
		return NewInvalidAddressError("address cannot be empty")
	}

	// Check minimum length for bech32 addresses
	if len(address) < 10 {
		return NewInvalidAddressError("address too short")
	}

	// Check maximum length to prevent excessively long addresses
	if len(address) > 100 {
		return NewInvalidAddressError("address too long")
	}

	// Basic bech32 format check (should start with letter and contain only valid characters)
	if !((address[0] >= 'a' && address[0] <= 'z') || (address[0] >= 'A' && address[0] <= 'Z')) {
		return NewInvalidAddressError("address should start with a letter")
	}

	// Check for valid bech32 characters
	for _, char := range address {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9')) {
			return NewInvalidAddressError("address contains invalid characters")
		}
	}

	return nil
}

// ValidatePageRequest validates pagination parameters
func ValidatePageRequest(page, limit uint64) error {
	if page == 0 {
		return NewInvalidRequestError("page must be greater than 0")
	}
	if limit == 0 {
		return NewInvalidRequestError("limit must be greater than 0")
	}
	if limit > MaxPageLimit {
		return NewInvalidRequestErrorf("limit %d exceeds maximum %d", limit, MaxPageLimit)
	}
	return nil
}

// ValidateCategory validates category format
func ValidateCategory(category string) error {
	if strings.TrimSpace(category) == "" {
		return NewInvalidRequestError("category cannot be empty")
	}
	return nil
}

// ValidateTopic validates topic format
func ValidateTopic(topic string) error {
	if strings.TrimSpace(topic) == "" {
		return NewInvalidRequestError("topic cannot be empty")
	}
	return nil
}
