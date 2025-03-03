package keeper

import (
	"context"
	"cosmossdk.io/errors"
	"encoding/base64"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/rollchains/tlock/x/post/types"
	profiletypes "github.com/rollchains/tlock/x/profile/types"
	"math"
)

const MaxImageSize = 500 * 1024 // 500 KB

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

// SetServiceName implements types.MsgServer.
func (ms msgServer) SetServiceName(ctx context.Context, msg *types.MsgSetServiceName) (*types.MsgSetServiceNameResponse, error) {
	// ctx := sdk.UnwrapSDKContext(goCtx)
	//panic("SetServiceName is unimplemented")
	//return &types.MsgSetServiceNameResponse{}, nil
	if err := ms.k.NameMapping.Set(ctx, msg.Sender, msg.Name); err != nil {
		return nil, err
	}

	return &types.MsgSetServiceNameResponse{}, nil
}

// SetApprove implements types.MsgServer.
func (ms msgServer) GrantAllowanceFromModule(goCtx context.Context, msg *types.MsgGrantAllowanceFromModuleRequest) (*types.MsgGrantAllowanceFromModuleResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, senderErr := sdk.AccAddressFromBech32(msg.Sender)
	fmt.Printf("====sender: %s\n", sender)
	if senderErr != nil {
		return &types.MsgGrantAllowanceFromModuleResponse{Status: false}, errors.Wrapf(types.ErrInvalidAddress, "Invalid sender address: %s", senderErr)
	}
	userAddress, userErr := sdk.AccAddressFromBech32(msg.UserAddress)
	fmt.Printf("=====userAddress: %s\n", userAddress)
	if userErr != nil {
		return &types.MsgGrantAllowanceFromModuleResponse{Status: false}, errors.Wrapf(types.ErrInvalidAddress, "Invalid sender address: %s", userErr)
	}

	ms.k.GrantPeriodicAllowance(ctx, sender, userAddress)

	return &types.MsgGrantAllowanceFromModuleResponse{Status: true}, nil
}

// CreateFreePostWithTitle implements types.MsgServer.
func (ms msgServer) CreateFreePostWithTitle(goCtx context.Context, msg *types.MsgCreateFreePostWithTitle) (*types.MsgCreateFreePostWithTitleResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate the message
	if len(msg.Title) == 0 {
		return nil, errors.Wrapf(types.ErrInvalidRequest, "Title cannot be empty")
	}
	if len(msg.Content) == 0 {
		return nil, errors.Wrapf(types.ErrInvalidRequest, "Content cannot be empty")
	}
	if err := types.ValidatePostWithTitleContent(msg.Content); err != nil {
		return nil, errors.Wrap(types.ErrInvalidRequest, err.Error())
	}

	// Validate sender address
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errors.Wrapf(types.ErrInvalidAddress, "Invalid sender address: %s", err)
	}

	//if len(msg.Image) > MaxImageSize {
	//	return nil, errors.Wrap(types.ErrInvalidRequest, "Image size exceeds the maximum allowed limit")
	//}

	blockTime := ctx.BlockTime().Unix()
	// Generate a unique post ID
	data := fmt.Sprintf("%s|%s|%s|%d", msg.Creator, msg.Title, msg.Content, blockTime)
	postId := ms.k.sha256Generate(data)

	// Create the post
	post := types.Post{
		Id:              postId,
		PostType:        types.PostType_ARTICLE,
		Title:           msg.Title,
		Content:         msg.Content,
		Creator:         msg.Creator,
		Timestamp:       blockTime,
		ImagesUrl:       msg.ImagesUrl,
		VideosUrl:       msg.VideosUrl,
		HomePostsUpdate: blockTime,
		Poll:            msg.Poll,
	}
	if msg.Poll != nil {
		post.PostType = types.PostType_POLL
	}

	// Store the post in the state
	ms.k.SetPost(ctx, post)
	// post reward
	ms.k.PostReward(ctx, post)
	// add home posts
	ms.addToHomePosts(ctx, post)
	// add to user created posts
	ms.addToUserCreatedPosts(ctx, msg.Creator, post)

	userHandleList := msg.Mention
	if len(userHandleList) > 0 {
		if len(userHandleList) > 10 {
			return nil, fmt.Errorf("cannot mention more than 10 users")
		}
		for _, userHandle := range userHandleList {
			address := ms.k.ProfileKeeper.GetAddressByUserHandle(ctx, userHandle)
			if address != "" {
				ms.addActivitiesReceived(ctx, post, "", "", msg.Creator, address, profiletypes.ActivitiesType_ACTIVITIES_MENTION)
			}
		}
	}

	topicList := msg.Topic
	category := msg.Category
	err = ms.handleCategoryTopicPost(ctx, topicList, category, blockTime, postId)
	if err != nil {
		return nil, err
	}

	//Emit an event for the creation
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeCreateFreePostWithTitle,
			sdk.NewAttribute(types.AttributeKeyCreator, msg.Creator),
			sdk.NewAttribute(types.AttributeKeyPostID, postId),
			sdk.NewAttribute(types.AttributeKeyTitle, msg.Title),
			sdk.NewAttribute(types.AttributeKeyTimestamp, fmt.Sprintf("%d", blockTime)),
		),
	})

	return &types.MsgCreateFreePostWithTitleResponse{PostId: postId}, nil
}

