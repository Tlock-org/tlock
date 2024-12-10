package keeper

import (
	"context"
	"google.golang.org/grpc/codes"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/rollchains/tlock/x/post/types"

	"google.golang.org/grpc/status"
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

// likesIMade implements types.QueryServer.
func (k Querier) LikesIMade(ctx context.Context, request *types.LikesIMadeRequest) (*types.LikesIMadeResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	k.Logger().Warn("======Processing LikesIMade query", "wallet", request.Wallet)
	likes, err := k.GetLikesIMade(sdkCtx, request.Wallet)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	//var likesList []*types.LikesIMade
	//pageRes, err := query.Paginate(
	//	prefix.NewStore(sdkCtx.KVStore(k.storeKey), []byte(types.LikesIMadePrefix+request.Wallet+"/")),
	//	request.Pagination,
	//	func(key []byte, value []byte) error {
	//		var like types.LikesIMade
	//		if err := k.cdc.Unmarshal(value, &like); err != nil {
	//			return err
	//		}
	//		likeCopy := like
	//		likesList = append(likesList, &likeCopy)
	//		return nil
	//	},
	//)
	//
	//if err != nil {
	//	return nil, status.Error(codes.Internal, err.Error())
	//}

	//return &types.LikesIMadeResponse{
	//	LikesMade:  likesList,
	//	Pagination: pageRes,
	//}, nil
	//var likesList []*types.LikesIMade
	//var like types.LikesIMade
	//likeCopy := like
	//likesList = append(likesList, &likeCopy)
	return &types.LikesIMadeResponse{
		LikesMade: likes,
	}, nil
}

// SavesIMade implements types.QueryServer.
func (k Querier) SavesIMade(ctx context.Context, request *types.SavesIMadeRequest) (*types.SavesIMadeResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	k.Logger().Warn("======Processing LikesIMade query", "wallet", request.Wallet)
	saves, err := k.GetSavesIMade(sdkCtx, request.Wallet)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.SavesIMadeResponse{
		SavesMade: saves,
	}, nil
}

// LikesReceived implements types.QueryServer.
func (k Querier) LikesReceived(ctx context.Context, request *types.LikesReceivedRequest) (*types.LikesReceivedResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	k.Logger().Warn("======Processing LikesReceived query", "wallet", request.Wallet)
	likesReceived, err := k.GetLikesReceived(sdkCtx, request.Wallet)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.LikesReceivedResponse{
		LikesReceived: likesReceived,
	}, nil
}
