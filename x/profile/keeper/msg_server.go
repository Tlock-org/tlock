package keeper

import (
	"context"
	sdk "github.com/cosmos/cosmos-sdk/types"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"cosmossdk.io/errors"
	"github.com/rollchains/tlock/x/profile/types"
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

// AddProfile implements types.MsgServer.
func (ms msgServer) AddProfile(goCtx context.Context, msg *types.MsgAddProfileRequest) (*types.MsgAddProfileResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate the message
	//if len(msg.WalletAddress) == 0 {
	//	return nil, errors.Wrapf(types.ErrInvalidRequest, "Content cannot be empty")
	//}

	// Validate sender address
	//sender, err := sdk.AccAddressFromBech32(msg.Sender)
	//_, err := sdk.AccAddressFromBech32(msg.Creator)
	//if err != nil {
	//	return nil, errors.Wrapf(types.ErrInvalidAddress, "Invalid sender address: %s", err)
	//}

	//if len(msg.Image) > MaxImageSize {
	//	return nil, errors.Wrap(types.ErrInvalidRequest, "Image size exceeds the maximum allowed limit")
	//}

	blockTime := ctx.BlockTime().Unix()
	//data := fmt.Sprintf("%s|%s|%d", msg.Creator, msg.Nickname, blockTime)
	//postID := ms.k.generatePostID(data)

	// Create the post
	profile := types.Profile{
		WalletAddress: msg.Creator,
		Nickname:      msg.Nickname,
		CreationTime:  blockTime,
	}

	// post payment
	ms.k.SetProfile(ctx, profile)

	// Store the post in the state
	//ms.k.SetPost(ctx, post)

	//Emit an event for the creation
	//ctx.EventManager().EmitEvents(sdk.Events{
	//	sdk.NewEvent(
	//		types.EventTypeCreatePaidPost,
	//		sdk.NewAttribute(types.AttributeKeyCreator, msg.Creator),
	//		sdk.NewAttribute(types.AttributeKeyPostID, postID),
	//		sdk.NewAttribute(types.AttributeKeyTitle, msg.Title),
	//		sdk.NewAttribute(types.AttributeKeyTimestamp, fmt.Sprintf("%d", blockTime)),
	//	),
	//})

	return &types.MsgAddProfileResponse{}, nil
}
