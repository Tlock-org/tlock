package types

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrPostNotFound     = errorsmod.Register(ModuleName, 1100, "post not found")
	ErrInvalidPostType  = errorsmod.Register(ModuleName, 1104, "invalid post type")
	ErrInvalidRequest   = errorsmod.Register(ModuleName, 1101, "invalid request")
	ErrInvalidAddress   = errorsmod.Register(ModuleName, 1102, "invalid address error")
	ErrInvalidLikeCount = errorsmod.Register(ModuleName, 1103, "invalid like count")
)
