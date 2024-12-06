package keeper

import (
	"context"
	"fmt"
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

	// Validate creator address
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errors.Wrapf(types.ErrInvalidAddress, "Invalid creator address: %s", err)
	}

	blockTime := ctx.BlockTime().Unix()

	profileJson := msg.ProfileJson
	// Create the profile
	profile := types.Profile{
		WalletAddress: msg.Creator,
		Nickname:      profileJson.Nickname,
		UserHandle:    profileJson.UserHandle,
		Avatar:        profileJson.Avatar,
		CreationTime:  blockTime,
	}

	ms.k.SetProfile(ctx, profile)

	//Emit an event for the creation
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeAddProfile,
			sdk.NewAttribute(types.AttributeKeyCreator, msg.Creator),
			sdk.NewAttribute(types.AttributeKeyNickname, profileJson.Nickname),
			sdk.NewAttribute(types.AttributeKeyUserHandle, profileJson.UserHandle),
			sdk.NewAttribute(types.AttributeKeyAvatar, profileJson.Avatar),
			sdk.NewAttribute(types.AttributeKeyTimestamp, fmt.Sprintf("%d", blockTime)),
		),
	})
	return &types.MsgAddProfileResponse{}, nil
}
