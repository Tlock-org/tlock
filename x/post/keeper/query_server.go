package keeper

import (
	"context"
	"cosmossdk.io/errors"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	profilekeeper "github.com/rollchains/tlock/x/profile/keeper"
	"google.golang.org/grpc/codes"
	"sort"
	"time"

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
	v, err := k.Keeper.NameMapping.Get(goCtx, req.Address)
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

	postIDs, _, page, err := k.Keeper.GetHomePosts(ctx, req)
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

		//profile, _ := k.ProfileKeeper.GetProfile(ctx, post.Creator)
		//
		//profileResponseCopy := profile

		postResponse := types.PostResponse{
			Post: &postCopy,
			//Profile: &profileResponseCopy,
		}
		if post.Quote != "" {
			quotePost, _ := k.GetPost(ctx, post.Quote)
			//quoteProfile, _ := k.ProfileKeeper.GetProfile(ctx, quotePost.Creator)
			postResponse.QuotePost = &quotePost
			//postResponse.QuoteProfile = &quoteProfile
		}

		postResponses = append(postResponses, &postResponse)
	}

	return &types.QueryHomePostsResponse{
		Page:  page,
		Posts: postResponses,
	}, nil
}

// QueryFirstPageHomePosts implements types.QueryServer.
func (k Querier) QueryFirstPageHomePosts(goCtx context.Context, req *types.QueryFirstPageHomePostsRequest) (*types.QueryFirstPageHomePostsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	page := req.Page
	postIDs, _, page, err := k.Keeper.GetFirstPageHomePosts(ctx, page)
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

		//profile, _ := k.ProfileKeeper.GetProfile(ctx, post.Creator)
		//profileResponseCopy := profile
		postResponse := types.PostResponse{
			Post: &postCopy,
			//Profile: &profileResponseCopy,
		}

		if post.Quote != "" {
			quotePost, _ := k.GetPost(ctx, post.Quote)
			//quoteProfile, _ := k.ProfileKeeper.GetProfile(ctx, quotePost.Creator)
			postResponse.QuotePost = &quotePost
			//postResponse.QuoteProfile = &quoteProfile
		}
		postResponses = append(postResponses, &postResponse)
	}
	return &types.QueryFirstPageHomePostsResponse{
		Page:  page,
		Posts: postResponses,
	}, nil
}

// QueryTopicPosts implements types.QueryServer.
func (k Querier) QueryTopicPosts(goCtx context.Context, req *types.QueryTopicPostsRequest) (*types.QueryTopicPostsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	topicHash := req.TopicId
	page := req.Page
	postIDs, _, page, err := k.Keeper.GetTopicPosts(ctx, topicHash, page)
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

		//profile, _ := k.ProfileKeeper.GetProfile(ctx, post.Creator)

		//profileResponseCopy := profile

		postResponse := types.PostResponse{
			Post: &postCopy,
			//Profile: &profileResponseCopy,
		}
		if post.Quote != "" {
			quotePost, _ := k.GetPost(ctx, post.Quote)
			//quoteProfile, _ := k.ProfileKeeper.GetProfile(ctx, quotePost.Creator)
			postResponse.QuotePost = &quotePost
			//postResponse.QuoteProfile = &quoteProfile
		}

		postResponses = append(postResponses, &postResponse)
	}
	return &types.QueryTopicPostsResponse{
		Page:  page,
		Posts: postResponses,
	}, nil
}

