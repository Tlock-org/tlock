package keeper

import (
	"context"
	sdk "github.com/cosmos/cosmos-sdk/types"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"cosmossdk.io/errors"
	"github.com/rollchains/tlock/x/post/types"
)

type msgServer struct {
	k Keeper
}

var _ types.MsgServer = msgServer{}

// NewMsgServerImpl returns an implementation of the module MsgServer interface.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{k: keeper}
}

func (ms msgServer) UpdateParams(ctx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if ms.k.authority != msg.Authority {
		return nil, errors.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", ms.k.authority, msg.Authority)
	}

	return nil, ms.k.Params.Set(ctx, msg.Params)
}

// SetServiceName implements types.MsgServer.
func (ms msgServer) SetServiceName(ctx context.Context, msg *types.MsgSetServiceName) (*types.MsgSetServiceNameResponse, error) {
	// ctx := sdk.UnwrapSDKContext(goCtx)
	//panic("SetServiceName is unimplemented")
	//return &types.MsgSetServiceNameResponse{}, nil
	if err := ms.k.NameMapping.Set(ctx, msg.Sender, msg.Name); err != nil {
		return nil, err
	}

	return &types.MsgSetServiceNameResponse{}, nil
}

// CreatePost implements types.MsgServer.
func (ms msgServer) CreatePost(goCtx context.Context, msg *types.MsgCreatePost) (*types.MsgCreatePostResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate the message
	if len(msg.Title) == 0 {
		return nil, errors.Wrapf(types.ErrInvalidRequest, "Title cannot be empty")
	}
	if len(msg.Content) == 0 {
		return nil, errors.Wrapf(types.ErrInvalidRequest, "Content cannot be empty")
	}

	// Validate sender address
	//sender, err := sdk.AccAddressFromBech32(msg.Sender)
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, errors.Wrapf(types.ErrInvalidAddress, "Invalid sender address: %s", err)
	}

	// Generate a unique post ID
	//postID := ms.k.generatePostID(ctx) // 需要在 Keeper 中实现 generatePostID 方法
	postID := msg.PostId // 需要在 Keeper 中实现 generatePostID 方法

	// Create the post
	post := types.Post{
		Id:        postID,
		Title:     msg.Title,
		Content:   msg.Content,
		Sender:    msg.Sender,
		Timestamp: msg.Timestamp,
	}

	// Store the post in the state
	ms.k.SetPost(ctx, post)
	// post reward
	ms.k.PostReward(ctx, post)

	// Emit an event for the creation
	//ctx.EventManager().EmitEvent(
	//	sdk.NewEvent(
	//		types.EventTypeCreatePost,
	//		sdk.NewAttribute(types.AttributeKeyPostID, postID),
	//		sdk.NewAttribute(types.AttributeKeySender, msg.Sender),
	//	),
	//)

	return &types.MsgCreatePostResponse{PostId: postID}, nil
}

// SetApprove implements types.MsgServer.
func (ms msgServer) SetApprove(goCtx context.Context, msg *types.MsgSetApprove) (*types.MsgSetApproveResponse, error) {
	// ctx := sdk.UnwrapSDKContext(goCtx)
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return &types.MsgSetApproveResponse{Status: false}, errors.Wrapf(types.ErrInvalidAddress, "Invalid sender address: %s", err)
	}
	ms.k.ApproveFeegrant(ctx, sender)

	return &types.MsgSetApproveResponse{Status: true}, nil
}