// CreateFreePost implements types.MsgServer.
func (ms msgServer) CreateFreePost(goCtx context.Context, msg *types.MsgCreateFreePost) (*types.MsgCreateFreePostResponse, error) {
	base64.StdEncoding.EncodeToString([]byte("{\"body\":{\"messages\":[{\"@type\":\"/cosmos.bank.v1beta1.MsgSend\",\"from_address\":\"tlock1hj5fveer5cjtn4wd6wstzugjfdxzl0xp5u7j9p\",\"to_address\":\"tlock1efd63aw40lxf3n4mhf7dzhjkr453axurggdkvg\",\"amount\":[{\"denom\":\"uTOK\",\"amount\":\"10\"}]}],\"memo\":\"\",\"timeout_height\":\"0\",\"extension_options\":[],\"non_critical_extension_options\":[]},\"auth_info\":{\"signer_infos\":[{\"public_key\":{\"@type\":\"/cosmos.crypto.secp256k1.PubKey\",\"key\":\"ApZa31BR3NWLylRT6Qi5+f+zXtj2OpqtC76vgkUGLyww\"},\"mode_info\":{\"single\":{\"mode\":\"SIGN_MODE_DIRECT\"}},\"sequence\":\"1\"}],\"fee\":{\"amount\":[{\"denom\":\"uTOK\",\"amount\":\"77886\"}],\"gas_limit\":\"77886\",\"payer\":\"\",\"granter\":\"\"},\"tip\":null},\"signatures\":[\"RWGJJcQ8Ioul52d6HbBW6F1FJwuNPRSTTny6xpAZCfpYgHSIGuk0+uupaC5gx0FKur8qOA9tZlhKhAfLyf9hWg==\"]}"))

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate the message
	if len(msg.Content) == 0 {
		return nil, errors.Wrapf(types.ErrInvalidRequest, "Content cannot be empty")
	}
	if err := types.ValidatePostContent(msg.Content); err != nil {
		return nil, errors.Wrap(types.ErrInvalidRequest, err.Error())
	}

	// Validate sender address
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errors.Wrapf(types.ErrInvalidAddress, "Invalid sender address: %s", err)
	}

	//if len(msg.Image) > MaxImageSize {
	//	return nil, errors.Wrap(types.ErrInvalidRequest, "Image size exceeds the maximum allowed limit")
	//}

	// Generate a unique post ID
	blockTime := ctx.BlockTime().Unix()
	data := fmt.Sprintf("%s|%s|%d", msg.Creator, msg.Content, blockTime)
	postId := ms.k.sha256Generate(data)

	// Create the post
	post := types.Post{
		Id:              postId,
		PostType:        types.PostType_ORIGINAL,
		Content:         msg.Content,
		Creator:         msg.Creator,
		Timestamp:       blockTime,
		ImagesUrl:       msg.ImagesUrl,
		VideosUrl:       msg.VideosUrl,
		HomePostsUpdate: blockTime,
		Poll:            msg.Poll,
	}
	if msg.Poll != nil {
		post.PostType = types.PostType_POLL
	}

	// Store the post in the state
	ms.k.SetPost(ctx, post)
	// post reward
	ms.k.PostReward(ctx, post)
	// add to home posts
	ms.addToHomePosts(ctx, post)
	// add to user created posts
	ms.addToUserCreatedPosts(ctx, msg.Creator, post)

	ms.k.ProfileKeeper.CheckAndCreateUserHandle(ctx, msg.Creator)

	// mentions add to activitiesReceived
	userHandleList := msg.Mention
	if len(userHandleList) > 0 {
		if len(userHandleList) > 10 {
			return nil, fmt.Errorf("cannot mention more than 10 users")
		}
		for _, userHandle := range userHandleList {
			address := ms.k.ProfileKeeper.GetAddressByUserHandle(ctx, userHandle)
			if address != "" {
				ms.addActivitiesReceived(ctx, post, "", "", msg.Creator, address, profiletypes.ActivitiesType_ACTIVITIES_MENTION)
			}
		}
	}

	topicList := msg.Topic
	category := msg.Category
	err = ms.handleCategoryTopicPost(ctx, topicList, category, blockTime, postId)
	if err != nil {
		return nil, err
	}

	//Emit an event for the creation
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeCreateFreePost,
			sdk.NewAttribute(types.AttributeKeyCreator, msg.Creator),
			sdk.NewAttribute(types.AttributeKeyPostID, postId),
			sdk.NewAttribute(types.AttributeKeyTimestamp, fmt.Sprintf("%d", blockTime)),
		),
	})
	//ms.k.Logger().Warn("============= create free post end ==============")
	return &types.MsgCreateFreePostResponse{PostId: postId}, nil
}

// CreateFreePostImagePayable implements types.MsgServer.
func (ms msgServer) CreateFreePostImagePayable(goCtx context.Context, msg *types.MsgCreateFreePostImagePayable) (*types.MsgCreateFreePostImagePayableResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate the message
	if len(msg.Content) == 0 {
		return nil, errors.Wrapf(types.ErrInvalidRequest, "Content cannot be empty")
	}
	if err := types.ValidatePostContent(msg.Content); err != nil {
		return nil, errors.Wrap(types.ErrInvalidRequest, err.Error())
	}

	// Validate sender address
	//sender, err := sdk.AccAddressFromBech32(msg.Sender)
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errors.Wrapf(types.ErrInvalidAddress, "Invalid sender address: %s", err)
	}

	//if len(msg.Image) > MaxImageSize {
	//	return nil, errors.Wrap(types.ErrInvalidRequest, "Image size exceeds the maximum allowed limit")
	//}

	blockTime := ctx.BlockTime().Unix()
	data := fmt.Sprintf("%s|%s|%d", msg.Creator, msg.Content, blockTime)
	postId := ms.k.sha256Generate(data)

	// Create the post
	post := types.Post{
		Id:              postId,
		PostType:        types.PostType_ADVERTISEMENT,
		Content:         msg.Content,
		Creator:         msg.Creator,
		Timestamp:       blockTime,
		ImagesBase64:    msg.ImagesBase64,
		ImagesUrl:       msg.ImagesUrl,
		VideosUrl:       msg.VideosUrl,
		HomePostsUpdate: blockTime,
	}

	// post payment
	ms.k.postPayment(ctx, post)
	// Store the post in the state
	ms.k.SetPost(ctx, post)
	// add home posts
	ms.addToHomePosts(ctx, post)
	// add to user created posts
	ms.addToUserCreatedPosts(ctx, msg.Creator, post)

	ms.k.ProfileKeeper.CheckAndCreateUserHandle(ctx, msg.Creator)

	// mentions add to activitiesReceived
	userHandleList := msg.Mention
	if len(userHandleList) > 0 {
		if len(userHandleList) > 10 {
			return nil, fmt.Errorf("cannot mention more than 10 users")
		}
		for _, userHandle := range userHandleList {
			address := ms.k.ProfileKeeper.GetAddressByUserHandle(ctx, userHandle)
			if address != "" {
				ms.addActivitiesReceived(ctx, post, "", "", msg.Creator, address, profiletypes.ActivitiesType_ACTIVITIES_MENTION)
			}
		}
	}

	topicList := msg.Topic
	category := msg.Category
	err = ms.handleCategoryTopicPost(ctx, topicList, category, blockTime, postId)
	if err != nil {
		return nil, err
	}

	//Emit an event for the creation
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeCreatePaidPost,
			sdk.NewAttribute(types.AttributeKeyCreator, msg.Creator),
			sdk.NewAttribute(types.AttributeKeyPostID, postId),
			sdk.NewAttribute(types.AttributeKeyTitle, msg.Title),
			sdk.NewAttribute(types.AttributeKeyTimestamp, fmt.Sprintf("%d", blockTime)),
		),
	})
	return &types.MsgCreateFreePostImagePayableResponse{}, nil
}

