// types/errors.go

package types

import (
	//"github.com/cosmos/cosmos-sdk/types/errors"
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrPostNotFound   = errorsmod.Register(ModuleName, 1100, "post not found")
	ErrInvalidRequest = errorsmod.Register(ModuleName, 1101, "invalid request")
	ErrInvalidAddress = errorsmod.Register(ModuleName, 1102, "invalid address error")
	// 其他错误...
)
