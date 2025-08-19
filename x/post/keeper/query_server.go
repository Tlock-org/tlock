package keeper

import (
	"context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	profilekeeper "github.com/rollchains/tlock/x/profile/keeper"
	"sort"
	"strings"
	"time"

	"github.com/rollchains/tlock/x/post/types"
	profileTypes "github.com/rollchains/tlock/x/profile/types"
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
		types.LogError(k.logger, "get_params", err)
		return nil, types.ToGRPCError(types.WrapError(types.ErrDatabaseOperation, "failed to get params"))
	}

	return &types.QueryParamsResponse{Params: &p}, nil
}

// ResolveName implements types.QueryServer.
func (k Querier) ResolveName(goCtx context.Context, req *types.QueryResolveNameRequest) (*types.QueryResolveNameResponse, error) {
	v, err := k.Keeper.NameMapping.Get(goCtx, req.Address)
	if err != nil {
		types.LogError(k.logger, "get_name_mapping", err, "address", req.Address)
		return nil, types.ToGRPCError(types.WrapError(types.ErrDatabaseOperation, "failed to get name mapping"))
	}
	return &types.QueryResolveNameResponse{
		Name: v,
	}, nil
}

// BatchGetPostsWithProfiles retrieves posts and associated profiles in batch to optimize performance
func (k Querier) batchGetPostsWithProfiles(ctx sdk.Context, postIDs []string) ([]*types.PostResponse, error) {
	// Step 1: Batch retrieve all main posts
	posts := make(map[string]types.Post)
	quotePosts := make(map[string]types.Post)

	// Collect all post IDs that need to be queried (including quoted posts)
	allPostIDs := make(map[string]bool)
	for _, postID := range postIDs {
		allPostIDs[postID] = true
	}

	// Batch get main posts
	for _, postID := range postIDs {
		if post, found := k.GetPost(ctx, postID); found {
			posts[postID] = post
			// If there's a quoted post, add it to the query list
			if post.Quote != "" {
				allPostIDs[post.Quote] = true
			}
		}
	}

	// Batch get quoted posts
	for postID := range allPostIDs {
		if _, exists := posts[postID]; !exists {
			if post, found := k.GetPost(ctx, postID); found {
				quotePosts[postID] = post
			}
		}
	}

	// Step 2: Collect all unique creator addresses
	creators := make(map[string]bool)
	for _, post := range posts {
		creators[post.Creator] = true
	}
	for _, post := range quotePosts {
		creators[post.Creator] = true
	}

	// Step 3: Batch retrieve user profiles
	profiles := make(map[string]profileTypes.Profile)
	for creator := range creators {
		if profile, found := k.ProfileKeeper.GetProfile(ctx, creator); found {
			profiles[creator] = profile
		}
	}

	// Step 4: Assemble responses
	var responses []*types.PostResponse
	for _, postID := range postIDs {
		post, exists := posts[postID]
		if !exists {
			types.LogError(k.logger, "batchGetPostsWithProfiles", types.ErrPostNotFound, "post_id", postID)
			continue
		}

		profile, hasProfile := profiles[post.Creator]
		if !hasProfile {
			types.LogError(k.logger, "profile_not_found", types.ErrProfileNotFound, "creator", post.Creator, "post_id", postID)
			// Create default profile or skip
			profile = profileTypes.Profile{WalletAddress: post.Creator}
		}

		postResponse := &types.PostResponse{
			Post:    &post,
			Profile: &profile,
		}

		// Handle quoted posts
		if post.Quote != "" {
			if quotePost, hasQuote := quotePosts[post.Quote]; hasQuote {
				postResponse.QuotePost = &quotePost
				if quoteProfile, hasQuoteProfile := profiles[quotePost.Creator]; hasQuoteProfile {
					postResponse.QuoteProfile = &quoteProfile
				}
			}
		}

		responses = append(responses, postResponse)
	}

	return responses, nil
}

