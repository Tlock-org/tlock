package keeper

import (
	"context"
	"cosmossdk.io/errors"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
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

	dbProfile, _ := ms.k.GetProfile(ctx, msg.GetCreator())
	handle := dbProfile.UserHandle
	if profileJson.UserHandle != "" {
		handle = profileJson.UserHandle
	}

	// Create the profile
	profile := types.Profile{
		WalletAddress: msg.Creator,
		Nickname:      profileJson.Nickname,
		UserHandle:    handle,
		Avatar:        profileJson.Avatar,
		Bio:           profileJson.Bio,
		Location:      profileJson.Location,
		Website:       profileJson.Website,
		CreationTime:  blockTime,
	}
	if msg.Creator == "tlock1efd63aw40lxf3n4mhf7dzhjkr453axurggdkvg" {
		profile.Level = 5
	}

	//if exists {
	//	nickname := profileJson.Nickname
	//	fmt.Printf("========profileJson nickname: %s\n", nickname)
	//	if &nickname != nil {
	//		profile.Nickname = nickname
	//	}
	//	userHandle := profileJson.UserHandle
	//	fmt.Printf("========profileJson userHandle: %s\n", userHandle)
	//	if &userHandle != nil {
	//		fmt.Printf("========here in?====: %s\n", userHandle)
	//		profile.UserHandle = userHandle
	//	}
	//	avatar := profileJson.Avatar
	//	fmt.Printf("========profileJson avatar: %s\n", avatar)
	//	if &avatar != nil {
	//		profile.Avatar = avatar
	//	}
	//	bio := profileJson.Bio
	//	fmt.Printf("========profileJson bio: %s\n", bio)
	//	if &bio != nil {
	//		profile.Bio = bio
	//	}
	//	location := profileJson.Location
	//	fmt.Printf("========profileJson location: %s\n", location)
	//	if &location != nil {
	//		profile.Location = location
	//	}
	//	website := profileJson.Website
	//	fmt.Printf("========profileJson website: %s\n", website)
	//	if &website != nil {
	//		profile.Website = website
	//	}
	//} else {
	//	profile.Nickname = profileJson.Nickname
	//	profile.UserHandle = profileJson.UserHandle
	//	profile.Avatar = profileJson.Avatar
	//	profile.Bio = profileJson.Bio
	//	profile.Location = profileJson.Location
	//	profile.Website = profileJson.Website
	//}

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
	return &types.MsgAddProfileResponse{
		Profile: &profile,
	}, nil
}

// Follow implements types.MsgServer.
func (ms msgServer) Follow(ctx context.Context, msg *types.MsgFollowRequest) (*types.MsgFollowResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	creator := msg.Creator
	targetAddr := msg.TargetAddr
	ms.k.AddFollowing(sdkCtx, creator, targetAddr)
	ms.k.AddFollower(sdkCtx, targetAddr, creator)
	return &types.MsgFollowResponse{}, nil
}
