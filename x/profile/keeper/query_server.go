package keeper

import (
	"context"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/rollchains/tlock/x/profile/types"
)

var _ types.QueryServer = Querier{}

type Querier struct {
	Keeper
}

func NewQuerier(keeper Keeper) Querier {
	return Querier{Keeper: keeper}
}

func (k Querier) Params(c context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	p, err := k.Keeper.Params.Get(ctx)
	if err != nil {
		types.LogError(k.logger, "get_params", err)
		return nil, types.ToGRPCError(types.WrapError(types.ErrDatabaseOperation, "failed to get params"))
	}

	return &types.QueryParamsResponse{Params: &p}, nil
}

// QueryProfile implements types.QueryServer.
func (k Querier) QueryProfile(goCtx context.Context, req *types.QueryProfileRequest) (*types.QueryProfileResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if req == nil {
		return nil, types.ToGRPCError(types.ErrInvalidRequest)
	}

	// Retrieve the post from the state
	profile, _ := k.Keeper.GetProfile(sdk.UnwrapSDKContext(goCtx), req.Address)
	//if !found {
	//	return nil, types.ErrPostNotFound
	//}
	avatar := k.Keeper.GetAvatarByAddress(ctx, req.Address)
	profile.Avatar = avatar

	return &types.QueryProfileResponse{Profile: &profile}, nil
}

// QueryProfileAvatar implements types.QueryServer.
func (k Querier) QueryProfileAvatar(goCtx context.Context, req *types.QueryProfileAvatarRequest) (*types.QueryProfileAvatarResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	avatar := k.GetAvatarByAddress(ctx, req.Address)
	return &types.QueryProfileAvatarResponse{
		Avatar: avatar,
	}, nil
}

// QueryFollowRelationship implements types.QueryServer.
func (k Querier) QueryFollowRelationship(goCtx context.Context, req *types.QueryFollowRelationshipRequest) (*types.QueryFollowRelationshipResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	aFollowsB := k.Keeper.IsFollowing(ctx, req.AddressA, req.AddressB)
	bFollowsA := k.Keeper.IsFollowing(ctx, req.AddressB, req.AddressA)
	relationship := uint64(0)
	if aFollowsB {
		if bFollowsA {
			relationship = 3
		} else {
			relationship = 1
		}
	} else {
		if bFollowsA {
			relationship = 2
		}
	}
	return &types.QueryFollowRelationshipResponse{
		Relationship: relationship,
	}, nil
}

// QueryFollowing implements types.QueryServer.
func (k Querier) QueryFollowing(goCtx context.Context, req *types.QueryFollowingRequest) (*types.QueryFollowingResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	followings, _, _ := k.Keeper.GetFollowingPagination(ctx, req.Address, req.Page, req.Limit)
	var profiles []*types.Profile
	for _, address := range followings {
		profile, success := k.GetProfile(ctx, address)
		if !success {
			types.LogError(k.logger, "query_following", types.ErrProfileNotFound, "address", address)
			return nil, types.ToGRPCError(types.NewProfileNotFoundError(address))
		}
		profileCopy := profile
		profiles = append(profiles, &profileCopy)
	}
	return &types.QueryFollowingResponse{
		Profiles: profiles,
	}, nil
}

// QueryFollowers implements types.QueryServer.
func (k Querier) QueryFollowers(goCtx context.Context, req *types.QueryFollowersRequest) (*types.QueryFollowersResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	followers := k.Keeper.GetFollowers(ctx, req.Address)

	var profiles []*types.Profile
	for _, address := range followers {
		profile, success := k.GetProfile(ctx, address)
		if !success {
			types.LogError(k.logger, "query_followers", types.ErrProfileNotFound, "address", address)
			return nil, types.ToGRPCError(types.NewProfileNotFoundError(address))
		}
		profileCopy := profile
		profiles = append(profiles, &profileCopy)
	}
	return &types.QueryFollowersResponse{
		Profiles: profiles,
	}, nil
}

// GetMentionSuggestions implements types.QueryServer.
func (k Querier) GetMentionSuggestions(goCtx context.Context, req *types.QueryGetMentionSuggestionsRequest) (*types.QueryGetMentionSuggestionsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	var profiles []*types.Profile
	addedAddresses := make(map[string]bool)
	addresses, _ := k.Keeper.GetFollowingSearch(ctx, req.Address, req.Matching)
	if len(addresses) < 10 {
		users, _ := k.SearchUsersLimitByMatching(ctx, req.Matching, 10-len(addresses))
		for _, address := range users {
			addresses = append(addresses, address)
		}
	}
	for _, address := range addresses {
		if addedAddresses[address] {
			continue
		}
		profile, success := k.GetProfile(ctx, address)
		if !success {
			types.LogError(k.logger, "get_mention_suggestions", types.ErrProfileNotFound, "address", address)
			return nil, types.ToGRPCError(types.NewProfileNotFoundError(address))
		}
		profileCopy := profile
		profiles = append(profiles, &profileCopy)
		addedAddresses[address] = true
	}
	return &types.QueryGetMentionSuggestionsResponse{
		Profiles: profiles,
	}, nil
}