func (ms msgServer) CreatePaidPost(goCtx context.Context, msg *types.MsgCreatePaidPost) (*types.MsgCreatePaidPostResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate the message
	if len(msg.Content) == 0 {
		return nil, errors.Wrapf(types.ErrInvalidRequest, "Content cannot be empty")
	}
	if err := types.ValidatePostContent(msg.Content); err != nil {
		return nil, errors.Wrap(types.ErrInvalidRequest, err.Error())
	}

	// Validate sender address
	//sender, err := sdk.AccAddressFromBech32(msg.Sender)
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errors.Wrapf(types.ErrInvalidAddress, "Invalid sender address: %s", err)
	}

	//if len(msg.Image) > MaxImageSize {
	//	return nil, errors.Wrap(types.ErrInvalidRequest, "Image size exceeds the maximum allowed limit")
	//}

	blockTime := ctx.BlockTime().Unix()
	data := fmt.Sprintf("%s|%s|%d", msg.Creator, msg.Content, blockTime)
	postId := ms.k.sha256Generate(data)

	// Create the post
	post := types.Post{
		Id:              postId,
		PostType:        types.PostType_ADVERTISEMENT,
		Content:         msg.Content,
		Creator:         msg.Creator,
		Timestamp:       blockTime,
		ImagesBase64:    msg.ImagesBase64,
		ImagesUrl:       msg.ImagesUrl,
		VideosUrl:       msg.VideosUrl,
		HomePostsUpdate: blockTime,
	}

	// post payment
	ms.k.postPayment(ctx, post)
	// Store the post in the state
	ms.k.SetPost(ctx, post)
	// add home posts
	ms.addToHomePosts(ctx, post)
	// add to user created posts
	ms.addToUserCreatedPosts(ctx, msg.Creator, post)

	ms.k.ProfileKeeper.CheckAndCreateUserHandle(ctx, msg.Creator)

	// mentions add to activitiesReceived
	userHandleList := msg.Mention
	if len(userHandleList) > 0 {
		if len(userHandleList) > 10 {
			return nil, fmt.Errorf("cannot mention more than 10 users")
		}
		for _, userHandle := range userHandleList {
			address := ms.k.ProfileKeeper.GetAddressByUserHandle(ctx, userHandle)
			if address != "" {
				ms.addActivitiesReceived(ctx, post, "", "", msg.Creator, address, profiletypes.ActivitiesType_ACTIVITIES_MENTION)
			}
		}
	}

	topicList := msg.Topic
	category := msg.Category
	err = ms.handleCategoryTopicPost(ctx, topicList, category, blockTime, postId)
	if err != nil {
		return nil, err
	}

	//Emit an event for the creation
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeCreatePaidPost,
			sdk.NewAttribute(types.AttributeKeyCreator, msg.Creator),
			sdk.NewAttribute(types.AttributeKeyPostID, postId),
			sdk.NewAttribute(types.AttributeKeyTitle, msg.Title),
			sdk.NewAttribute(types.AttributeKeyTimestamp, fmt.Sprintf("%d", blockTime)),
		),
	})

	return &types.MsgCreatePaidPostResponse{PostId: postId}, nil
}

// QuotePost implements types.MsgServer.
func (ms msgServer) QuotePost(goCtx context.Context, msg *types.MsgQuotePostRequest) (*types.MsgQuotePostResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate the message
	if len(msg.Comment) == 0 {
		return nil, errors.Wrapf(types.ErrInvalidRequest, "Comment cannot be empty")
	}
	if len(msg.Quote) == 0 {
		return nil, errors.Wrapf(types.ErrInvalidRequest, "Quote cannot be empty")
	}
	if err := types.ValidatePostContent(msg.Comment); err != nil {
		return nil, errors.Wrap(types.ErrInvalidRequest, err.Error())
	}

	// Validate sender address
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errors.Wrapf(types.ErrInvalidAddress, "Invalid sender address: %s", err)
	}

	blockTime := ctx.BlockTime().Unix()
	data := fmt.Sprintf("%s|%s|%s|%d", msg.Creator, msg.Quote, msg.Comment, blockTime)
	postId := ms.k.sha256Generate(data)

	// Create the post
	post := types.Post{
		Id:              postId,
		PostType:        types.PostType_QUOTE,
		Content:         msg.Comment,
		Creator:         msg.Creator,
		Timestamp:       blockTime,
		Quote:           msg.Quote,
		HomePostsUpdate: blockTime,
	}

	// Store the post in the state
	ms.k.SetPost(ctx, post)
	// post reward
	ms.k.PostReward(ctx, post)
	// add home posts
	ms.addToHomePosts(ctx, post)
	// add to user created posts
	ms.addToUserCreatedPosts(ctx, msg.Creator, post)
	ms.k.ProfileKeeper.CheckAndCreateUserHandle(ctx, msg.Creator)

	parentPost, _ := ms.k.GetPost(ctx, msg.Quote)
	parentPost.LikeCount += 1
	ms.k.SetPost(ctx, parentPost)

	// mentions add to activitiesReceived
	userHandleList := msg.Mention
	if len(userHandleList) > 0 {
		if len(userHandleList) > 10 {
			return nil, fmt.Errorf("cannot mention more than 10 users")
		}
		for _, userHandle := range userHandleList {
			address := ms.k.ProfileKeeper.GetAddressByUserHandle(ctx, userHandle)
			if address != "" {
				ms.addActivitiesReceived(ctx, post, "", "", msg.Creator, address, profiletypes.ActivitiesType_ACTIVITIES_MENTION)
			}
		}
	}

	topicList := msg.Topic
	category := msg.Category
	err = ms.handleCategoryTopicPost(ctx, topicList, category, blockTime, postId)
	if err != nil {
		return nil, err
	}

	//Emit an event for the creation
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeCreateFreePostWithTitle,
			sdk.NewAttribute(types.AttributeKeyCreator, msg.Creator),
			sdk.NewAttribute(types.AttributeKeyPostID, postId),
			sdk.NewAttribute(types.AttributeKeyTimestamp, fmt.Sprintf("%d", blockTime)),
		),
	})

	return &types.MsgQuotePostResponse{}, nil
}

// Repost implements types.MsgServer.
func (ms msgServer) Repost(goCtx context.Context, msg *types.MsgRepostRequest) (*types.MsgRepostResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate the message
	if len(msg.Quote) == 0 {
		return nil, errors.Wrapf(types.ErrInvalidRequest, "Quote cannot be empty")
	}

	// Validate sender address
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errors.Wrapf(types.ErrInvalidAddress, "Invalid sender address: %s", err)
	}

	parentPost, _ := ms.k.GetPost(ctx, msg.Quote)
	// add home posts
	//ms.addToHomePosts(ctx, parentPost)
	// add to user created posts
	ms.addToUserCreatedPosts(ctx, msg.Creator, parentPost)
	ms.k.ProfileKeeper.CheckAndCreateUserHandle(ctx, msg.Creator)

	parentPost.RepostCount += 1
	ms.k.SetPost(ctx, parentPost)

	return &types.MsgRepostResponse{}, nil
}

