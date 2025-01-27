package types

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrProfileNotFound   = errorsmod.Register(ModuleName, 1100, "profile not found")
	ErrInvalidRequest    = errorsmod.Register(ModuleName, 1101, "invalid request")
	ErrInvalidAddress    = errorsmod.Register(ModuleName, 1102, "invalid address error")
	ErrInvalidUserHandle = errorsmod.Register(ModuleName, 1103, "userHandle unavailable")
	ErrCannotFollowSelf  = errorsmod.Register(ModuleName, 1104, "cannot follow oneself")
)