// QueryUserCreatedPosts implements types.QueryServer.
func (k Querier) QueryUserCreatedPosts(goCtx context.Context, req *types.QueryUserCreatedPostsRequest) (*types.QueryUserCreatedPostsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	postIDs, _, page, err := k.Keeper.GetUserCreatedPosts(ctx, req.Address, req.Page)
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

		//profile, _ := k.ProfileKeeper.GetProfile(ctx, post.Creator)
		//profileResponseCopy := profile
		postResponse := types.PostResponse{
			Post: &postCopy,
			//Profile: &profileResponseCopy,
		}

		if post.Quote != "" {
			quotePost, _ := k.GetPost(ctx, post.Quote)
			//quoteProfile, _ := k.ProfileKeeper.GetProfile(ctx, quotePost.Creator)
			postResponse.QuotePost = &quotePost
			//postResponse.QuoteProfile = &quoteProfile
		}
		postResponses = append(postResponses, &postResponse)
	}
	return &types.QueryUserCreatedPostsResponse{
		Page:  page,
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

	//profile, _ := k.ProfileKeeper.GetProfile(ctx, post.Creator)
	//profileResponseCopy := profile
	postResponse := types.PostResponse{
		Post: &postCopy,
		//Profile: &profileResponseCopy,
	}

	if post.Quote != "" {
		quotePost, _ := k.GetPost(ctx, post.Quote)
		//quoteProfile, _ := k.ProfileKeeper.GetProfile(ctx, quotePost.Creator)
		postResponse.QuotePost = &quotePost
		//postResponse.QuoteProfile = &quoteProfile
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

	likes, _, page, err := k.GetLikesIMade(sdkCtx, request.Address, request.Page)
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

		profile, _ := k.ProfileKeeper.GetProfile(sdkCtx, post.Creator)

		profileResponseCopy := profile
		postResponse := types.PostResponse{
			Post:    &postCopy,
			Profile: &profileResponseCopy,
		}

		postResponses = append(postResponses, &postResponse)
	}
	return &types.LikesIMadeResponse{
		Page:  page,
		Posts: postResponses,
	}, nil
}

// SavesIMade implements types.QueryServer.
func (k Querier) SavesIMade(ctx context.Context, request *types.SavesIMadeRequest) (*types.SavesIMadeResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	saves, _, page, err := k.GetSavesIMade(sdkCtx, request.Address, request.Page)
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

		profile, _ := k.ProfileKeeper.GetProfile(sdkCtx, post.Creator)

		profileResponseCopy := profile
		postResponse := types.PostResponse{
			Post:    &postCopy,
			Profile: &profileResponseCopy,
		}

		postResponses = append(postResponses, &postResponse)
	}

	return &types.SavesIMadeResponse{
		Page:  page,
		Posts: postResponses,
	}, nil
}

// LikesReceived implements types.QueryServer.
func (k Querier) LikesReceived(ctx context.Context, request *types.LikesReceivedRequest) (*types.LikesReceivedResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	likesReceived, _, page, err := k.GetLikesReceived(sdkCtx, request.Address, request.Page)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.LikesReceivedResponse{
		Page:          page,
		LikesReceived: likesReceived,
	}, nil
}

// QueryComments implements types.QueryServer.
func (k Querier) QueryComments(goCtx context.Context, req *types.QueryCommentsRequest) (*types.QueryCommentsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	ids, _, page, _ := k.Keeper.GetCommentsByParentId(ctx, req.Id, req.Page)

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
		Page:     page,
		Comments: commentResponses,
	}, nil
}

// QueryCommentsReceived implements types.QueryServer.
func (k Querier) QueryCommentsReceived(goCtx context.Context, req *types.QueryCommentsReceivedRequest) (*types.QueryCommentsReceivedResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	ids, _, page, err := k.Keeper.GetCommentsReceived(ctx, req.Address, req.Page)
	if err != nil {
		return nil, fmt.Errorf("failed to get comments received %s: %w", req.Address, err)
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
		Page:     page,
		Comments: commentReceivedResponses,
	}, nil
}

