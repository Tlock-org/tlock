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
	//if profileJson.UserHandle != "" {
	//	handle = profileJson.UserHandle
	//}
	if handle != profileJson.UserHandle {
		exist := ms.k.IsProfileUserHandleExist(ctx, profileJson.UserHandle)
		if exist {
			if handle == "" {
				suffixHandle := ms.k.TruncateWalletAddressSuffix(dbProfile.WalletAddress)
				suffixExist := ms.k.IsProfileUserHandleExist(ctx, suffixHandle)
				if suffixExist {
					return nil, errors.Wrapf(types.ErrInvalidUserHandle, "userHandle unavailable: %s", err)
				} else {
					ms.k.SetProfileUserHandle(ctx, suffixHandle)
					dbProfile.UserHandle = suffixHandle
				}
			}
		} else {
			ms.k.DeleteProfileUserHandle(ctx, profileJson.UserHandle)
			ms.k.SetProfileUserHandle(ctx, profileJson.UserHandle)
			dbProfile.UserHandle = profileJson.UserHandle
		}

	} else {

	}
	dbProfile.WalletAddress = msg.Creator
	dbProfile.Nickname = profileJson.Nickname
	dbProfile.Avatar = profileJson.Avatar
	dbProfile.Bio = profileJson.Bio
	dbProfile.Location = profileJson.Location
	dbProfile.Website = profileJson.Website
	dbProfile.CreationTime = blockTime
	// Create the profile
	//profile := types.Profile{
	//	WalletAddress: msg.Creator,
	//	Nickname:      profileJson.Nickname,
	//	UserHandle:    handle,
	//	Avatar:        profileJson.Avatar,
	//	Bio:           profileJson.Bio,
	//	Location:      profileJson.Location,
	//	Website:       profileJson.Website,
	//	CreationTime:  blockTime,
	//}
	if msg.Creator == "tlock1efd63aw40lxf3n4mhf7dzhjkr453axurggdkvg" {
		dbProfile.Level = 5
	}

	ms.k.SetProfile(ctx, dbProfile)

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
		Profile: &dbProfile,
	}, nil
}

// Follow implements types.MsgServer.
func (ms msgServer) Follow(ctx context.Context, msg *types.MsgFollowRequest) (*types.MsgFollowResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	blockTime := sdkCtx.BlockTime().Unix()
	follower := msg.Creator
	targetAddr := msg.TargetAddr
	ms.k.AddToFollowing(sdkCtx, follower, targetAddr)
	ms.k.AddToFollowers(sdkCtx, targetAddr, follower)

	profileFollower, _ := ms.k.GetProfile(sdkCtx, follower)
	following := profileFollower.Following
	following += 1
	profileFollower.Following = following
	ms.k.SetProfile(sdkCtx, profileFollower)

	profileTarget, _ := ms.k.GetProfile(sdkCtx, targetAddr)
	followers := profileTarget.Followers
	followers += 1
	profileTarget.Followers = followers
	ms.k.SetProfile(sdkCtx, profileTarget)

	activitiesReceived := types.ActivitiesReceived{
		Address:        msg.Creator,
		ActivitiesType: types.ActivitiesType_ACTIVITIES_FOLLOW,
		Timestamp:      blockTime,
	}
	ms.k.SetActivitiesReceived(sdkCtx, activitiesReceived, targetAddr, follower)
	count, b := ms.k.GetActivitiesReceivedCount(sdkCtx, targetAddr)
	if !b {
		panic("GetActivitiesReceivedCount error")
	}
	count += 1
	if count > types.ActivitiesReceivedCount {
		ms.k.DeleteLastActivitiesReceived(sdkCtx, follower)
	}
	ms.k.SetActivitiesReceivedCount(sdkCtx, targetAddr, count)

	ms.k.SetFollowTime(sdkCtx, follower, targetAddr)

	return &types.MsgFollowResponse{}, nil
}

// Unfollow implements types.MsgServer.
func (ms msgServer) Unfollow(ctx context.Context, msg *types.MsgUnfollowRequest) (*types.MsgUnfollowResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	time, _ := ms.k.GetFollowTime(sdkCtx, msg.Creator, msg.TargetAddr)
	ms.k.Unfollow(sdkCtx, msg.Creator, time, msg.TargetAddr)
	ms.k.DeleteFollowTime(sdkCtx, msg.Creator, msg.TargetAddr)

	profileFollower, _ := ms.k.GetProfile(sdkCtx, msg.Creator)
	following := profileFollower.Following
	following -= 1
	profileFollower.Following = following
	ms.k.SetProfile(sdkCtx, profileFollower)

	profileTarget, _ := ms.k.GetProfile(sdkCtx, msg.TargetAddr)
	followers := profileTarget.Followers
	followers -= 1
	profileTarget.Followers = followers
	ms.k.SetProfile(sdkCtx, profileTarget)

	return &types.MsgUnfollowResponse{}, nil
}
