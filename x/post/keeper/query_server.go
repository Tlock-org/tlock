package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/rollchains/tlock/x/post/types"
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

// ResolveName implements types.QueryServer.
func (k Querier) ResolveName(goCtx context.Context, req *types.QueryResolveNameRequest) (*types.QueryResolveNameResponse, error) {
	// ctx := sdk.UnwrapSDKContext(goCtx)
	//panic("ResolveName is unimplemented")
	//return &types.QueryResolveNameResponse{}, nil
	v, err := k.Keeper.NameMapping.Get(goCtx, req.Wallet)
	if err != nil {
		return nil, err
	}

	return &types.QueryResolveNameResponse{
		Name: v,
	}, nil
}

// QueryPost implements types.QueryServer.
func (k Querier) QueryPost(goCtx context.Context, req *types.QueryPostRequest) (*types.QueryPostResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidRequest
	}

	// Retrieve the post from the state
	post, found := k.Keeper.GetPost(sdk.UnwrapSDKContext(goCtx), req.PostId)
	if !found {
		return nil, types.ErrPostNotFound
	}

	return &types.QueryPostResponse{Post: &post}, nil
}
