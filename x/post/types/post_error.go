package types

import (
	errorsmod "cosmossdk.io/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Core error types
var (
	ErrRequestDenied       = errorsmod.Register(ModuleName, 10000, "request denied")
	ErrPostNotFound        = errorsmod.Register(ModuleName, 1100, "post not found")
	ErrInvalidRequest      = errorsmod.Register(ModuleName, 1101, "invalid request")
	ErrInvalidAddress      = errorsmod.Register(ModuleName, 1102, "invalid address")
	ErrAlreadyLiked        = errorsmod.Register(ModuleName, 1103, "user has already liked this post")
	ErrAlreadySaved        = errorsmod.Register(ModuleName, 1104, "user has already saved this post")
	ErrInvalidLikeCount    = errorsmod.Register(ModuleName, 1105, "invalid like count")
	ErrInvalidPostType     = errorsmod.Register(ModuleName, 1106, "invalid post type")
	ErrLikesIMadeRemove    = errorsmod.Register(ModuleName, 1107, "error removing likes I made")
	ErrSavesIMadeRemove    = errorsmod.Register(ModuleName, 1108, "error removing saves I made")
	ErrLikesReceivedRemove = errorsmod.Register(ModuleName, 1109, "error removing likes received")
	ErrVotingNotStarted    = errorsmod.Register(ModuleName, 1110, "voting has not started yet")
	ErrVotingEnded         = errorsmod.Register(ModuleName, 1111, "voting has ended")
	ErrAlreadyVoted        = errorsmod.Register(ModuleName, 1112, "already voted")
)

// Additional error types for better coverage
var (
	ErrInvalidContent        = errorsmod.Register(ModuleName, 1200, "invalid content")
	ErrContentTooLong        = errorsmod.Register(ModuleName, 1201, "content exceeds maximum length")
	ErrEmptyContent          = errorsmod.Register(ModuleName, 1202, "content cannot be empty")
	ErrTooManyMentions       = errorsmod.Register(ModuleName, 1203, "too many mentions")
	ErrTooManyTopics         = errorsmod.Register(ModuleName, 1204, "too many topics")
	ErrTooManyImages         = errorsmod.Register(ModuleName, 1205, "too many images")
	ErrImageTooLarge         = errorsmod.Register(ModuleName, 1206, "image size exceeds limit")
	ErrTopicNotFound         = errorsmod.Register(ModuleName, 1207, "topic not found")
	ErrCategoryNotFound      = errorsmod.Register(ModuleName, 1208, "category not found")
	ErrUnauthorized          = errorsmod.Register(ModuleName, 1209, "unauthorized operation")
	ErrCommentNotFound       = errorsmod.Register(ModuleName, 1210, "comment not found")
	ErrInvalidPagination     = errorsmod.Register(ModuleName, 1211, "invalid pagination parameters")
	ErrDatabaseOperation     = errorsmod.Register(ModuleName, 1212, "database operation failed")
	ErrProfileNotFound       = errorsmod.Register(ModuleName, 1213, "profile not found")
	ErrQuotePostNotFound     = errorsmod.Register(ModuleName, 1214, "quoted post not found")
	ErrRepostNotFound        = errorsmod.Register(ModuleName, 1215, "repost not found")
	ErrFollowingNotFound     = errorsmod.Register(ModuleName, 1216, "following relationship not found")
	ErrPermissionDenied      = errorsmod.Register(ModuleName, 1217, "permission denied")
	ErrInvalidParameter      = errorsmod.Register(ModuleName, 1218, "invalid parameter")
	ErrOperationTimeout      = errorsmod.Register(ModuleName, 1219, "operation timeout")
	ErrResourceLimitExceeded = errorsmod.Register(ModuleName, 1220, "resource limit exceeded")
	ErrResourceNotFound      = errorsmod.Register(ModuleName, 1221, "resource not found")
)

// Error helper functions

// WrapError wraps an error with additional context
func WrapError(err error, msg string) error {
	if err == nil {
		return nil
	}
	return errorsmod.Wrap(err, msg)
}

// WrapErrorf wraps an error with formatted context
func WrapErrorf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return errorsmod.Wrapf(err, format, args...)
}

// NewInvalidRequestError creates a new invalid request error with context
func NewInvalidRequestError(msg string) error {
	return errorsmod.Wrap(ErrInvalidRequest, msg)
}

// NewInvalidRequestErrorf creates a new invalid request error with formatted context
func NewInvalidRequestErrorf(format string, args ...interface{}) error {
	return errorsmod.Wrapf(ErrInvalidRequest, format, args...)
}

// NewInvalidAddressError creates a new invalid address error with context
func NewInvalidAddressError(msg string) error {
	return errorsmod.Wrap(ErrInvalidAddress, msg)
}

// NewInvalidAddressErrorf creates a new invalid address error with formatted context
func NewInvalidAddressErrorf(format string, args ...interface{}) error {
	return errorsmod.Wrapf(ErrInvalidAddress, format, args...)
}

