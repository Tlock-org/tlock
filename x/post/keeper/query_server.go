package keeper

import (
	"context"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	profilekeeper "github.com/rollchains/tlock/x/profile/keeper"
	"google.golang.org/grpc/codes"

	"github.com/rollchains/tlock/x/post/types"

	"google.golang.org/grpc/status"
)

var _ types.QueryServer = Querier{}

type Querier struct {
	Keeper
	ProfileKeeper profilekeeper.Keeper
}

func NewQuerier(keeper Keeper, pk profilekeeper.Keeper) Querier {
	return Querier{
		Keeper:        keeper,
		ProfileKeeper: pk,
	}
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

// QueryHomePosts implements types.QueryServer.
func (k Querier) QueryHomePosts(goCtx context.Context, req *types.QueryHomePostsRequest) (*types.QueryHomePostsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	postIDs, _, err := k.Keeper.GetHomePosts(ctx, req)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var postResponses []*types.PostResponse
	for _, postID := range postIDs {
		post, success := k.GetPost(ctx, postID)
		if !success {
			return nil, fmt.Errorf("failed to get post with ID %s: %w", postID, success)
		}
		postCopy := post

		profile, _ := k.ProfileKeeper.GetProfile(ctx, post.Creator)

		profileResponseCopy := profile

		postResponse := types.PostResponse{
			Post:    &postCopy,
			Profile: &profileResponseCopy,
		}
		if post.Quote != "" {
			quotePost, _ := k.GetPost(ctx, post.Quote)
			quoteProfile, _ := k.ProfileKeeper.GetProfile(ctx, quotePost.Creator)
			postResponse.QuotePost = &quotePost
			postResponse.QuoteProfile = &quoteProfile
		}

		postResponses = append(postResponses, &postResponse)
	}

	return &types.QueryHomePostsResponse{
		Posts: postResponses,
	}, nil
}

// QueryFirstPageHomePosts implements types.QueryServer.
func (k Querier) QueryFirstPageHomePosts(goCtx context.Context, req *types.QueryFirstPageHomePostsRequest) (*types.QueryFirstPageHomePostsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	postIDs, _, err := k.Keeper.GetFirstPageHomePosts(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var postResponses []*types.PostResponse
	for _, postID := range postIDs {
		post, success := k.GetPost(ctx, postID)
		if !success {
			return nil, fmt.Errorf("failed to get post with ID %s: %w", postID, success)
		}
		postCopy := post

		profile, _ := k.ProfileKeeper.GetProfile(ctx, post.Creator)
		profileResponseCopy := profile
		postResponse := types.PostResponse{
			Post:    &postCopy,
			Profile: &profileResponseCopy,
		}

		if post.Quote != "" {
			quotePost, _ := k.GetPost(ctx, post.Quote)
			quoteProfile, _ := k.ProfileKeeper.GetProfile(ctx, quotePost.Creator)
			postResponse.QuotePost = &quotePost
			postResponse.QuoteProfile = &quoteProfile
		}
		postResponses = append(postResponses, &postResponse)
	}
	return &types.QueryFirstPageHomePostsResponse{
		Posts: postResponses,
	}, nil
}

// QueryUserCreatedPosts implements types.QueryServer.
func (k Querier) QueryUserCreatedPosts(goCtx context.Context, req *types.QueryUserCreatedPostsRequest) (*types.QueryUserCreatedPostsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	postIDs, _, err := k.Keeper.GetUserCreatedPosts(ctx, req.Wallet)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var postResponses []*types.PostResponse
	for _, postID := range postIDs {
		post, success := k.GetPost(ctx, postID)
		if !success {
			return nil, fmt.Errorf("failed to get post with ID %s: %w", postID, success)
		}
		postCopy := post
		//posts = append(posts, &postCopy)

		profile, _ := k.ProfileKeeper.GetProfile(ctx, post.Creator)
		profileResponseCopy := profile
		postResponse := types.PostResponse{
			Post:    &postCopy,
			Profile: &profileResponseCopy,
		}

		if post.Quote != "" {
			quotePost, _ := k.GetPost(ctx, post.Quote)
			quoteProfile, _ := k.ProfileKeeper.GetProfile(ctx, quotePost.Creator)
			postResponse.QuotePost = &quotePost
			postResponse.QuoteProfile = &quoteProfile
		}
		postResponses = append(postResponses, &postResponse)
	}
	return &types.QueryUserCreatedPostsResponse{
		Posts: postResponses,
	}, nil
}

// QueryPost implements types.QueryServer.
func (k Querier) QueryPost(goCtx context.Context, req *types.QueryPostRequest) (*types.QueryPostResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidRequest
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Retrieve the post from the state
	post, found := k.Keeper.GetPost(sdk.UnwrapSDKContext(goCtx), req.PostId)
	if !found {
		return nil, types.ErrPostNotFound
	}

	postCopy := post
	//posts = append(posts, &postCopy)

	profile, _ := k.ProfileKeeper.GetProfile(ctx, post.Creator)
	//k.Logger().Error("======nickName:{}", profile.Nickname)
	//profileResponse := profileTypes.ProfileResponse{
	//	UserHandle: profile.UserHandle,
	//	Nickname:   profile.Nickname,
	//	Avatar:     profile.Avatar,
	//	Level:      profile.Level,
	//	AdminLevel: profile.AdminLevel,
	//}
	profileResponseCopy := profile
	postResponse := types.PostResponse{
		Post:    &postCopy,
		Profile: &profileResponseCopy,
	}

	return &types.QueryPostResponse{Post: &postResponse}, nil
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

	var postResponses []*types.PostResponse
	for _, likesIMade := range likes {
		postID := likesIMade.PostId
		post, success := k.GetPost(sdkCtx, postID)
		if !success {
			return nil, fmt.Errorf("failed to get post with ID %s: %w", postID, success)
		}
		postCopy := post
		//posts = append(posts, &postCopy)

		profile, _ := k.ProfileKeeper.GetProfile(sdkCtx, post.Creator)
		//profileResponse := profileTypes.ProfileResponse{
		//	UserHandle: profile.UserHandle,
		//	Nickname:   profile.Nickname,
		//	Avatar:     profile.Avatar,
		//	Level:      profile.Level,
		//	AdminLevel: profile.AdminLevel,
		//}
		profileResponseCopy := profile
		postResponse := types.PostResponse{
			Post:    &postCopy,
			Profile: &profileResponseCopy,
		}

		postResponses = append(postResponses, &postResponse)
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
		Posts: postResponses,
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

	var postResponses []*types.PostResponse
	for _, savesIMade := range saves {
		postID := savesIMade.PostId
		post, success := k.GetPost(sdkCtx, postID)
		if !success {
			return nil, fmt.Errorf("failed to get post with ID %s: %w", postID, success)
		}
		postCopy := post
		//posts = append(posts, &postCopy)

		profile, _ := k.ProfileKeeper.GetProfile(sdkCtx, post.Creator)
		//profileResponse := profileTypes.ProfileResponse{
		//	UserHandle: profile.UserHandle,
		//	Nickname:   profile.Nickname,
		//	Avatar:     profile.Avatar,
		//	Level:      profile.Level,
		//	AdminLevel: profile.AdminLevel,
		//}
		profileResponseCopy := profile
		postResponse := types.PostResponse{
			Post:    &postCopy,
			Profile: &profileResponseCopy,
		}

		postResponses = append(postResponses, &postResponse)
	}

	return &types.SavesIMadeResponse{
		Posts: postResponses,
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

// QueryComments implements types.QueryServer.
func (k Querier) QueryComments(goCtx context.Context, req *types.QueryCommentsRequest) (*types.QueryCommentsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	ids := k.Keeper.GetCommentsByParentId(ctx, req.Id)

	var commentResponses []*types.CommentResponse
	for _, commentId := range ids {
		comment, success := k.GetPost(ctx, commentId)
		if !success {
			return nil, fmt.Errorf("failed to get post with ID %s: %w", commentId, success)
		}
		commentCopy := comment
		profile, _ := k.ProfileKeeper.GetProfile(ctx, comment.Creator)
		pid := comment.ParentId
		parent, b := k.GetPost(ctx, pid)
		if !b {
			return nil, fmt.Errorf("failed to get post with ID %s: %w", pid, b)
		}
		targetProfile, _ := k.ProfileKeeper.GetProfile(ctx, parent.Creator)
		profileResponseCopy := profile
		targetProfileCopy := targetProfile
		commentResponse := types.CommentResponse{
			Post:          &commentCopy,
			Profile:       &profileResponseCopy,
			TargetProfile: &targetProfileCopy,
		}
		commentResponses = append(commentResponses, &commentResponse)
	}
	return &types.QueryCommentsResponse{
		Comments: commentResponses,
	}, nil
}

// QueryCommentsReceived implements types.QueryServer.
func (k Querier) QueryCommentsReceived(goCtx context.Context, req *types.QueryCommentsReceivedRequest) (*types.QueryCommentsReceivedResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	ids, err := k.Keeper.GetCommentsReceived(ctx, req.Wallet)
	if err != nil {
		return nil, fmt.Errorf("failed to get comments received %s: %w", req.Wallet, err)
	}

	var commentReceivedResponses []*types.CommentReceivedResponse
	for _, commentID := range ids {
		comment, success := k.GetPost(ctx, commentID)
		if !success {
			return nil, fmt.Errorf("failed to get comments received %s: %w", commentID, success)
		}
		commentCopy := comment
		postParent, _ := k.GetPost(ctx, comment.ParentId)
		postParentCopy := postParent
		CommentReceivedResponse := types.CommentReceivedResponse{
			Comment: &commentCopy,
			Parent:  &postParentCopy,
		}
		commentReceivedResponses = append(commentReceivedResponses, &CommentReceivedResponse)
	}
	return &types.QueryCommentsReceivedResponse{
		Comments: commentReceivedResponses,
	}, nil
}
