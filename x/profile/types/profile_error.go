package types

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrRequestDenied         = errorsmod.Register(ModuleName, 10000, "Request denied.")
	ErrProfileNotFound       = errorsmod.Register(ModuleName, 1100, "profile not found")
	ErrInvalidRequest        = errorsmod.Register(ModuleName, 1101, "invalid request")
	ErrInvalidAddress        = errorsmod.Register(ModuleName, 1102, "invalid address error")
	ErrInvalidUserHandle     = errorsmod.Register(ModuleName, 1103, "userHandle unavailable")
	ErrCannotFollowSelf      = errorsmod.Register(ModuleName, 1104, "cannot follow oneself")
	ErrInvalidAdminAddress   = errorsmod.Register(ModuleName, 1105, "invalid admin address")
	ErrInvalidChiefModerator = errorsmod.Register(ModuleName, 1106, "invalid chief moderator")
)