// QueryHomePosts implements types.QueryServer.
func (k Querier) QueryHomePosts(goCtx context.Context, req *types.QueryHomePostsRequest) (*types.QueryHomePostsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get list of post IDs
	postIDs, _, page, err := k.Keeper.GetHomePosts(ctx, req)
	if err != nil {
		return nil, types.ToGRPCError(err)
	}

	// Use batch query to optimize performance
	postResponses, err := k.batchGetPostsWithProfiles(ctx, postIDs)
	if err != nil {
		types.LogError(k.logger, "batchGetPostsWithProfiles", err, "operation", "QueryHomePosts")
		return nil, types.ToGRPCError(types.ErrDatabaseOperation)
	}

	return &types.QueryHomePostsResponse{
		Page:  page,
		Posts: postResponses,
	}, nil

}

// QueryFirstPageHomePosts implements types.QueryServer.
func (k Querier) QueryFirstPageHomePosts(goCtx context.Context, req *types.QueryFirstPageHomePostsRequest) (*types.QueryFirstPageHomePostsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get post IDs for the requested page
	postIDs, _, page, err := k.Keeper.GetFirstPageHomePosts(ctx, req)
	if err != nil {
		return nil, types.ToGRPCError(err)
	}

	// Use batch query for better performance
	postResponses, err := k.batchGetPostsWithProfiles(ctx, postIDs)
	if err != nil {
		types.LogError(k.logger, "batchGetPostsWithProfiles", err, "operation", "QueryFirstPageHomePosts")
		return nil, types.ToGRPCError(types.ErrDatabaseOperation)
	}

	return &types.QueryFirstPageHomePostsResponse{
		Page:  page,
		Posts: postResponses,
	}, nil
}

// QueryTopicPosts implements types.QueryServer.
func (k Querier) QueryTopicPosts(goCtx context.Context, req *types.QueryTopicPostsRequest) (*types.QueryTopicPostsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Retrieve posts for the specified topic
	postIDs, _, page, err := k.Keeper.GetTopicPosts(ctx, req.TopicId, req.Page)
	if err != nil {
		return nil, types.ToGRPCError(err)
	}

	// Apply batch optimization for posts and profiles
	postResponses, err := k.batchGetPostsWithProfiles(ctx, postIDs)
	if err != nil {
		types.LogError(k.logger, "batchGetPostsWithProfiles", err, "operation", "QueryTopicPosts")
		return nil, types.ToGRPCError(types.ErrDatabaseOperation)
	}

	return &types.QueryTopicPostsResponse{
		Page:  page,
		Posts: postResponses,
	}, nil
}

// QueryUserCreatedPosts implements types.QueryServer.
func (k Querier) QueryUserCreatedPosts(goCtx context.Context, req *types.QueryUserCreatedPostsRequest) (*types.QueryUserCreatedPostsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get user's created posts
	postIDs, _, page, err := k.Keeper.GetUserCreatedPosts(ctx, req.Address, req.Page)
	if err != nil {
		types.LogError(k.logger, "get_user_created_posts", err, "address", req.Address)
		return nil, types.ToGRPCError(types.WrapError(types.ErrDatabaseOperation, "failed to get user created posts"))
	}

	// Use batch processing to reduce database queries
	postResponses, err := k.batchGetPostsWithProfiles(ctx, postIDs)
	if err != nil {
		types.LogError(k.logger, "batch_get_posts_with_profiles", err, "post_count", len(postIDs))
		return nil, types.ToGRPCError(types.WrapError(types.ErrDatabaseOperation, "failed to retrieve posts and profiles"))
	}

	return &types.QueryUserCreatedPostsResponse{
		Page:  page,
		Posts: postResponses,
	}, nil

}