// LikePost implements types.MsgServer.
func (ms msgServer) Like(goCtx context.Context, msg *types.MsgLikeRequest) (*types.MsgLikeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	hasLiked := ms.k.HasUserLikedPost(ctx, msg.Sender, msg.Id)
	if hasLiked {
		return nil, errors.Wrap(types.ErrAlreadyLiked, "user has already liked this post")
	}
	post, found := ms.k.GetPost(ctx, msg.Id)
	if !found {
		return nil, errors.Wrap(types.ErrPostNotFound, msg.Id)
	}

	oldScore := post.Score
	// Score Accumulation
	uintExponent := ms.ScoreAccumulation(ctx, msg.Sender, post, 3)
	if uintExponent >= 1 {
		newScore := oldScore + uintExponent
		post.Score = newScore
	}
	// update commentList
	if post.PostType == types.PostType_COMMENT {
		ms.k.DeleteFromCommentList(ctx, post.ParentId, post.Id, oldScore)
		ms.k.AddToCommentList(ctx, post.ParentId, post.Id, post.Score)
	}

	// add home posts
	if post.PostType != types.PostType_COMMENT {
		exist := ms.k.IsPostInHomePosts(ctx, post.Id, post.HomePostsUpdate)

		ms.k.Logger().Warn("============exist:", "exist", exist)
		if exist {
			ms.updateHomePosts(ctx, post)
		} else {
			ms.addToHomePosts(ctx, post)
		}

		topicExist := ms.k.IsPostInTopics(ctx, post.Id)
		if topicExist {
			ms.updateTopicPosts(ctx, post, uintExponent)
		}
		categoryExist := ms.k.IsPostInCategory(ctx, post.Id)
		if categoryExist {
			ms.updateCategoryPosts(ctx, post)
		}

	}

	// update post
	blockTime := ctx.BlockTime().Unix()
	post.LikeCount += 1
	post.HomePostsUpdate = blockTime
	ms.k.SetPost(ctx, post)

	// set likes I made
	likesIMade := types.LikesIMade{
		PostId:    post.Id,
		Timestamp: blockTime,
	}
	ms.k.SetLikesIMade(ctx, likesIMade, msg.Sender)
	ms.k.MarkUserLikedPost(ctx, msg.Sender, post.Id)

	// set likes received
	likesReceived := types.LikesReceived{
		LikerAddress: msg.Sender,
		PostId:       post.Id,
		LikeType:     types.LikeType_LIKE,
		Timestamp:    blockTime,
	}
	ms.k.SetLikesReceived(ctx, likesReceived, post.Creator)

	// like add to activitiesReceived
	ms.addActivitiesReceived(ctx, post, "", "", msg.Sender, post.Creator, profiletypes.ActivitiesType_ACTIVITIES_LIKE)

	ms.k.ProfileKeeper.CheckAndCreateUserHandle(ctx, msg.Sender)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeLikePost,
			sdk.NewAttribute(types.AttributeKeyPostID, msg.Id),
			sdk.NewAttribute(types.AttributeKeySender, msg.Sender),
		),
	})

	return &types.MsgLikeResponse{Status: true}, nil
}

// UnlikePost implements types.MsgServer.
func (ms msgServer) Unlike(goCtx context.Context, msg *types.MsgUnlikeRequest) (*types.MsgUnlikeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	post, found := ms.k.GetPost(ctx, msg.Id)
	if !found {
		return nil, errors.Wrap(types.ErrPostNotFound, msg.Id)
	}

	if post.LikeCount > 0 {
		post.LikeCount -= 1
	} else {
		return nil, errors.Wrap(types.ErrInvalidLikeCount, "like count is already zero")
	}

	ms.k.SetPost(ctx, post)

	// remove from likes I made
	err := ms.k.RemoveFromLikesIMade(ctx, msg.Sender, msg.Id)
	if err != nil {
		return nil, errors.Wrap(types.ErrLikesIMadeRemove, msg.Id)
	}

	err2 := ms.k.RemoveFromLikesReceived(ctx, post.Creator, msg.Sender, msg.Id)
	if err2 != nil {
		return nil, errors.Wrap(types.ErrLikesReceivedRemove, msg.Id)
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeUnlikePost,
			sdk.NewAttribute(types.AttributeKeyPostID, msg.Id),
			sdk.NewAttribute(types.AttributeKeySender, msg.Sender),
		),
	})

	return &types.MsgUnlikeResponse{Status: true}, nil
}

// SavePost implements types.MsgServer.
func (ms msgServer) SavePost(goCtx context.Context, msg *types.MsgSaveRequest) (*types.MsgSaveResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	hasSaved := ms.k.HasUserSavedPost(ctx, msg.Sender, msg.Id)
	if hasSaved {
		return nil, errors.Wrap(types.ErrAlreadySaved, "user has already saved this post")
	}
	// Retrieve the post by ID
	post, found := ms.k.GetPost(ctx, msg.Id)
	if !found {
		return nil, errors.Wrap(types.ErrPostNotFound, msg.Id)
	}

	// Check if the post type is COMMENT
	if post.PostType == types.PostType_COMMENT {
		return nil, errors.Wrap(types.ErrInvalidPostType, "cannot save a comment type post")
	}

	blockTime := ctx.BlockTime().Unix()
	post.HomePostsUpdate = blockTime
	ms.k.SetPost(ctx, post)

	// set saves I made
	likesIMade := types.LikesIMade{
		PostId:    post.Id,
		Timestamp: blockTime,
	}
	ms.k.SetSavesIMade(ctx, likesIMade, msg.Sender)
	ms.k.MarkUserSavedPost(ctx, msg.Sender, post.Id)

	// set likes received
	likesReceived := types.LikesReceived{
		LikerAddress: msg.Sender,
		PostId:       post.Id,
		LikeType:     types.LikeType_SAVE,
		Timestamp:    blockTime,
	}
	ms.k.SetLikesReceived(ctx, likesReceived, post.Creator)

	// save add to activitiesReceived
	ms.addActivitiesReceived(ctx, post, "", "", msg.Sender, post.Creator, profiletypes.ActivitiesType_ACTIVITIES_SAVE)

	ms.k.ProfileKeeper.CheckAndCreateUserHandle(ctx, msg.Sender)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeSavePost,
			sdk.NewAttribute(types.AttributeKeyPostID, msg.Id),
			sdk.NewAttribute(types.AttributeKeySender, msg.Sender),
		),
	})

	return &types.MsgSaveResponse{Status: true}, nil
}

