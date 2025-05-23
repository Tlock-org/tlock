package keeper

import (
	"context"
	"cosmossdk.io/errors"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/rollchains/tlock/x/profile/types"
	"math/rand"
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
								Avatar:        profileJson.Avatar,
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
							Avatar:        profileJson.Avatar,
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
		validateNickName, err := types.ValidateNickName(nickname)
		if validateNickName {
			if dbNickname != nickname {
				ms.k.DeleteFromUserSearchList(ctx, dbNickname, msg.Creator)
				userSearch := types.UserSearch{
					UserHandle:    userHandle,
					WalletAddress: msg.Creator,
					Nickname:      nickname,
					Avatar:        profileJson.Avatar,
				}
				ms.k.AddToUserSearchList(ctx, nickname, userSearch)
			}
		} else {
			return &types.MsgAddProfileResponse{}, err
		}
	}

	dbProfile.WalletAddress = msg.Creator
	dbProfile.Nickname = profileJson.Nickname
	if profileJson.Avatar != "" {
		//dbProfile.Avatar = profileJson.Avatar
		ms.k.SetProfileAvatar(ctx, msg.Creator, profileJson.Avatar)
	}
	dbProfile.Bio = profileJson.Bio
	dbProfile.Location = profileJson.Location
	dbProfile.Website = profileJson.Website
	dbProfile.CreationTime = blockTime

	if msg.Creator == "tlock1hj5fveer5cjtn4wd6wstzugjfdxzl0xp5u7j9p" {
		if dbProfile.Level < 6 {
			dbProfile.Level = 6
		}
		if dbProfile.AdminLevel < 5 {
			dbProfile.AdminLevel = 5
		}
	}
	if msg.Creator == "tlock1efd63aw40lxf3n4mhf7dzhjkr453axurggdkvg" {
		if dbProfile.Level < 6 {
			dbProfile.Level = 6
		}
		if dbProfile.AdminLevel < 3 {
			dbProfile.AdminLevel = 3
		}
	}
	if msg.Creator == "tlock1qvmhf9qw5xhefm6jpqpggneahanjhkgw0szlzn" {
		if dbProfile.Level < 5 {
			dbProfile.Level = 5
		}
		if dbProfile.AdminLevel < 4 {
			dbProfile.AdminLevel = 4
		}
	}
	if msg.Creator == "tlock16f7etm42yp5nup77q3027rvvkl73q2gr8wkjcm" {
		if dbProfile.Level < 5 {
			dbProfile.Level = 5
		}
		if dbProfile.AdminLevel < 3 {
			dbProfile.AdminLevel = 3
		}
	}
	if msg.Creator == "tlock1wfvjqmkekyuy59r535nm2ca3yjkf706nu8x49r" {
		if dbProfile.Level < 4 {
			dbProfile.Level = 4
		}
		if dbProfile.AdminLevel < 5 {
			dbProfile.AdminLevel = 5
		}
	}

	if dbProfile.Level < 3 {
		dbProfile.Level = uint64(rand.Intn(3) + 3)
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

		ms.k.AddToFollowingSearch(sdkCtx, follower, profileTarget)

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
			ms.k.DeleteLastActivitiesReceived(sdkCtx, targetAddr)
		}
		ms.k.SetActivitiesReceivedCount(sdkCtx, targetAddr, count)

		ms.k.SetFollowTime(sdkCtx, follower, targetAddr)
	}

	return &types.MsgFollowResponse{}, nil
}

// Unfollow implements types.MsgServer.
func (ms msgServer) Unfollow(ctx context.Context, msg *types.MsgUnfollowRequest) (*types.MsgUnfollowResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	follower := msg.Creator
	targetAddr := msg.TargetAddr
	isFollowing := ms.k.IsFollowing(sdkCtx, follower, targetAddr)
	if isFollowing {
		time, _ := ms.k.GetFollowTime(sdkCtx, follower, targetAddr)
		ms.k.Unfollow(sdkCtx, follower, time, targetAddr)
		ms.k.DeleteFollowTime(sdkCtx, follower, targetAddr)

		profileFollower, _ := ms.k.GetProfile(sdkCtx, follower)
		following := profileFollower.Following
		if following > 0 {
			following -= 1
			profileFollower.Following = following
			ms.k.SetProfile(sdkCtx, profileFollower)
		}

		profileTarget, _ := ms.k.GetProfile(sdkCtx, targetAddr)
		followers := profileTarget.Followers
		if followers > 0 {
			followers -= 1
			profileTarget.Followers = followers
			ms.k.SetProfile(sdkCtx, profileTarget)
		}
		ms.k.DeleteFromFollowingSearch(sdkCtx, follower, profileTarget)
	}

	return &types.MsgUnfollowResponse{}, nil
}

