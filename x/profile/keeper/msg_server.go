package keeper

import (
	"context"
	"cosmossdk.io/errors"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/rollchains/tlock/x/profile/types"
	"strings"
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
	dbUserHandle := dbProfile.UserHandle
	dbNickname := dbProfile.Nickname

	userHandle := profileJson.UserHandle
	nickname := profileJson.Nickname
	if userHandle != "" {
		validateUserHandle, err := types.ValidateUserHandle(userHandle)
		if validateUserHandle {
			userHandle = strings.ToLower(userHandle)
			if dbUserHandle != userHandle {
				exist := ms.k.HasUserHandle(ctx, userHandle)
				if exist {
					if dbUserHandle == "" {
						suffixHandle := ms.k.TruncateAddressSuffix(dbProfile.WalletAddress)
						suffixExist := ms.k.HasUserHandle(ctx, suffixHandle)
						if suffixExist {
							return nil, errors.Wrapf(types.ErrInvalidUserHandle, "userHandle unavailable: %s", err)
						} else {
							ms.k.AddToUserHandleList(ctx, suffixHandle, msg.Creator)
							dbProfile.UserHandle = suffixHandle

							// add userHandle to userSearch
							ms.k.DeleteFromUserSearchList(ctx, suffixHandle, msg.Creator)
							userSearch := types.UserSearch{
								UserHandle:    suffixHandle,
								WalletAddress: nickname,
								Nickname:      msg.Creator,
							}
							ms.k.AddToUserSearchList(ctx, suffixHandle, userSearch)
						}
					}
				} else {
					if dbUserHandle != "" {
						ms.k.DeleteFromUserHandleList(ctx, dbUserHandle)
						ms.k.AddToUserHandleList(ctx, userHandle, msg.Creator)
						dbProfile.UserHandle = userHandle

						// add userHandle to userSearch
						ms.k.DeleteFromUserSearchList(ctx, dbUserHandle, msg.Creator)
						userSearch := types.UserSearch{
							UserHandle:    userHandle,
							WalletAddress: msg.Creator,
							Nickname:      nickname,
						}
						ms.k.AddToUserSearchList(ctx, userHandle, userSearch)
					}
				}

			} else {

			}
		} else {
			return &types.MsgAddProfileResponse{}, err
		}
	}

	// add nickName to userSearch
	if nickname != "" {
		if dbNickname != nickname {
			ms.k.DeleteFromUserSearchList(ctx, dbNickname, msg.Creator)
			userSearch := types.UserSearch{
				userHandle,
				nickname,
				msg.Creator,
			}
			ms.k.AddToUserSearchList(ctx, nickname, userSearch)
		}
	}

	dbProfile.WalletAddress = msg.Creator
	dbProfile.Nickname = profileJson.Nickname
	dbProfile.Avatar = profileJson.Avatar
	dbProfile.Bio = profileJson.Bio
	dbProfile.Location = profileJson.Location
	dbProfile.Website = profileJson.Website
	dbProfile.CreationTime = blockTime

	if msg.Creator == "tlock1hj5fveer5cjtn4wd6wstzugjfdxzl0xp5u7j9p" || msg.Creator == "tlock1efd63aw40lxf3n4mhf7dzhjkr453axurggdkvg" {
		dbProfile.Level = 6
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
	if follower == targetAddr {
		return nil, errors.Wrap(types.ErrCannotFollowSelf, "follower and targetAddr cannot be the same")
	}
	isFollowing := ms.k.IsFollowing(sdkCtx, follower, targetAddr)
	if !isFollowing {
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
			Address:        follower,
			TargetAddress:  targetAddr,
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
	}

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
