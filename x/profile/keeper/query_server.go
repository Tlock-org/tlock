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