// AddAdmin implements types.MsgServer.
func (ms msgServer) AddAdmin(goCtx context.Context, msg *types.MsgAddAdminRequest) (*types.MsgAddAdminResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	adminAddress, _ := ms.k.GetAdminAddress(ctx)
	if msg.Creator == adminAddress {
		err := ms.k.AddAdmin(ctx, msg.Address)
		if err != nil {
			return nil, err
		}
	}
	return &types.MsgAddAdminResponse{
		Status: true,
	}, nil
}

// RemoveAdmin implements types.MsgServer.
func (ms msgServer) RemoveAdmin(goCtx context.Context, msg *types.MsgRemoveAdminRequest) (*types.MsgRemoveAdminResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	adminAddress, _ := ms.k.GetAdminAddress(ctx)
	if msg.Creator == adminAddress {
		err := ms.k.RemoveAdmin(ctx, msg.Address)
		if err != nil {
			return nil, err
		}
	} else {
		return &types.MsgRemoveAdminResponse{
			Status: false,
		}, fmt.Errorf("address %s is not an admin", msg.Creator)
	}
	return &types.MsgRemoveAdminResponse{
		Status: true,
	}, nil
}

// AppointAdmin implements types.MsgServer.
func (ms msgServer) ManageAdmin(goCtx context.Context, msg *types.MsgManageAdminRequest) (*types.MsgManageAdminResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	moderator, _ := ms.k.GetChiefModerator(ctx)
	creator := msg.Creator
	creatorProfile, _ := ms.k.GetProfile(ctx, creator)
	json := msg.ManageJson
	address := json.AdminAddress
	adminProfile, _ := ms.k.GetProfile(ctx, address)
	action := msg.Action
	if action == types.AdminActionAppoint {
		lineProfile, _ := ms.k.GetProfile(ctx, json.LineManager)
		level := json.AdminLevel
		if (creatorProfile.AdminLevel > 4 || moderator == creator) || (json.LineManager != creator && address != json.LineManager && lineProfile.LineManager == creator && lineProfile.AdminLevel > 1 && lineProfile.AdminLevel < creatorProfile.AdminLevel) {
			lineAddr := json.LineManager
			adminProfile.LineManager = lineAddr
			adminProfile.AdminLevel = level
			ms.k.SetProfile(ctx, adminProfile)
			editable := json.Editable
			if editable {
				err := ms.k.AddEditableAdmin(ctx, address)
				if err != nil {
					return nil, err
				}
			}
		} else {
			return &types.MsgManageAdminResponse{
				Status: false,
			}, errors.Wrapf(types.ErrRequestDenied, "request denied")
		}
	} else if action == types.AdminActionRemove {
		lineProfile, _ := ms.k.GetProfile(ctx, adminProfile.LineManager)
		if (creatorProfile.AdminLevel > 4 || moderator == creator) || lineProfile.LineManager == creator {
			adminProfile.LineManager = ""
			adminProfile.AdminLevel = 0
			ms.k.SetProfile(ctx, adminProfile)
			editable := ms.k.IsEditableAdmin(ctx, address)
			if editable {
				err := ms.k.RemoveEditableAdmin(ctx, address)
				if err != nil {
					return nil, err
				}
			}
		} else {
			return &types.MsgManageAdminResponse{
				Status: false,
			}, errors.Wrapf(types.ErrRequestDenied, "request denied")
		}
	} else {
		return &types.MsgManageAdminResponse{
			Status: false,
		}, errors.Wrapf(types.ErrRequestDenied, "request denied")
	}

	return &types.MsgManageAdminResponse{
		Status: true,
	}, nil

}