// QueryActivitiesReceived implements types.QueryServer.
func (k Querier) QueryActivitiesReceived(goCtx context.Context, req *types.QueryActivitiesReceivedRequest) (*types.QueryActivitiesReceivedResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	list, _, page, _ := k.ProfileKeeper.GetActivitiesReceived(ctx, req.Address, req.Page)
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
		Page:               page,
		ActivitiesReceived: activitiesReceivedList,
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
	topics, _, page, err := k.GetCategoryTopics(ctx, req.CategoryId, req.Page)
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
		}
		TopicResponseList = append(TopicResponseList, &topicResponse)
	}
	categoryTopicResponse := types.CategoryTopicResponse{
		Category: &categoryResponse,
		Topics:   TopicResponseList,
	}

	return &types.QueryTopicsByCategoryResponse{
		Page:     page,
		Response: &categoryTopicResponse,
	}, nil
}

// QueryCategoryByTopic implements types.QueryServer.
func (k Querier) QueryCategoryByTopic(goCtx context.Context, req *types.QueryCategoryByTopicRequest) (*types.QueryCategoryByTopicResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	topicHash := req.TopicId
	categoryHash := k.getCategoryByTopicHash(ctx, topicHash)
	if categoryHash != "" {
		//categoryHash := k.sha256Generate(categoryDb)
		category := k.GetCategory(ctx, categoryHash)
		k.Logger().Error("==========category:", "category", category)
		categoryResponse := types.CategoryResponse{
			Id:     categoryHash,
			Name:   category.Name,
			Avatar: category.Avatar,
			Index:  category.Index,
		}
		return &types.QueryCategoryByTopicResponse{
			Category: &categoryResponse,
		}, nil
	} else {
		return &types.QueryCategoryByTopicResponse{}, nil
	}
}

// QueryCategoryPosts implements types.QueryServer.
func (k Querier) QueryCategoryPosts(goCtx context.Context, req *types.QueryCategoryPostsRequest) (*types.QueryCategoryPostsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	category := k.GetCategory(ctx, req.CategoryId)
	posts, _, page, err := k.Keeper.GetCategoryPosts(ctx, req.CategoryId, req.Page)
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
		Page:     page,
		Response: &categoryPostsResponse,
	}, nil
}

// QueryHotTopics72 implements types.QueryServer.
func (k Querier) QueryHotTopics72(goCtx context.Context, req *types.QueryHotTopics72Request) (*types.QueryHotTopics72Response, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	topics72, _, page, err := k.GetHotTopics72(ctx, req.Page)
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
		}
		topicResponseList = append(topicResponseList, &topicResponse)
	}
	return &types.QueryHotTopics72Response{
		Page:   page,
		Topics: topicResponseList,
	}, nil
}

// QueryFollowingPosts implements types.QueryServer.
func (k Querier) QueryFollowingPosts(goCtx context.Context, req *types.QueryFollowingPostsRequest) (*types.QueryFollowingPostsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	address := req.Address
	followingList := k.ProfileKeeper.GetFollowing(ctx, address)
	followingCount := len(followingList)
	var postResponses []*types.PostResponse
	if followingCount > 0 {
		var postIds []string
		if followingCount <= 5 {
			for _, follow := range followingList {
				ids, _, _ := k.GetLastPostsByAddress(ctx, follow, 5)
				postIds = append(postIds, ids...)
			}
		} else if followingCount <= 50 {
			y := 50 / followingCount
			if y < 1 {
				y = 1
			}
			for _, follow := range followingList {
				ids, _, _ := k.GetLastPostsByAddress(ctx, follow, y)
				postIds = append(postIds, ids...)
			}

		} else {
			selectedUsers := select50Users(followingList)

			for _, follow := range selectedUsers {
				ids, _, _ := k.GetLastPostsByAddress(ctx, follow, 1)
				if len(ids) > 0 {
					postIds = append(postIds, ids...)
				}
			}
		}
		for _, id := range postIds {
			post, _ := k.GetPost(ctx, id)
			postTime := post.Timestamp
			if isToday(ctx, postTime) {
				profile, _ := k.ProfileKeeper.GetProfile(ctx, post.Creator)
				response := types.PostResponse{
					Post:    &post,
					Profile: &profile,
				}
				if post.Quote != "" {
					quotePost, _ := k.GetPost(ctx, post.Quote)
					quoteProfile, _ := k.ProfileKeeper.GetProfile(ctx, quotePost.Creator)
					response.QuotePost = &quotePost
					response.QuoteProfile = &quoteProfile
				}
				postResponses = append(postResponses, &response)
			}
		}

	} else {

	}
	return &types.QueryFollowingPostsResponse{
		Posts: postResponses,
	}, nil
}