// UnSavePost implements types.MsgServer.
func (ms msgServer) UnsavePost(goCtx context.Context, msg *types.MsgUnsaveRequest) (*types.MsgUnsaveResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	post, found := ms.k.GetPost(ctx, msg.Id)
	if !found {
		return nil, errors.Wrap(types.ErrPostNotFound, msg.Id)
	}

	// remove from saves I made
	err := ms.k.RemoveFromSavesIMade(ctx, msg.Sender, msg.Id)
	if err != nil {
		return nil, errors.Wrap(types.ErrSavesIMadeRemove, msg.Id)
	}

	err2 := ms.k.RemoveFromLikesReceived(ctx, post.Creator, msg.Sender, msg.Id)
	if err2 != nil {
		return nil, errors.Wrap(types.ErrLikesReceivedRemove, msg.Id)
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeUnSavePost,
			sdk.NewAttribute(types.AttributeKeyPostID, msg.Id),
			sdk.NewAttribute(types.AttributeKeySender, msg.Sender),
		),
	})

	return &types.MsgUnsaveResponse{Status: true}, nil
}

// Comment implements types.MsgServer.
func (ms msgServer) Comment(goCtx context.Context, msg *types.MsgCommentRequest) (*types.MsgCommentResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// Validate the message
	if len(msg.ParentId) == 0 {
		return nil, errors.Wrapf(types.ErrInvalidRequest, "parent id cannot be empty")
	}
	if len(msg.Comment) == 0 {
		return nil, errors.Wrapf(types.ErrInvalidRequest, "Comment cannot be empty")
	}

	// Validate sender address
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errors.Wrapf(types.ErrInvalidAddress, "Invalid sender address: %s", err)
	}

	blockTime := ctx.BlockTime().Unix()
	data := fmt.Sprintf("%s|%s|%s|%d", msg.Creator, msg.ParentId, msg.Comment, blockTime)
	commentID := ms.k.sha256Generate(data)

	// Create the post
	comment := types.Post{
		Id:        commentID,
		PostType:  types.PostType_COMMENT,
		ParentId:  msg.ParentId,
		Content:   msg.Comment,
		Creator:   msg.Creator,
		Timestamp: blockTime,
	}

	// Store the post in the state
	ms.k.SetPost(ctx, comment)
	// post reward
	ms.k.PostReward(ctx, comment)

	post, found := ms.k.GetPost(ctx, msg.ParentId)
	if !found {
		return nil, errors.Wrap(types.ErrPostNotFound, msg.ParentId)
	}

	oldScore := post.Score
	// Score Accumulation
	uintExponent := ms.ScoreAccumulation(ctx, msg.Creator, post, 3)
	if uintExponent >= 1 {
		newScore := oldScore + uintExponent
		post.Score = newScore
	}
	// update commentList
	if post.PostType == types.PostType_COMMENT {
		ms.k.DeleteFromCommentList(ctx, post.ParentId, post.Id, oldScore)
		ms.k.AddToCommentList(ctx, post.ParentId, post.Id, post.Score)
	}

	if post.PostType != types.PostType_COMMENT {
		exist := ms.k.IsPostInHomePosts(ctx, post.Id, post.HomePostsUpdate)
		if exist {
			ms.updateHomePosts(ctx, post)
		} else {
			ms.addToHomePosts(ctx, post)
		}

		topicExist := ms.k.IsPostInTopics(ctx, post.Id)
		if topicExist {
			ms.updateTopicPosts(ctx, post, uintExponent)
		}
		categoryExist := ms.k.IsPostInCategory(ctx, post.Id)
		if categoryExist {
			ms.updateCategoryPosts(ctx, post)
		}
	}
	//ms.addCommentList(ctx, post, commentID)
	ms.k.AddToCommentList(ctx, post.Id, commentID, 0)

	// update post
	post.CommentCount += 1
	post.HomePostsUpdate = blockTime
	ms.k.SetPost(ctx, post)

	ms.k.ProfileKeeper.CheckAndCreateUserHandle(ctx, msg.Creator)
	ms.k.SetCommentsReceived(ctx, post.Creator, commentID)

	// comment add to activitiesReceived
	ms.addActivitiesReceived(ctx, post, commentID, msg.Comment, msg.Creator, post.Creator, profiletypes.ActivitiesType_ACTIVITIES_COMMENT)

	// mentions add to activitiesReceived
	userHandleList := msg.Mention
	if len(userHandleList) > 0 {
		if len(userHandleList) > 10 {
			return nil, fmt.Errorf("cannot mention more than 10 users")
		}
		for _, userHandle := range userHandleList {
			address := ms.k.ProfileKeeper.GetAddressByUserHandle(ctx, userHandle)
			if address != "" {
				ms.addActivitiesReceived(ctx, post, "", "", msg.Creator, address, profiletypes.ActivitiesType_ACTIVITIES_MENTION)
			}
		}
	}

	//Emit an event for the creation
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeCreateFreePostWithTitle,
			sdk.NewAttribute(types.AttributeKeyCreator, msg.Creator),
			sdk.NewAttribute(types.AttributeKeyCommentID, commentID),
			sdk.NewAttribute(types.AttributeKeyTimestamp, fmt.Sprintf("%d", blockTime)),
		),
	})
	return &types.MsgCommentResponse{}, nil
}

func (ms msgServer) updateHomePosts(ctx sdk.Context, post types.Post) {
	ms.k.DeleteFromHomePostsByPostId(ctx, post.Id, post.HomePostsUpdate)
	ms.k.SetHomePosts(ctx, post.Id)
	count, b := ms.k.GetHomePostsCount(ctx)
	if !b {
		panic("GetHomePostsCount error")
	}
	if count > types.HomePostsCount {
		ms.k.DeleteLastPostFromHomePosts(ctx)
		count -= 1
		ms.k.SetHomePostsCount(ctx, count)
	}
}

func (ms msgServer) addToHomePosts(ctx sdk.Context, post types.Post) {
	ms.k.SetHomePosts(ctx, post.Id)
	count, b := ms.k.GetHomePostsCount(ctx)
	if !b {
		panic("GetHomePostsCount error")
	}
	count += 1
	if count > types.HomePostsCount {
		ms.k.DeleteLastPostFromHomePosts(ctx)
	} else {
		ms.k.SetHomePostsCount(ctx, count)
	}
}

func (ms msgServer) addToUserCreatedPosts(ctx sdk.Context, creator string, post types.Post) {
	ms.k.AddToUserCreatedPosts(ctx, creator, post.Id)
	count, b := ms.k.GetUserCreatedPostsCount(ctx, creator)
	if !b {
		panic("GetUserCreatedPostsCount error")
	}
	count += 1
	if count > types.UserCreatedPostsCount {
		ms.k.DeleteLastPostFromUserCreated(ctx, creator)
	} else {
		ms.k.SetUserCreatedPostsCount(ctx, creator, count)
	}
}

