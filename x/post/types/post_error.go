package types

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrRequestDenied       = errorsmod.Register(ModuleName, 10000, "Request denied.")
	ErrPostNotFound        = errorsmod.Register(ModuleName, 1100, "post not found")
	ErrInvalidRequest      = errorsmod.Register(ModuleName, 1101, "invalid request")
	ErrInvalidAddress      = errorsmod.Register(ModuleName, 1102, "invalid address error")
	ErrAlreadyLiked        = errorsmod.Register(ModuleName, 1103, "user has already liked this post")
	ErrAlreadySaved        = errorsmod.Register(ModuleName, 1104, "user has already saved this post")
	ErrInvalidLikeCount    = errorsmod.Register(ModuleName, 1105, "invalid like count")
	ErrInvalidPostType     = errorsmod.Register(ModuleName, 1106, "invalid post type")
	ErrLikesIMadeRemove    = errorsmod.Register(ModuleName, 1107, "likes i made remove error")
	ErrSavesIMadeRemove    = errorsmod.Register(ModuleName, 1108, "saves i made remove error")
	ErrLikesReceivedRemove = errorsmod.Register(ModuleName, 1109, "likes received remove error")
	ErrVotingNotStarted    = errorsmod.Register(ModuleName, 1110, "Voting has not started yet.")
	ErrVotingEnded         = errorsmod.Register(ModuleName, 1111, "Voting has ended.")
	ErrAlreadyVoted        = errorsmod.Register(ModuleName, 1112, "Already voted.")
)