// NewPostNotFoundError creates a new post not found error with context
func NewPostNotFoundError(postID string) error {
	return errorsmod.Wrapf(ErrPostNotFound, "post with ID %s not found", postID)
}

// NewTopicNotFoundError creates a new topic not found error with context
func NewTopicNotFoundError(topicID string) error {
	return errorsmod.Wrapf(ErrTopicNotFound, "topic with ID %s not found", topicID)
}

// NewCategoryNotFoundError creates a new category not found error with context
func NewCategoryNotFoundError(categoryID string) error {
	return errorsmod.Wrapf(ErrCategoryNotFound, "category with ID %s not found", categoryID)
}

// NewContentTooLongError creates a new content too long error with context
func NewContentTooLongError(actualLength, maxLength int) error {
	return errorsmod.Wrapf(ErrContentTooLong, "content length %d exceeds maximum %d", actualLength, maxLength)
}

// NewTooManyMentionsError creates a new too many mentions error with context
func NewTooManyMentionsError(actualCount, maxCount int) error {
	return errorsmod.Wrapf(ErrTooManyMentions, "mention count %d exceeds maximum %d", actualCount, maxCount)
}

// NewTooManyTopicsError creates a new too many topics error with context
func NewTooManyTopicsError(actualCount, maxCount int) error {
	return errorsmod.Wrapf(ErrTooManyTopics, "topic count %d exceeds maximum %d", actualCount, maxCount)
}

// NewTooManyImagesError creates a new too many images error with context
func NewTooManyImagesError(actualCount, maxCount int) error {
	return errorsmod.Wrapf(ErrTooManyImages, "image count %d exceeds maximum %d", actualCount, maxCount)
}

// NewResourceNotFoundError creates a new resource not found error with context
func NewResourceNotFoundError(msg string) error {
	return errorsmod.Wrap(ErrResourceNotFound, msg)
}

// NewResourceNotFoundErrorf creates a new resource not found error with formatted context
func NewResourceNotFoundErrorf(format string, args ...interface{}) error {
	return errorsmod.Wrapf(ErrResourceNotFound, format, args...)
}

// ToGRPCError converts an error to a gRPC status error
func ToGRPCError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errorsmod.IsOf(err, ErrPostNotFound, ErrTopicNotFound, ErrCategoryNotFound, ErrCommentNotFound, ErrProfileNotFound, ErrResourceNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errorsmod.IsOf(err, ErrInvalidRequest, ErrInvalidAddress, ErrInvalidContent, ErrInvalidParameter, ErrInvalidPagination):
		return status.Error(codes.InvalidArgument, err.Error())
	case errorsmod.IsOf(err, ErrUnauthorized, ErrPermissionDenied):
		return status.Error(codes.PermissionDenied, err.Error())
	case errorsmod.IsOf(err, ErrAlreadyLiked, ErrAlreadySaved, ErrAlreadyVoted):
		return status.Error(codes.AlreadyExists, err.Error())
	case errorsmod.IsOf(err, ErrResourceLimitExceeded, ErrContentTooLong, ErrTooManyMentions, ErrTooManyTopics, ErrTooManyImages):
		return status.Error(codes.ResourceExhausted, err.Error())
	case errorsmod.IsOf(err, ErrOperationTimeout):
		return status.Error(codes.DeadlineExceeded, err.Error())
	case errorsmod.IsOf(err, ErrDatabaseOperation):
		return status.Error(codes.Internal, err.Error())
	default:
		return status.Error(codes.Unknown, err.Error())
	}
}

// IsValidationError checks if an error is a validation error
func IsValidationError(err error) bool {
	return errorsmod.IsOf(err,
		ErrInvalidRequest,
		ErrInvalidAddress,
		ErrInvalidContent,
		ErrContentTooLong,
		ErrEmptyContent,
		ErrTooManyMentions,
		ErrTooManyTopics,
		ErrTooManyImages,
		ErrImageTooLarge,
		ErrInvalidParameter,
		ErrInvalidPagination,
	)
}

// IsNotFoundError checks if an error is a not found error
func IsNotFoundError(err error) bool {
	return errorsmod.IsOf(err,
		ErrPostNotFound,
		ErrTopicNotFound,
		ErrCategoryNotFound,
		ErrCommentNotFound,
		ErrProfileNotFound,
		ErrQuotePostNotFound,
		ErrRepostNotFound,
		ErrFollowingNotFound,
		ErrResourceNotFound,
	)
}

// IsPermissionError checks if an error is a permission error
func IsPermissionError(err error) bool {
	return errorsmod.IsOf(err,
		ErrUnauthorized,
		ErrPermissionDenied,
		ErrRequestDenied,
	)
}

// LogError logs an error with context
func LogError(logger interface{ Error(string, ...interface{}) }, operation string, err error, context ...interface{}) {
	if err == nil {
		return
	}

	args := []interface{}{"operation", operation, "error", err}
	args = append(args, context...)
	logger.Error("Post module error", args...)
}
