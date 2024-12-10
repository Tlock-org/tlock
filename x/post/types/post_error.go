package types

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrPostNotFound        = errorsmod.Register(ModuleName, 1100, "post not found")
	ErrInvalidRequest      = errorsmod.Register(ModuleName, 1101, "invalid request")
	ErrInvalidAddress      = errorsmod.Register(ModuleName, 1102, "invalid address error")
	ErrInvalidLikeCount    = errorsmod.Register(ModuleName, 1103, "invalid like count")
	ErrInvalidPostType     = errorsmod.Register(ModuleName, 1104, "invalid post type")
	ErrLikesIMadeRemove    = errorsmod.Register(ModuleName, 1105, "likes i made remove error")
	ErrSavesIMadeRemove    = errorsmod.Register(ModuleName, 1106, "saves i made remove error")
	ErrLikesReceivedRemove = errorsmod.Register(ModuleName, 1107, "likes received remove error")
)