func (ms msgServer) addToTopicPosts(ctx sdk.Context, topicHash string, postId string) {
	ms.k.SetTopicPosts(ctx, topicHash, postId)
	count, b := ms.k.GetTopicPostsCount(ctx, topicHash)
	if !b {
		panic("GetTopicPostsCount error")
	}
	count += 1
	if count > types.TopicPostsCount {
		ms.k.DeleteLastPostFromTopicPosts(ctx, topicHash)
	} else {
		ms.k.SetTopicPostsCount(ctx, topicHash, count)
	}
}

func (ms msgServer) updateTopicPosts(ctx sdk.Context, post types.Post, uintExponent uint64) {
	topics := ms.k.GetTopicsByPostId(ctx, post.Id)
	ms.k.Logger().Warn("=======topics:", "topics", topics)
	if len(topics) > 0 {
		for _, topicHash := range topics {
			ms.k.Logger().Warn("=======updateTopicPosts time:", "HomePostsUpdate", post.HomePostsUpdate)
			ms.k.DeleteFromTopicPostsByTopicAndPostId(ctx, topicHash, post.Id, post.HomePostsUpdate)
			ms.k.SetTopicPosts(ctx, topicHash, post.Id)
			count, b := ms.k.GetTopicPostsCount(ctx, topicHash)
			if !b {
				panic("GetTopicPostsCount error")
			}
			if count > types.TopicPostsCount {
				ms.k.DeleteLastPostFromTopicPosts(ctx, topicHash)
				count -= 1
				ms.k.SetTopicPostsCount(ctx, topicHash, count)
			}

			//update topic
			topic, _ := ms.k.GetTopic(ctx, topicHash)
			oldScore := topic.Score

			if topic.Id != "" {
				newScore := oldScore + uintExponent
				topic.Score = newScore
				topic.LikeCount += 1
				ms.k.AddTopic(ctx, topic)

				ms.k.Logger().Warn("==========score:", "old", oldScore, "new", newScore)
				// update hotTopics72
				createTime := topic.CreateTime
				blockTime := ctx.BlockTime().Unix()
				isWithin72 := isWithin72Hours(createTime, blockTime)
				ms.k.deleteFormHotTopics72(ctx, topicHash, oldScore)
				if isWithin72 {
					ms.k.addToHotTopics72(ctx, topicHash, newScore)
				}
				hotTopics72Count, b := ms.k.GetHotTopics72Count(ctx)
				if !b {
					panic("GetHotTopics72Count error")
				}
				hotTopics72Count += 1
				if hotTopics72Count > types.HotTopics72Count {
					ms.k.DeleteLastFromHotTopics72(ctx)
				} else {
					ms.k.SetHotTopics72Count(ctx, count)
				}

			}
		}
	}
}

func isWithin72Hours(createTime int64, blockTime int64) bool {
	const seventyTwoHoursInSeconds int64 = 72 * 3600 // 259200秒
	timeDifference := blockTime - createTime
	if timeDifference < 0 {
		return false
	}
	return timeDifference <= seventyTwoHoursInSeconds
}

func (ms msgServer) ScoreAccumulation(ctx sdk.Context, operator string, post types.Post, num int64) uint64 {
	operatorProfile, b1 := ms.k.ProfileKeeper.GetProfile(ctx, operator)
	creatorProfile, b2 := ms.k.ProfileKeeper.GetProfile(ctx, post.Creator)
	uintExponent := uint64(0)
	// score
	if b1 && b2 {
		operatorLevel := operatorProfile.Level
		operatorLevelSigned := int64(operatorLevel)

		creatorScore := creatorProfile.Score
		//creatorScoreSigned := int64(creatorScore)
		//postScore := post.Score
		exponent := math.Pow(5, float64(operatorLevelSigned-num))
		if exponent >= 1 {
			uintExponent = uint64(exponent)
			//postScore += uint64(exponent)
			//post.Score = postScore

			creatorScore += uint64(exponent)
			creatorProfile.Score = creatorScore

			// user level
			level := creatorProfile.Level
			levelSigned := int64(level)
			pow := math.Pow(5, float64(levelSigned-1))
			if creatorScore >= uint64(1000*pow) {
				level += 1
				creatorProfile.Level = level
			}
			ms.k.ProfileKeeper.SetProfile(ctx, creatorProfile)
		}
	}
	return uintExponent
}

// Mention implements types.MsgServer.
//func (ms msgServer) Mention(ctx context.Context, msg *types.MsgMentionRequest) (*types.MsgMentionResponse, error) {
//	sdkCtx := sdk.UnwrapSDKContext(ctx)
//	json := msg.MentionJson
//	id := json.Id
//	addressList := json.MentionedAddress
//	if id == "" {
//		return nil, fmt.Errorf("id cannot be empty")
//	}
//	if len(addressList) == 0 {
//		return nil, fmt.Errorf("no users to mention")
//	}
//	if len(addressList) > 10 {
//		return nil, fmt.Errorf("cannot mention more than 10 users")
//	}
//	blockTime := sdkCtx.BlockTime().Unix()
//
//	post, _ := ms.k.GetPost(sdkCtx, id)
//	imageUrl := ""
//	if len(post.ImagesUrl) > 0 {
//		imageUrl = post.ImagesUrl[0]
//	}
//	for _, address := range addressList {
//		activitiesReceived := profiletypes.ActivitiesReceived{
//			Address:        address,
//			TargetAddress:  post.Creator,
//			ParentId:       id,
//			ActivitiesType: profiletypes.ActivitiesType_ACTIVITIES_MENTION,
//			ParentImageUrl: imageUrl,
//			Timestamp:      blockTime,
//		}
//		ms.k.ProfileKeeper.SetActivitiesReceived(sdkCtx, activitiesReceived, address, msg.Creator)
//		count, b := ms.k.ProfileKeeper.GetActivitiesReceivedCount(sdkCtx, address)
//		if !b {
//			panic("GetActivitiesReceivedCount error")
//		}
//		count += 1
//		if count > profiletypes.ActivitiesReceivedCount {
//			ms.k.ProfileKeeper.DeleteLastActivitiesReceived(sdkCtx, address)
//		}
//		ms.k.ProfileKeeper.SetActivitiesReceivedCount(sdkCtx, address, count)
//	}
//	return &types.MsgMentionResponse{}, nil
//}