// QueryPost implements types.QueryServer.
func (k Querier) QueryPost(goCtx context.Context, req *types.QueryPostRequest) (*types.QueryPostResponse, error) {
	if req == nil {
		return nil, types.ToGRPCError(types.ErrInvalidRequest)
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Retrieve the post from the state
	post, found := k.Keeper.GetPost(sdk.UnwrapSDKContext(goCtx), req.PostId)
	if !found {
		return nil, types.ToGRPCError(types.NewPostNotFoundError(req.PostId))
	}
	postCopy := post

	profile, _ := k.ProfileKeeper.GetProfile(ctx, post.Creator)
	profileResponseCopy := profile
	postResponse := types.PostResponse{
		Post:    &postCopy,
		Profile: &profileResponseCopy,
	}

	topics := k.Keeper.GetTopicsByPostId(ctx, post.Id)
	var topicResponses []*types.TopicResponse
	if len(topics) > 0 {
		for _, topicHash := range topics {
			topic, _ := k.Keeper.GetTopic(ctx, topicHash)
			topicResponse := types.TopicResponse{
				Id:                    topic.Id,
				Name:                  topic.Name,
				Image:                 topic.Image,
				Title:                 topic.Title,
				Summary:               topic.Summary,
				Score:                 topic.Score,
				TrendingKeywordsScore: topic.TrendingKeywordsScore,
				CategoryId:            topic.CategoryId,
				Creator:               topic.Creator,
			}
			topicResponses = append(topicResponses, &topicResponse)
		}
	}

	if post.Quote != "" {
		quotePost, _ := k.GetPost(ctx, post.Quote)
		quoteProfile, _ := k.ProfileKeeper.GetProfile(ctx, quotePost.Creator)
		postResponse.QuotePost = &quotePost
		postResponse.QuoteProfile = &quoteProfile
	}
	return &types.QueryPostResponse{
		Post:   &postResponse,
		Topics: topicResponses,
	}, nil
}

// SearchTopic implements types.QueryServer.
func (k Querier) SearchTopics(goCtx context.Context, req *types.SearchTopicsRequest) (*types.SearchTopicsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	topicOriginalList, _ := k.SearchTopicMatches(ctx, req.Matching)
	var topicResponses []*types.TopicResponse
	for _, topicOriginal := range topicOriginalList {
		topicHash := k.sha256Generate(strings.ToLower(topicOriginal))
		topic, _ := k.GetTopic(ctx, topicHash)
		if topic.Id != "" {
			topicResponse := types.TopicResponse{
				Id:                    topic.Id,
				Name:                  topic.Name,
				Image:                 topic.Image,
				Title:                 topic.Title,
				Summary:               topic.Summary,
				Score:                 topic.Score,
				TrendingKeywordsScore: topic.TrendingKeywordsScore,
				CategoryId:            topic.CategoryId,
				Creator:               topic.Creator,
			}
			topicResponses = append(topicResponses, &topicResponse)
		}
	}
	return &types.SearchTopicsResponse{
		Topics: topicResponses,
	}, nil
}

// likesIMade implements types.QueryServer.
func (k Querier) LikesIMade(ctx context.Context, request *types.LikesIMadeRequest) (*types.LikesIMadeResponse, error) {
	if request == nil {
		return nil, types.ToGRPCError(types.ErrInvalidRequest)
	}
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	likes, _, page, err := k.GetLikesIMade(sdkCtx, request.Address, request.Page)
	if err != nil {
		types.LogError(k.logger, "get_likes_i_made", err, "address", request.Address)
		return nil, types.ToGRPCError(types.WrapError(types.ErrDatabaseOperation, "failed to get likes made"))
	}

	var postResponses []*types.PostResponse
	for _, likesIMade := range likes {
		postID := likesIMade.PostId
		post, success := k.GetPost(sdkCtx, postID)
		if !success {
			types.LogError(k.logger, "get_post_for_likes", types.ErrPostNotFound, "post_id", postID)
			return nil, types.ToGRPCError(types.NewPostNotFoundError(postID))
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
		return nil, types.ToGRPCError(types.ErrInvalidRequest)
	}
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	saves, _, page, err := k.GetSavesIMade(sdkCtx, request.Address, request.Page)
	if err != nil {
		types.LogError(k.logger, "get_saves_i_made", err, "address", request.Address)
		return nil, types.ToGRPCError(types.WrapError(types.ErrDatabaseOperation, "failed to get saves made"))
	}

	var postResponses []*types.PostResponse
	for _, savesIMade := range saves {
		postID := savesIMade.PostId
		post, success := k.GetPost(sdkCtx, postID)
		if !success {
			types.LogError(k.logger, "get_post_for_saves", types.ErrPostNotFound, "post_id", postID)
			return nil, types.ToGRPCError(types.NewPostNotFoundError(postID))
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
		return nil, types.ToGRPCError(types.ErrInvalidRequest)
	}
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	likesReceived, _, page, err := k.GetLikesReceived(sdkCtx, request.Address, request.Page)
	if err != nil {
		types.LogError(k.logger, "get_likes_received", err, "address", request.Address)
		return nil, types.ToGRPCError(types.WrapError(types.ErrDatabaseOperation, "failed to get likes received"))
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
			types.LogError(k.logger, "get_comment_for_query", types.ErrPostNotFound, "comment_id", commentId)
			return nil, types.ToGRPCError(types.NewPostNotFoundError(commentId))
		}
		commentCopy := comment
		profile, _ := k.ProfileKeeper.GetProfile(ctx, comment.Creator)
		pid := comment.ParentId
		parent, b := k.GetPost(ctx, pid)
		if !b {
			types.LogError(k.logger, "get_parent_post_for_comment", types.ErrPostNotFound, "parent_id", pid)
			return nil, types.ToGRPCError(types.NewPostNotFoundError(pid))
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
		types.LogError(k.logger, "get_comments_received", err, "address", req.Address)
		return nil, types.ToGRPCError(types.WrapError(types.ErrDatabaseOperation, "failed to get comments received"))
	}

	var commentReceivedResponses []*types.CommentReceivedResponse
	for _, commentID := range ids {
		comment, success := k.GetPost(ctx, commentID)
		if !success {
			types.LogError(k.logger, "get_comment_received", types.ErrPostNotFound, "comment_id", commentID)
			return nil, types.ToGRPCError(types.NewPostNotFoundError(commentID))
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
		types.LogError(k.logger, "get_category_topics", err, "category_id", req.CategoryId)
		return nil, types.ToGRPCError(types.WrapError(types.ErrDatabaseOperation, "failed to get category topics"))
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
			Id:                    topic.Id,
			Name:                  topic.Name,
			Image:                 topic.Image,
			Title:                 topic.Title,
			Summary:               topic.Summary,
			Score:                 topic.Score,
			TrendingKeywordsScore: topic.TrendingKeywordsScore,
			CategoryId:            topic.CategoryId,
			Creator:               topic.Creator,
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
		types.LogError(k.logger, "get_category_posts", err, "category_id", req.CategoryId)
		return nil, types.ToGRPCError(types.WrapError(types.ErrDatabaseOperation, "failed to get posts by category"))
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
	// sort followed posts by timestamp
	sort.Slice(postResponses, func(i, j int) bool {
		if postResponses[i].Post == nil && postResponses[j].Post == nil {
			return false
		}
		if postResponses[i].Post == nil {
			return false
		}
		if postResponses[j].Post == nil {
			return true
		}
		return postResponses[i].Post.Timestamp > postResponses[j].Post.Timestamp
	})
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
				Id:                    topic.Id,
				Name:                  topic.Name,
				Image:                 topic.Image,
				Title:                 topic.Title,
				Summary:               topic.Summary,
				Score:                 topic.Score,
				TrendingKeywordsScore: topic.TrendingKeywordsScore,
				CategoryId:            topic.CategoryId,
				Creator:               topic.Creator,
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
				Id:                    topic.Id,
				Name:                  topic.Name,
				Image:                 topic.Image,
				Title:                 topic.Title,
				Summary:               topic.Summary,
				Score:                 topic.Score,
				TrendingKeywordsScore: topic.TrendingKeywordsScore,
				CategoryId:            topic.CategoryId,
				Creator:               topic.Creator,
			}
			topicResponseList = append(topicResponseList, &topicResponse)
		}
		return &types.QueryUncategorizedTopicsResponse{
			Total:  total,
			Topics: topicResponseList,
		}, nil
	} else {
		return nil, types.ToGRPCError(types.ErrRequestDenied)
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

// QueryTopicAvatar implements types.QueryServer.
func (k Querier) QueryTopicImage(goCtx context.Context, req *types.QueryTopicImageRequest) (*types.QueryTopicImageResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	image := k.GetImageByTopic(ctx, req.TopicId)
	return &types.QueryTopicImageResponse{
		Image: image,
	}, nil
}

// QueryTopic implements types.QueryServer.
func (k Querier) QueryTopic(goCtx context.Context, req *types.QueryTopicRequest) (*types.QueryTopicResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	topic, _ := k.GetTopic(ctx, req.TopicId)
	image := k.GetImageByTopic(ctx, req.TopicId)
	topicResponse := types.TopicResponse{
		Id:                    topic.Id,
		Name:                  topic.Name,
		Image:                 topic.Image,
		Title:                 topic.Title,
		Summary:               topic.Summary,
		Score:                 topic.Score,
		TrendingKeywordsScore: topic.TrendingKeywordsScore,
		CategoryId:            topic.CategoryId,
		Creator:               topic.Creator,
	}
	if image != "" {
		topicResponse.Image = image
	}
	return &types.QueryTopicResponse{
		Topic: &topicResponse,
	}, nil
}

// QueryTrendingKeywords implements types.QueryServer.
func (k Querier) QueryTrendingKeywords(goCtx context.Context, req *types.QueryTrendingKeywordsRequest) (*types.QueryTrendingKeywordsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	keywords, _, page, err := k.GetTrendingKeywords(ctx, req.Page)
	if err != nil {
		return &types.QueryTrendingKeywordsResponse{}, nil
	}
	var topicResponseList []*types.TopicResponse
	for _, topicHash := range keywords {
		topic, _ := k.GetTopic(ctx, topicHash)
		topicResponse := types.TopicResponse{
			Id:                    topic.Id,
			Name:                  topic.Name,
			Image:                 topic.Image,
			Title:                 topic.Title,
			Summary:               topic.Summary,
			Score:                 topic.Score,
			TrendingKeywordsScore: topic.TrendingKeywordsScore,
			CategoryId:            topic.CategoryId,
			Creator:               topic.Creator,
		}
		topicResponseList = append(topicResponseList, &topicResponse)
	}
	return &types.QueryTrendingKeywordsResponse{
		Page:   page,
		Topics: topicResponseList,
	}, nil
}

// QueryTrendingTopics implements types.QueryServer.
func (k Querier) QueryTrendingTopics(goCtx context.Context, req *types.QueryTrendingTopicsRequest) (*types.QueryTrendingTopicsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	topics, _, page, err := k.GetTrendingTopics(ctx, req.Page)
	if err != nil {
		return &types.QueryTrendingTopicsResponse{}, nil
	}
	var topicResponseList []*types.TopicResponse
	for _, topicHash := range topics {
		topic, _ := k.GetTopic(ctx, topicHash)
		topicResponse := types.TopicResponse{
			Id:                    topic.Id,
			Name:                  topic.Name,
			Image:                 topic.Image,
			Title:                 topic.Title,
			Summary:               topic.Summary,
			Score:                 topic.Score,
			TrendingKeywordsScore: topic.TrendingKeywordsScore,
			CategoryId:            topic.CategoryId,
			Creator:               topic.Creator,
		}
		topicResponseList = append(topicResponseList, &topicResponse)
	}
	return &types.QueryTrendingTopicsResponse{
		Page:   page,
		Topics: topicResponseList,
	}, nil

}

// QueryPaidPostImage implements types.QueryServer.
func (k Querier) QueryPaidPostImage(goCtx context.Context, req *types.QueryPaidPostImageRequest) (*types.QueryPaidPostImageResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	image, found := k.GetPaidPostImage(ctx, req.ImageId)
	if !found {
		types.LogError(k.logger, "get_paid_post_image", types.ErrResourceNotFound, "image_id", req.ImageId)
		return nil, types.ToGRPCError(types.NewResourceNotFoundErrorf("no image found for image ID %s", req.ImageId))
	}
	return &types.QueryPaidPostImageResponse{
		Image: image,
	}, nil
}
