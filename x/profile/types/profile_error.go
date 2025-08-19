package types

import (
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Error codes for profile module
var (
	ErrRequestDenied         = errorsmod.Register(ModuleName, 10000, "request denied")
	ErrProfileNotFound       = errorsmod.Register(ModuleName, 1100, "profile not found")
	ErrInvalidRequest        = errorsmod.Register(ModuleName, 1101, "invalid request")
	ErrInvalidAddress        = errorsmod.Register(ModuleName, 1102, "invalid address")
	ErrInvalidUserHandle     = errorsmod.Register(ModuleName, 1103, "user handle unavailable")
	ErrCannotFollowSelf      = errorsmod.Register(ModuleName, 1104, "cannot follow oneself")
	ErrInvalidAdminAddress   = errorsmod.Register(ModuleName, 1105, "invalid admin address")
	ErrInvalidChiefModerator = errorsmod.Register(ModuleName, 1106, "invalid chief moderator")
	ErrDatabaseOperation     = errorsmod.Register(ModuleName, 1107, "database operation failed")
	ErrInvalidParameter      = errorsmod.Register(ModuleName, 1108, "invalid parameter")
	ErrResourceNotFound      = errorsmod.Register(ModuleName, 1109, "resource not found")
	ErrInvalidNickname       = errorsmod.Register(ModuleName, 1110, "invalid nickname")
	ErrValidationFailed      = errorsmod.Register(ModuleName, 1111, "validation failed")
)

// Error helper functions

// WrapError wraps an error with additional context
func WrapError(err error, msg string) error {
	return errorsmod.Wrap(err, msg)
}

// WrapErrorf wraps an error with formatted message
func WrapErrorf(err error, format string, args ...interface{}) error {
	return errorsmod.Wrapf(err, format, args...)
}

// NewInvalidRequestError creates a new invalid request error
func NewInvalidRequestError(msg string) error {
	return errorsmod.Wrap(ErrInvalidRequest, msg)
}

// NewInvalidRequestErrorf creates a new invalid request error with formatting
func NewInvalidRequestErrorf(format string, args ...interface{}) error {
	return errorsmod.Wrapf(ErrInvalidRequest, format, args...)
}

// NewProfileNotFoundError creates a new profile not found error
func NewProfileNotFoundError(address string) error {
	return errorsmod.Wrapf(ErrProfileNotFound, "profile with address %s not found", address)
}

// NewInvalidUserHandleError creates a new invalid user handle error
func NewInvalidUserHandleError(msg string) error {
	return errorsmod.Wrap(ErrInvalidUserHandle, msg)
}

// NewInvalidNicknameError creates a new invalid nickname error
func NewInvalidNicknameError(msg string) error {
	return errorsmod.Wrap(ErrInvalidNickname, msg)
}

// NewValidationError creates a new validation error
func NewValidationError(msg string) error {
	return errorsmod.Wrap(ErrValidationFailed, msg)
}

// NewResourceNotFoundErrorf creates a new resource not found error with formatting
func NewResourceNotFoundErrorf(format string, args ...interface{}) error {
	return errorsmod.Wrapf(ErrResourceNotFound, format, args...)
}

// ToGRPCError converts custom errors to gRPC errors
func ToGRPCError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errorsmod.IsOf(err, ErrProfileNotFound, ErrResourceNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errorsmod.IsOf(err, ErrInvalidRequest, ErrInvalidAddress, ErrInvalidUserHandle,
		ErrInvalidNickname, ErrInvalidParameter, ErrValidationFailed):
		return status.Error(codes.InvalidArgument, err.Error())
	case errorsmod.IsOf(err, ErrRequestDenied, ErrInvalidAdminAddress, ErrInvalidChiefModerator):
		return status.Error(codes.PermissionDenied, err.Error())
	case errorsmod.IsOf(err, ErrCannotFollowSelf):
		return status.Error(codes.InvalidArgument, err.Error())
	case errorsmod.IsOf(err, ErrDatabaseOperation):
		return status.Error(codes.Internal, err.Error())
	default:
		return status.Error(codes.Unknown, err.Error())
	}
}

// Error classification functions

// IsValidationError checks if error is a validation error
func IsValidationError(err error) bool {
	return errorsmod.IsOf(err,
		ErrInvalidRequest,
		ErrInvalidAddress,
		ErrInvalidUserHandle,
		ErrInvalidNickname,
		ErrInvalidParameter,
		ErrValidationFailed,
	)
}

// IsNotFoundError checks if error is a not found error
func IsNotFoundError(err error) bool {
	return errorsmod.IsOf(err,
		ErrProfileNotFound,
		ErrResourceNotFound,
	)
}

// IsPermissionError checks if error is a permission error
func IsPermissionError(err error) bool {
	return errorsmod.IsOf(err,
		ErrRequestDenied,
		ErrInvalidAdminAddress,
		ErrInvalidChiefModerator,
	)
}

// LogError logs an error with context information
func LogError(logger log.Logger, operation string, err error, keyvals ...interface{}) {
	args := []interface{}{"operation", operation, "error", err}
	args = append(args, keyvals...)
	logger.Error("Profile module error", args...)
}
