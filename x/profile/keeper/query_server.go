package keeper

import (
	"context"
	"fmt"

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
		return nil, err
	}

	return &types.QueryParamsResponse{Params: &p}, nil
}

// QueryProfile implements types.QueryServer.
func (k Querier) QueryProfile(goCtx context.Context, req *types.QueryProfileRequest) (*types.QueryProfileResponse, error) {
	//ctx := sdk.UnwrapSDKContext(goCtx)
	//if req == nil {
	//	return nil, types.ErrInvalidRequest
	//}

	// Retrieve the post from the state
	profile, _ := k.Keeper.GetProfile(sdk.UnwrapSDKContext(goCtx), req.WalletAddress)
	//if !found {
	//	return nil, types.ErrPostNotFound
	//}

	return &types.QueryProfileResponse{Profile: &profile}, nil
}

// QueryFollowing implements types.QueryServer.
func (k Querier) QueryFollowing(goCtx context.Context, req *types.QueryFollowingRequest) (*types.QueryFollowingResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	followings := k.Keeper.GetFollowing(ctx, req.WalletAddress)

	var profiles []*types.Profile
	for _, address := range followings {
		profile, success := k.GetProfile(ctx, address)
		if !success {
			return nil, fmt.Errorf("failed to get profile %s: %w", profile, success)
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
	followers := k.Keeper.GetFollowers(ctx, req.WalletAddress)

	var profiles []*types.Profile
	for _, address := range followers {
		profile, success := k.GetProfile(ctx, address)
		if !success {
			return nil, fmt.Errorf("failed to get profile %s: %w", profile, success)
		}
		profileCopy := profile
		profiles = append(profiles, &profileCopy)
	}
	return &types.QueryFollowersResponse{
		Profiles: profiles,
	}, nil
}
