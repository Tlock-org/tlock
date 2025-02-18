package keeper

import (
	"context"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	profilekeeper "github.com/rollchains/tlock/x/profile/keeper"
	"google.golang.org/grpc/codes"

	"github.com/rollchains/tlock/x/post/types"
	profileTypes "github.com/rollchains/tlock/x/profile/types"
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

// QueryTopicPosts implements types.QueryServer.
func (k Querier) QueryTopicPosts(goCtx context.Context, req *types.QueryTopicPostsRequest) (*types.QueryTopicPostsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	topicHash := req.TopicId
	//topicHash := k.sha256Generate(topic)
	postIDs, _, err := k.Keeper.GetTopicPosts(ctx, topicHash)
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
	return &types.QueryTopicPostsResponse{
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

// SearchTopic implements types.QueryServer.
func (k Querier) SearchTopics(goCtx context.Context, req *types.SearchTopicsRequest) (*types.SearchTopicsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	topicOriginalList, _ := k.SearchTopicMatches(ctx, req.Matching)
	var topicResponses []*types.TopicResponse
	for _, topicOriginal := range topicOriginalList {
		topicHash := k.sha256Generate(topicOriginal)
		topic, _ := k.GetTopic(ctx, topicHash)
		topicResponse := types.TopicResponse{
			Id:      topic.Id,
			Name:    topic.Name,
			Avatar:  topic.Avatar,
			Title:   topic.Title,
			Summary: topic.Summary,
			Score:   topic.Score,
			Hotness: topic.Hotness,
		}
		topicResponses = append(topicResponses, &topicResponse)
	}
	return &types.SearchTopicsResponse{
		Topics: topicResponses,
	}, nil
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

// QueryActivitiesReceived implements types.QueryServer.
func (k Querier) QueryActivitiesReceived(goCtx context.Context, req *types.QueryActivitiesReceivedRequest) (*types.QueryActivitiesReceivedResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	list := k.ProfileKeeper.GetActivitiesReceived(ctx, req.Address)
	var activitiesReceivedList []*types.ActivitiesReceivedResponse
	for _, activitiesReceived := range list {
		activitiesReceivedResponse := types.ActivitiesReceivedResponse{
			Address:        activitiesReceived.Address,
			TargetAddress:  activitiesReceived.TargetAddress,
			CommentId:      activitiesReceived.CommentId,
			ParentId:       activitiesReceived.ParentId,
			ActivitiesType: activitiesReceived.ActivitiesType,
			Content:        activitiesReceived.Content,
			ParentImageUrl: activitiesReceived.ParentImageUrl,
			Timestamp:      activitiesReceived.Timestamp,
		}

		parentId := activitiesReceived.ParentId
		if parentId != "" {
			parentPost, _ := k.GetPost(ctx, parentId)
			activitiesReceivedResponse.ParentPost = &parentPost
		}

		address := activitiesReceived.Address
		if address != "" {
			profile, _ := k.ProfileKeeper.GetProfile(ctx, address)
			profileResponse := profileTypes.ProfileResponse{
				UserHandle: profile.UserHandle,
				Nickname:   profile.Nickname,
				Avatar:     profile.Avatar,
			}
			activitiesReceivedResponse.Profile = &profileResponse
		}
		targetAddress := activitiesReceived.TargetAddress
		if targetAddress != "" {
			targetProfile, _ := k.ProfileKeeper.GetProfile(ctx, targetAddress)
			targetProfileResponse := profileTypes.ProfileResponse{
				UserHandle: targetProfile.UserHandle,
				Nickname:   targetProfile.Nickname,
				Avatar:     targetProfile.Avatar,
			}
			activitiesReceivedResponse.TargetProfile = &targetProfileResponse
		}

		activitiesReceivedList = append(activitiesReceivedList, &activitiesReceivedResponse)
	}
	return &types.QueryActivitiesReceivedResponse{
		activitiesReceivedList,
	}, nil
}

// QueryCategoryExists implements types.QueryServer.
func (k Querier) QueryCategoryExists(goCtx context.Context, req *types.QueryCategoryExistsRequest) (*types.QueryCategoryExistsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	category := req.Category
	id := k.sha256Generate(category)
	exists := k.CategoryExists(ctx, id)
	return &types.QueryCategoryExistsResponse{
		Exists: exists,
	}, nil
}

// QueryCategories implements types.QueryServer.
func (k Querier) QueryCategories(goCtx context.Context, req *types.QueryCategoriesRequest) (*types.QueryCategoriesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	categories := k.GetAllCategories(ctx)
	var categoryResponseList []*types.CategoryResponse
	for _, category := range categories {
		categoryResponse := types.CategoryResponse{
			Id:     category.Id,
			Name:   category.Name,
			Avatar: category.Avatar,
			Index:  category.Index,
		}
		categoryResponseList = append(categoryResponseList, &categoryResponse)
	}
	return &types.QueryCategoriesResponse{
		Categories: categoryResponseList,
	}, nil
}

// QueryTopicsByCategory implements types.QueryServer.
func (k Querier) QueryTopicsByCategory(goCtx context.Context, req *types.QueryTopicsByCategoryRequest) (*types.QueryTopicsByCategoryResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	category := k.GetCategory(ctx, req.CategoryId)
	topics, _, err := k.GetCategoryTopics(ctx, req.CategoryId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	categoryResponse := types.CategoryResponse{
		Id:     category.Id,
		Name:   category.Name,
		Avatar: category.Avatar,
		Index:  category.Index,
	}
	var TopicResponseList []*types.TopicResponse
	for _, topicHash := range topics {
		topic, _ := k.GetTopic(ctx, topicHash)
		topicResponse := types.TopicResponse{
			Id:      topic.Id,
			Name:    topic.Name,
			Avatar:  topic.Avatar,
			Title:   topic.Title,
			Summary: topic.Summary,
			Score:   topic.Score,
			Hotness: topic.Hotness,
		}
		TopicResponseList = append(TopicResponseList, &topicResponse)
	}
	categoryTopicResponse := types.CategoryTopicResponse{
		Category: &categoryResponse,
		Topics:   TopicResponseList,
	}

	return &types.QueryTopicsByCategoryResponse{
		Response: &categoryTopicResponse,
	}, nil
}

// QueryCategoryPosts implements types.QueryServer.
func (k Querier) QueryCategoryPosts(goCtx context.Context, req *types.QueryCategoryPostsRequest) (*types.QueryCategoryPostsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	category := k.GetCategory(ctx, req.CategoryId)
	posts, _, err := k.Keeper.GetCategoryPosts(ctx, req.CategoryId)
	if err != nil {
		return nil, fmt.Errorf("failed to get posts by category %s: %w", req.CategoryId, err)
	}
	categoryResponse := types.CategoryResponse{
		Id:     category.Id,
		Name:   category.Name,
		Avatar: category.Avatar,
	}
	var postResponseList []*types.PostResponse
	for _, postId := range posts {
		post, _ := k.GetPost(ctx, postId)
		profile, _ := k.ProfileKeeper.GetProfile(ctx, post.Creator)
		postResponse := types.PostResponse{
			Post:    &post,
			Profile: &profile,
		}
		if post.Quote != "" {
			quotePost, _ := k.GetPost(ctx, post.Quote)
			quoteProfile, _ := k.ProfileKeeper.GetProfile(ctx, quotePost.Creator)
			postResponse.QuotePost = &quotePost
			postResponse.QuoteProfile = &quoteProfile
		}
		postResponseList = append(postResponseList, &postResponse)
	}
	categoryPostsResponse := types.CategoryPostsResponse{
		Category: &categoryResponse,
		Posts:    postResponseList,
	}

	return &types.QueryCategoryPostsResponse{
		Response: &categoryPostsResponse,
	}, nil
}

// QueryHotTopics72 implements types.QueryServer.
func (k Querier) QueryHotTopics72(goCtx context.Context, req *types.QueryHotTopics72Request) (*types.QueryHotTopics72Response, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	topics72, _, err := k.GetHotTopics72(ctx)
	if err != nil {
		return &types.QueryHotTopics72Response{}, nil
	}
	var topicResponseList []*types.TopicResponse
	for _, topicHash := range topics72 {
		topic, _ := k.GetTopic(ctx, topicHash)
		topicResponse := types.TopicResponse{
			Id:      topic.Id,
			Name:    topic.Name,
			Avatar:  topic.Avatar,
			Title:   topic.Title,
			Summary: topic.Summary,
			Score:   topic.Score,
			Hotness: topic.Hotness,
		}
		topicResponseList = append(topicResponseList, &topicResponse)
	}
	return &types.QueryHotTopics72Response{
		Topics: topicResponseList,
	}, nil
}