func (ms msgServer) addActivitiesReceived(sdkCtx sdk.Context, parentPost types.Post, commentId string, content string,
	operator string, target string, activitiesType profiletypes.ActivitiesType) {
	blockTime := sdkCtx.BlockTime().Unix()
	imageUrl := ""
	if len(parentPost.ImagesUrl) > 0 {
		imageUrl = parentPost.ImagesUrl[0]
	}
	activitiesReceived := profiletypes.ActivitiesReceived{
		Address:        operator,
		TargetAddress:  target,
		ActivitiesType: activitiesType,
		ParentImageUrl: imageUrl,
		Timestamp:      blockTime,
	}
	if activitiesType == profiletypes.ActivitiesType_ACTIVITIES_LIKE || activitiesType == profiletypes.ActivitiesType_ACTIVITIES_SAVE {
		activitiesReceived.ParentId = parentPost.Id
	} else if activitiesType == profiletypes.ActivitiesType_ACTIVITIES_COMMENT {
		activitiesReceived.CommentId = commentId
		activitiesReceived.Content = content
		activitiesReceived.ParentId = parentPost.Id
	} else if activitiesType == profiletypes.ActivitiesType_ACTIVITIES_MENTION {
		activitiesReceived.ParentId = parentPost.Id
	}

	ms.k.ProfileKeeper.SetActivitiesReceived(sdkCtx, activitiesReceived, target, operator)
	count, b := ms.k.ProfileKeeper.GetActivitiesReceivedCount(sdkCtx, target)
	if !b {
		panic("GetActivitiesReceivedCount error")
	}
	count += 1
	if count > profiletypes.ActivitiesReceivedCount {
		ms.k.ProfileKeeper.DeleteLastActivitiesReceived(sdkCtx, target)
	}
	ms.k.ProfileKeeper.SetActivitiesReceivedCount(sdkCtx, target, count)
}

func (ms msgServer) handleCategoryTopicPost(ctx sdk.Context, topicList []string, category string, blockTime int64, postId string) error {
	if len(topicList) > 0 {
		if len(topicList) > 10 {
			return fmt.Errorf("topic list length can not be larger than 10")
		}
		var hashTopics []string
		ms.k.Logger().Warn("==========topicList:", "topicList", topicList)
		for _, topicName := range topicList {
			topicHash := ms.k.sha256Generate(topicName)
			// add topic
			exists := ms.k.TopicExists(ctx, topicHash)
			if !exists {
				topic := types.Topic{
					Id:         topicHash,
					Name:       topicName,
					CreateTime: blockTime,
					UpdateTime: blockTime,
				}
				ms.k.AddTopic(ctx, topic)

				ms.k.addToHotTopics72(ctx, topicHash, 0)
			}

			// add topic search
			ms.k.SetTopicSearch(ctx, topicName)
			ms.addToTopicPosts(ctx, topicHash, postId)
			hashTopics = append(hashTopics, topicHash)

			// add category posts
			categoryDb := ms.k.getCategoryByTopicHash(ctx, topicHash)
			if categoryDb != "" {
				//ms.k.SetCategorySearch(ctx, categoryDb)
				categoryHash := ms.k.sha256Generate(categoryDb)
				ms.addToCategoryPosts(ctx, categoryHash, postId)
				ms.k.SetPostCategoryMapping(ctx, categoryHash, postId)
			} else {
				if category != "" {
					categoryHash := ms.k.sha256Generate(category)
					exists := ms.k.CategoryExists(ctx, categoryHash)
					if exists {
						//ms.k.SetCategorySearch(ctx, category)
						ms.addToCategoryPosts(ctx, categoryHash, postId)
						topic, _ := ms.k.GetTopic(ctx, topicHash)
						ms.addToCategoryTopics(ctx, categoryHash, topic)

						ms.k.SetPostCategoryMapping(ctx, categoryHash, postId)
						ms.k.SetTopicCategoryMapping(ctx, topicHash, category)
					} else {
						isUncategorizedTopic := ms.k.IsUncategorizedTopic(ctx, topicHash)
						if !isUncategorizedTopic {
							ms.k.SetUncategorizedTopics(ctx, topicHash)
							count, _ := ms.k.GetUncategorizedTopicsCount(ctx)
							count += 1
							ms.k.SetUncategorizedTopicsCount(ctx, uint64(count))
						}
					}
				}
			}

		}
		ms.k.SetPostTopicsMapping(ctx, hashTopics, postId)
	} else {
		// connect category and post
		if category != "" {
			categoryHash := ms.k.sha256Generate(category)
			exists := ms.k.CategoryExists(ctx, categoryHash)
			if exists {
				//ms.k.SetCategorySearch(ctx, category)
				ms.addToCategoryPosts(ctx, categoryHash, postId)
				ms.k.SetPostCategoryMapping(ctx, categoryHash, postId)
			}
		}
	}
	return nil
}

func (ms msgServer) addToCategoryPosts(ctx sdk.Context, categoryHash string, postId string) {
	ms.k.SetCategoryPosts(ctx, categoryHash, postId)
	count, b := ms.k.GetCategoryPostsCount(ctx, categoryHash)
	if !b {
		panic("GetCategoryPostsCount error")
	}
	count += 1
	if count > types.CategoryPostsCount {
		ms.k.DeleteLastPostFromCategoryPosts(ctx, categoryHash)
	} else {
		ms.k.SetCategoryPostsCount(ctx, categoryHash, count)
	}
}
func (ms msgServer) updateCategoryPosts(ctx sdk.Context, post types.Post) {
	category := ms.k.GetCategoryByPostId(ctx, post.Id)
	ms.k.DeleteFromCategoryPostsByCategoryAndPostId(ctx, category, post.Id, post.HomePostsUpdate)
	ms.k.SetCategoryPosts(ctx, category, post.Id)
	count, b := ms.k.GetCategoryPostsCount(ctx, category)
	if !b {
		panic("GetCategoryPostsCount error")
	}
	if count > types.CategoryPostsCount {
		ms.k.DeleteLastPostFromCategoryPosts(ctx, category)
		count -= 1
		ms.k.SetCategoryPostsCount(ctx, category, count)
	}
}
func (ms msgServer) addToCategoryTopics(ctx sdk.Context, categoryHash string, topic types.Topic) {
	ms.k.SetCategoryTopics(ctx, topic.Score, categoryHash, topic.Id)
	count, b := ms.k.GetCategoryTopicsCount(ctx, categoryHash)
	if !b {
		panic("GetCategoryTopicsCount error")
	}
	count += 1
	if count > types.CategoryTopicsCount {
		ms.k.DeleteLastPostFromCategoryTopics(ctx, categoryHash)
	} else {
		ms.k.SetCategoryTopicsCount(ctx, categoryHash, count)
	}
}