func isToday(ctx sdk.Context, postTimestamp int64) bool {
	postTime := time.Unix(postTimestamp, 0).UTC()
	blockTime := ctx.BlockTime().UTC()
	return blockTime.Year() == postTime.Year() &&
		blockTime.Month() == postTime.Month() &&
		blockTime.Day() == postTime.Day()
}

func select50Users(followingList []string) []string {
	n := len(followingList)
	limit := 50
	if n < 50 {
		limit = n
	}
	sortedList := make([]string, len(followingList))
	copy(sortedList, followingList)
	sort.Strings(sortedList)
	return sortedList[:limit]
}

// QueryFollowingTopics implements types.QueryServer.
func (k Querier) QueryFollowingTopics(goCtx context.Context, req *types.QueryFollowingTopicsRequest) (*types.QueryFollowingTopicsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	topicIds, _, page, _ := k.GetFollowingTopics(ctx, req.Address, req.Page)
	var topicResponseList []*types.TopicResponse
	if len(topicIds) > 0 {
		for _, topicHash := range topicIds {
			topic, _ := k.GetTopic(ctx, topicHash)
			topicResponse := types.TopicResponse{
				Id:      topic.Id,
				Name:    topic.Name,
				Avatar:  topic.Avatar,
				Title:   topic.Title,
				Summary: topic.Summary,
				Score:   topic.Score,
			}
			topicResponseList = append(topicResponseList, &topicResponse)
		}
	}
	return &types.QueryFollowingTopicsResponse{
		Page:   page,
		Topics: topicResponseList,
	}, nil
}

// QueryIsFollowingTopic implements types.QueryServer.
func (k Querier) QueryIsFollowingTopic(goCtx context.Context, req *types.QueryIsFollowingTopicRequest) (*types.QueryIsFollowingTopicResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	isFollowingTopic := k.IsFollowingTopic(ctx, req.Address, req.TopicId)
	return &types.QueryIsFollowingTopicResponse{
		IsFollowing: isFollowingTopic,
	}, nil
}

// QueryUncategorizedTopics implements types.QueryServer.
func (k Querier) QueryUncategorizedTopics(goCtx context.Context, req *types.QueryUncategorizedTopicsRequest) (*types.QueryUncategorizedTopicsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	address := req.Address
	profile, _ := k.ProfileKeeper.GetProfile(ctx, address)
	if profile.AdminLevel > 0 {
		topics, response, _ := k.GetUncategorizedTopics(ctx, req.Page, req.Limit)
		total := response.Total
		var topicResponseList []*types.TopicResponse
		for _, topicHash := range topics {
			topic, _ := k.GetTopic(ctx, topicHash)
			topicResponse := types.TopicResponse{
				Id:      topic.Id,
				Name:    topic.Name,
				Avatar:  topic.Avatar,
				Title:   topic.Title,
				Summary: topic.Summary,
				Score:   topic.Score,
			}
			topicResponseList = append(topicResponseList, &topicResponse)
		}
		return &types.QueryUncategorizedTopicsResponse{
			Total:  total,
			Topics: topicResponseList,
		}, nil
	} else {
		return nil, errors.Wrapf(types.ErrRequestDenied, "request denied")
	}

}

// QueryVoteOption implements types.QueryServer.
func (k Querier) QueryVoteOption(goCtx context.Context, req *types.QueryVoteOptionRequest) (*types.QueryVoteOptionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	optionId, _ := k.GetPoll(ctx, req.PostId, req.Address)
	return &types.QueryVoteOptionResponse{
		OptionId: optionId,
	}, nil
}