// QueryActivitiesReceived implements types.QueryServer.
//func (k Querier) QueryActivitiesReceived(goCtx context.Context, req *types.QueryActivitiesReceivedRequest) (*types.QueryActivitiesReceivedResponse, error) {
//	ctx := sdk.UnwrapSDKContext(goCtx)
//
//	list := k.GetActivitiesReceived(ctx, req.WalletAddress)
//	var activitiesReceivedList []*types.ActivitiesReceivedResponse
//	for _, activitiesReceived := range list {
//		activitiesReceivedResponse := types.ActivitiesReceivedResponse{
//			Address:        activitiesReceived.Address,
//			TargetAddress:  activitiesReceived.TargetAddress,
//			CommentId:      activitiesReceived.CommentId,
//			ParentId:       activitiesReceived.ParentId,
//			ActivitiesType: activitiesReceived.ActivitiesType,
//			Content:        activitiesReceived.Content,
//			ParentImageUrl: activitiesReceived.ParentImageUrl,
//			Timestamp:      activitiesReceived.Timestamp,
//		}
//
//		//parentId := activitiesReceived.ParentId
//		//if parentId != "" {
//		//
//		//}
//
//		address := activitiesReceived.Address
//		if address != "" {
//			profile, _ := k.GetProfile(ctx, address)
//			profileResponse := types.ProfileResponse{
//				UserHandle: profile.UserHandle,
//				Nickname:   profile.Nickname,
//				Avatar:     profile.Avatar,
//			}
//			activitiesReceivedResponse.Profile = &profileResponse
//		}
//		targetAddress := activitiesReceived.TargetAddress
//		if targetAddress != "" {
//			targetProfile, _ := k.GetProfile(ctx, targetAddress)
//			targetProfileResponse := types.ProfileResponse{
//				UserHandle: targetProfile.UserHandle,
//				Nickname:   targetProfile.Nickname,
//				Avatar:     targetProfile.Avatar,
//			}
//			activitiesReceivedResponse.TargetProfile = &targetProfileResponse
//		}
//
//		activitiesReceivedList = append(activitiesReceivedList, &activitiesReceivedResponse)
//	}
//	return &types.QueryActivitiesReceivedResponse{
//		activitiesReceivedList,
//	}, nil
//}

// QueryActivitiesReceivedCount implements types.QueryServer.
func (k Querier) QueryActivitiesReceivedCount(goCtx context.Context, req *types.QueryActivitiesReceivedCountRequest) (*types.QueryActivitiesReceivedCountResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	count, _ := k.GetActivitiesReceivedCount(ctx, req.Address)
	return &types.QueryActivitiesReceivedCountResponse{
		Count: uint64(count),
	}, nil
}

// QueryHasUserHandle implements types.QueryServer.
//func (k Querier) QueryHasUserHandle(goCtx context.Context, req *types.QueryHasUserHandleRequest) (*types.QueryHasUserHandleResponse, error) {
//	ctx := sdk.UnwrapSDKContext(goCtx)
//	//handle := k.HasUserHandle(ctx, req.UserHandle)
//	address := k.GetAddressByUserHandle(ctx, req.UserHandle)
//	return &types.QueryHasUserHandleResponse{
//		address,
//	}, nil
//}

// SearchUsers implements types.QueryServer.
func (k Querier) SearchUsers(goCtx context.Context, req *types.SearchUsersRequest) (*types.SearchUsersResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	matching := req.Matching
	if matching == "" {
		return &types.SearchUsersResponse{}, nil
	}
	users, _ := k.SearchUsersByMatching(ctx, matching)
	var list []*types.Profile
	for _, user := range users {
		profile, _ := k.GetProfile(ctx, user.WalletAddress)
		list = append(list, &profile)
	}
	return &types.SearchUsersResponse{
		Users: list,
	}, nil
}

// IsAdmin implements types.QueryServer.
func (k Querier) IsAdmin(goCtx context.Context, req *types.IsAdminRequest) (*types.IsAdminResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	isAdmin := k.Keeper.IsAdmin(ctx, req.Address)
	return &types.IsAdminResponse{
		Status: isAdmin,
	}, nil
}