// CastVoteOnPoll implements types.MsgServer.
func (ms msgServer) CastVoteOnPoll(goCtx context.Context, msg *types.CastVoteOnPollRequest) (*types.CastVoteOnPollResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	blockTime := ctx.BlockTime().Unix()
	parentPost, _ := ms.k.GetPost(ctx, msg.Id)

	if blockTime < parentPost.Poll.VotingStart {
		return nil, errors.Wrapf(types.ErrVotingNotStarted, "voting has not started yet")
	}
	if blockTime > parentPost.Poll.VotingEnd {
		return nil, errors.Wrapf(types.ErrVotingEnded, "voting has ended")
	}
	_, b := ms.k.GetPoll(ctx, msg.Id, msg.Creator)
	if b {
		return nil, errors.Wrapf(types.ErrAlreadyVoted, "already voted")
	}
	ms.k.SetPoll(ctx, msg.Id, msg.Creator, msg.OptionId)
	parentPost.Poll.TotalVotes += 1
	voteList := parentPost.Poll.Vote
	for i, _ := range voteList {
		if parentPost.Poll.Vote[i].Id == msg.OptionId {
			parentPost.Poll.Vote[i].Count += 1
			break
		}
	}
	ms.k.SetPost(ctx, parentPost)
	return &types.CastVoteOnPollResponse{Status: true}, nil
}

// AddCategory implements types.MsgServer.
func (ms msgServer) AddCategory(ctx context.Context, msg *types.AddCategoryRequest) (*types.AddCategoryResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	creator := msg.Creator
	isAdmin := ms.k.ProfileKeeper.IsAdmin(sdkCtx, creator)
	if isAdmin {
		params := msg.Params
		name := params.Name
		id := ms.k.sha256Generate(name)
		index := params.Index
		if index == 0 {
			categoryIndex, _ := ms.k.GetCategoryIndex(sdkCtx)
			index = categoryIndex + 1
		}
		category := types.Category{
			Id:     id,
			Name:   params.Name,
			Avatar: params.Avatar,
			Index:  index,
		}
		ms.k.AddCategory(sdkCtx, category)
		ms.k.AddCategoryWithIndex(sdkCtx, category)
		ms.k.SetCategoryIndex(sdkCtx, index)

		sdkCtx.EventManager().EmitEvents(sdk.Events{
			sdk.NewEvent(
				types.EventTypeAddCategory,
				sdk.NewAttribute(types.AttributeKeyCreator, msg.Creator),
				sdk.NewAttribute(types.AttributeKeyCategoryID, id),
				sdk.NewAttribute(types.AttributeKeyTimestamp, fmt.Sprintf("%d", sdkCtx.BlockTime())),
			),
		})
		return &types.AddCategoryResponse{
			Status: true,
		}, nil
	} else {
		return &types.AddCategoryResponse{
			Status: false,
		}, errors.Wrapf(types.ErrRequestDenied, "request denied")
	}
}

// DeleteCategory implements types.MsgServer.
func (ms msgServer) DeleteCategory(ctx context.Context, msg *types.DeleteCategoryRequest) (*types.DeleteCategoryResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	creator := msg.Creator
	isAdmin := ms.k.ProfileKeeper.IsAdmin(sdkCtx, creator)
	if isAdmin {
		ms.k.DeleteCategory(sdkCtx, msg.Id)
		sdkCtx.EventManager().EmitEvents(sdk.Events{
			sdk.NewEvent(
				types.EventTypeDeleteCategory,
				sdk.NewAttribute(types.AttributeKeyCreator, msg.Creator),
				sdk.NewAttribute(types.AttributeKeyCategoryID, msg.Id),
				sdk.NewAttribute(types.AttributeKeyTimestamp, fmt.Sprintf("%d", sdkCtx.BlockTime())),
			),
		})
		return &types.DeleteCategoryResponse{
			Status: true,
		}, nil
	} else {
		return &types.DeleteCategoryResponse{
			Status: false,
		}, errors.Wrapf(types.ErrRequestDenied, "request denied")
	}
}

// UpdateTopic implements types.MsgServer.
func (ms msgServer) UpdateTopic(goCtx context.Context, msg *types.UpdateTopicRequest) (*types.UpdateTopicResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	creator := msg.Creator
	isAdmin := ms.k.ProfileKeeper.IsAdmin(ctx, creator)
	if isAdmin {
		json := msg.TopicJson
		id := json.Id
		topic, _ := ms.k.GetTopic(ctx, id)
		ms.k.deleteFormHotTopics72(ctx, id, topic.Score)
		topic.Score = json.Score
		topic.Avatar = json.Avatar
		ms.k.AddTopic(ctx, topic)
		ms.k.addToHotTopics72(ctx, id, topic.Score)
	}
	return &types.UpdateTopicResponse{}, nil
}

// FollowTopic implements types.MsgServer.
func (ms msgServer) FollowTopic(goCtx context.Context, msg *types.MsgFollowTopicRequest) (*types.MsgFollowTopicResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	follower := msg.Creator
	topicHash := msg.TopicId
	isFollowing := ms.k.IsFollowingTopic(ctx, follower, topicHash)
	if !isFollowing {
		exists := ms.k.TopicExists(ctx, topicHash)
		if exists {
			ms.k.FollowTopic(ctx, follower, topicHash)
			ms.k.SetFollowTopicTime(ctx, follower, topicHash)
		}
	}
	return &types.MsgFollowTopicResponse{
		Status: true,
	}, nil
}

func (ms msgServer) UnfollowTopic(goCtx context.Context, msg *types.MsgUnfollowTopicRequest) (*types.MsgUnfollowTopicResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	topicHash := msg.TopicId
	time, _ := ms.k.GetFollowTopicTime(ctx, msg.Creator, topicHash)
	ms.k.UnfollowTopic(ctx, msg.Creator, time, topicHash)
	ms.k.DeleteFollowTopicTime(ctx, msg.Creator, topicHash)

	return &types.MsgUnfollowTopicResponse{
		Status: true,
	}, nil
}

// ClassifyUncategorizedTopic implements types.MsgServer.
func (ms msgServer) ClassifyUncategorizedTopic(goCtx context.Context, msg *types.ClassifyUncategorizedTopicRequest) (*types.ClassifyUncategorizedTopicResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	creator := msg.Creator
	profile, _ := ms.k.ProfileKeeper.GetProfile(ctx, creator)
	adminLevel := profile.AdminLevel
	if adminLevel > 0 {
		operator, _ := ms.k.GetCategoryOperator(ctx)
		if operator != "" {
			operatorProfile, _ := ms.k.ProfileKeeper.GetProfile(ctx, operator)
			if adminLevel < 4 && operatorProfile.AdminLevel > adminLevel {
				return nil, errors.Wrapf(types.ErrRequestDenied, "request denied")
			}
			topic, _ := ms.k.GetTopic(ctx, msg.TopicId)
			ms.addToCategoryTopics(ctx, msg.CategoryId, topic)
			ms.k.RemoveFromUncategorizedTopics(ctx, msg.TopicId)
			ms.k.SetCategoryOperator(ctx, creator)
		}
	}
	return &types.ClassifyUncategorizedTopicResponse{Status: true}, nil
}
