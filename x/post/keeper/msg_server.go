package keeper

import (
	"context"
	"fmt"
	"math"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/rollchains/tlock/x/post/types"
	profiletypes "github.com/rollchains/tlock/x/profile/types"
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

// Helper function to validate addresses
func (ms msgServer) validateAddress(address string) error {
	if _, err := sdk.AccAddressFromBech32(address); err != nil {
		return types.NewInvalidAddressErrorf("invalid address %s: %s", address, err)
	}
	return nil
}

// Helper function to validate post ID
func (ms msgServer) validatePostID(postID string) error {
	return types.ValidatePostID(postID)
}

// Helper function to get post with validation
func (ms msgServer) getPostWithValidation(ctx sdk.Context, postID string) (types.Post, error) {
	if err := ms.validatePostID(postID); err != nil {
		return types.Post{}, err
	}

	post, found := ms.k.GetPost(ctx, postID)
	if !found {
		return types.Post{}, types.NewPostNotFoundError(postID)
	}

	return post, nil
}

// Helper function to log and return error
func (ms msgServer) logAndReturnError(operation string, err error, context ...interface{}) error {
	types.LogError(ms.k.logger, operation, err, context...)
	return err
}

func (ms msgServer) UpdateParams(ctx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if ms.k.authority != msg.Authority {
		return nil, types.NewInvalidAddressErrorf("invalid authority; expected %s, got %s", ms.k.authority, msg.Authority)
	}

	return nil, ms.k.Params.Set(ctx, msg.Params)
}

// SetServiceName implements types.MsgServer.
func (ms msgServer) SetServiceName(ctx context.Context, msg *types.MsgSetServiceName) (*types.MsgSetServiceNameResponse, error) {
	// ctx := sdk.UnwrapSDKContext(goCtx)
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
	if senderErr != nil {
		return &types.MsgGrantAllowanceFromModuleResponse{Status: false}, types.NewInvalidAddressErrorf("invalid sender address: %s", senderErr)
	}
	userAddress, userErr := sdk.AccAddressFromBech32(msg.UserAddress)
	if userErr != nil {
		return &types.MsgGrantAllowanceFromModuleResponse{Status: false}, types.NewInvalidAddressErrorf("invalid user address: %s", userErr)
	}
	ms.k.GrantPeriodicAllowance(ctx, sender, userAddress)

	return &types.MsgGrantAllowanceFromModuleResponse{Status: true}, nil
}

func (ms msgServer) validateCreatePostRequest(msg *types.MsgCreatePost) error {
	postDetail := msg.GetPostDetail()

	// Validate content
	if err := types.ValidatePostContent(postDetail.Content); err != nil {
		return err
	}

	// Validate title if present
	if postDetail.Title != "" {
		if err := types.ValidateTitle(postDetail.Title); err != nil {
			return err
		}
	}

	// Validate mentions
	if err := types.ValidateMentions(postDetail.Mention); err != nil {
		return err
	}

	// Validate topics
	if err := types.ValidateTopics(postDetail.Topic); err != nil {
		return err
	}

	// Validate images
	if err := types.ValidateImages(postDetail.ImagesBase64); err != nil {
		return err
	}

	// Validate category if present
	if postDetail.Category != "" {
		if err := types.ValidateCategory(postDetail.Category); err != nil {
			return err
		}
	}

	// Validate image sizes
	for _, img := range postDetail.ImagesBase64 {
		if len(img) > MaxImageSize {
			return types.ErrImageTooLarge
		}
	}

	return nil
}

// CreatePost implements types.MsgServer.
func (ms msgServer) CreatePost(goCtx context.Context, msg *types.MsgCreatePost) (*types.MsgCreatePostResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	postDetail := msg.GetPostDetail()
	// Validate params
	err := ms.validateCreatePostRequest(msg)
	if err != nil {
		return nil, err
	}

	// Validate address
	if err := ms.validateAddress(msg.Creator); err != nil {
		return nil, err
	}

	//if len(msg.Image) > MaxImageSize {
	//	return nil, errors.Wrap(types.ErrInvalidRequest, "Image size exceeds the maximum allowed limit")
	//}

	blockTime := ctx.BlockTime().Unix()
	// Generate a unique post ID
	var data string
	var postType types.PostType
	if postDetail.Title != "" {
		if err := types.ValidatePostWithTitleContent(postDetail.Content); err != nil {
			return nil, err
		}
		postType = types.PostType_ARTICLE
		data = fmt.Sprintf("%s|%s|%s|%d", msg.Creator, postDetail.Title, postDetail.Content, blockTime)
	} else {
		if err := types.ValidatePostContent(postDetail.Content); err != nil {
			return nil, err
		}
		postType = types.PostType_ORIGINAL
		data = fmt.Sprintf("%s|%s|%d", msg.Creator, postDetail.Content, blockTime)
	}
	postId := ms.k.sha256Generate(data)

	// Create the post
	post := types.Post{
		Id: postId,
		//PostType:        types.PostType_ORIGINAL,
		PostType: postType,
		//Title:           postDetail.Title,
		Content:         postDetail.Content,
		Creator:         msg.Creator,
		Timestamp:       blockTime,
		ImagesUrl:       postDetail.ImagesUrl,
		VideosUrl:       postDetail.VideosUrl,
		HomePostsUpdate: blockTime,
		Poll:            postDetail.Poll,
	}
	if postDetail.Poll != nil {
		post.PostType = types.PostType_POLL
	}
	if postDetail.GetTitle() != "" {
		post.Title = postDetail.GetTitle()
		//post.PostType = types.PostType_ARTICLE
	}

	imagesBase64 := postDetail.ImagesBase64
	if len(imagesBase64) > 0 {
		paidImageBase64 := imagesBase64[0]
		imageHash := ms.k.sha256Generate(paidImageBase64)
		setPaidPostImageErr := ms.k.SetPaidPostImage(ctx, imageHash, postDetail.ImagesBase64[0])
		if setPaidPostImageErr != nil {
			return nil, setPaidPostImageErr
		}
		post.ImageIds = []string{imageHash}
	}

	// post payment
	//ms.k.postPayment(ctx, post)
	// Store the post in the state
	ms.k.SetPost(ctx, post)
	// add home posts
	ms.addToHomePosts(ctx, post)
	// add to user created posts
	ms.addToUserCreatedPosts(ctx, msg.Creator, post)

	ms.k.ProfileKeeper.CheckAndCreateUserHandle(ctx, msg.Creator)

	// mentions add to activitiesReceived
	userHandleList := postDetail.Mention
	if len(userHandleList) > 0 {
		for _, userHandle := range userHandleList {
			address := ms.k.ProfileKeeper.GetAddressByUserHandle(ctx, userHandle)
			if address != "" {
				ms.addActivitiesReceived(ctx, post, "", "", msg.Creator, address, profiletypes.ActivitiesType_ACTIVITIES_MENTION)
			}
		}
	}

	topicList := postDetail.Topic
	category := postDetail.Category
	err = ms.handleCategoryTopicPost(ctx, msg.Creator, topicList, category, blockTime, postId)
	if err != nil {
		return nil, err
	}

	//Emit an event for the creation
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeCreatePaidPost,
			sdk.NewAttribute(types.AttributeKeyCreator, msg.Creator),
			sdk.NewAttribute(types.AttributeKeyPostID, postId),
			sdk.NewAttribute(types.AttributeKeyTitle, postDetail.Title),
			sdk.NewAttribute(types.AttributeKeyTimestamp, fmt.Sprintf("%d", blockTime)),
		),
	})
	return &types.MsgCreatePostResponse{
		PostId: postId,
	}, nil
}

// CreateFreePostWithTitle implements types.MsgServer.
//func (ms msgServer) CreateFreePostWithTitle(goCtx context.Context, msg *types.MsgCreateFreePostWithTitle) (*types.MsgCreateFreePostWithTitleResponse, error) {
//	ctx := sdk.UnwrapSDKContext(goCtx)
//
//	// Validate the message
//	if len(msg.Title) == 0 {
//		return nil, errors.Wrapf(types.ErrInvalidRequest, "Title cannot be empty")
//	}
//	if len(msg.Content) == 0 {
//		return nil, errors.Wrapf(types.ErrInvalidRequest, "Content cannot be empty")
//	}
//	if err := types.ValidatePostWithTitleContent(msg.Content); err != nil {
//		return nil, errors.Wrap(types.ErrInvalidRequest, err.Error())
//	}
//
//	// Validate sender address
//	_, err := sdk.AccAddressFromBech32(msg.Creator)
//	if err != nil {
//		return nil, errors.Wrapf(types.ErrInvalidAddress, "Invalid sender address: %s", err)
//	}
//
//	//if len(msg.Image) > MaxImageSize {
//	//	return nil, errors.Wrap(types.ErrInvalidRequest, "Image size exceeds the maximum allowed limit")
//	//}
//
//	blockTime := ctx.BlockTime().Unix()
//	// Generate a unique post ID
//	data := fmt.Sprintf("%s|%s|%s|%d", msg.Creator, msg.Title, msg.Content, blockTime)
//	postId := ms.k.sha256Generate(data)
//
//	// Create the post
//	post := types.Post{
//		Id:              postId,
//		PostType:        types.PostType_ARTICLE,
//		Title:           msg.Title,
//		Content:         msg.Content,
//		Creator:         msg.Creator,
//		Timestamp:       blockTime,
//		ImagesUrl:       msg.ImagesUrl,
//		VideosUrl:       msg.VideosUrl,
//		HomePostsUpdate: blockTime,
//		Poll:            msg.Poll,
//	}
//	if msg.Poll != nil {
//		post.PostType = types.PostType_POLL
//	}
//
//	// Store the post in the state
//	ms.k.SetPost(ctx, post)
//	// post reward
//	ms.k.PostReward(ctx, post)
//	// add home posts
//	ms.addToHomePosts(ctx, post)
//	// add to user created posts
//	ms.addToUserCreatedPosts(ctx, msg.Creator, post)
//
//	userHandleList := msg.Mention
//	if len(userHandleList) > 0 {
//		if len(userHandleList) > 10 {
//			return nil, fmt.Errorf("cannot mention more than 10 users")
//		}
//		for _, userHandle := range userHandleList {
//			address := ms.k.ProfileKeeper.GetAddressByUserHandle(ctx, userHandle)
//			if address != "" {
//				ms.addActivitiesReceived(ctx, post, "", "", msg.Creator, address, profiletypes.ActivitiesType_ACTIVITIES_MENTION)
//			}
//		}
//	}
//
//	topicList := msg.Topic
//	category := msg.Category
//	err = ms.handleCategoryTopicPost(ctx, msg.Creator, topicList, category, blockTime, postId)
//	if err != nil {
//		return nil, err
//	}
//
//	//Emit an event for the creation
//	ctx.EventManager().EmitEvents(sdk.Events{
//		sdk.NewEvent(
//			types.EventTypeCreateFreePostWithTitle,
//			sdk.NewAttribute(types.AttributeKeyCreator, msg.Creator),
//			sdk.NewAttribute(types.AttributeKeyPostID, postId),
//			sdk.NewAttribute(types.AttributeKeyTitle, msg.Title),
//			sdk.NewAttribute(types.AttributeKeyTimestamp, fmt.Sprintf("%d", blockTime)),
//		),
//	})
//
//	return &types.MsgCreateFreePostWithTitleResponse{PostId: postId}, nil
//}

// CreateFreePost implements types.MsgServer.
//func (ms msgServer) CreateFreePost(goCtx context.Context, msg *types.MsgCreateFreePost) (*types.MsgCreateFreePostResponse, error) {
//	//base64.StdEncoding.EncodeToString([]byte("{\"body\":{\"messages\":[{\"@type\":\"/cosmos.bank.v1beta1.MsgSend\",\"from_address\":\"tlock1hj5fveer5cjtn4wd6wstzugjfdxzl0xp5u7j9p\",\"to_address\":\"tlock1efd63aw40lxf3n4mhf7dzhjkr453axurggdkvg\",\"amount\":[{\"denom\":\"uTOK\",\"amount\":\"10\"}]}],\"memo\":\"\",\"timeout_height\":\"0\",\"extension_options\":[],\"non_critical_extension_options\":[]},\"auth_info\":{\"signer_infos\":[{\"public_key\":{\"@type\":\"/cosmos.crypto.secp256k1.PubKey\",\"key\":\"ApZa31BR3NWLylRT6Qi5+f+zXtj2OpqtC76vgkUGLyww\"},\"mode_info\":{\"single\":{\"mode\":\"SIGN_MODE_DIRECT\"}},\"sequence\":\"1\"}],\"fee\":{\"amount\":[{\"denom\":\"uTOK\",\"amount\":\"77886\"}],\"gas_limit\":\"77886\",\"payer\":\"\",\"granter\":\"\"},\"tip\":null},\"signatures\":[\"RWGJJcQ8Ioul52d6HbBW6F1FJwuNPRSTTny6xpAZCfpYgHSIGuk0+uupaC5gx0FKur8qOA9tZlhKhAfLyf9hWg==\"]}"))
//
//	ctx := sdk.UnwrapSDKContext(goCtx)
//
//	// Validate the message
//	if len(msg.Content) == 0 {
//		return nil, errors.Wrapf(types.ErrInvalidRequest, "Content cannot be empty")
//	}
//	if err := types.ValidatePostContent(msg.Content); err != nil {
//		return nil, errors.Wrap(types.ErrInvalidRequest, err.Error())
//	}
//
//	// Validate sender address
//	_, err := sdk.AccAddressFromBech32(msg.Creator)
//	if err != nil {
//		return nil, errors.Wrapf(types.ErrInvalidAddress, "Invalid sender address: %s", err)
//	}
//
//	//if len(msg.Image) > MaxImageSize {
//	//	return nil, errors.Wrap(types.ErrInvalidRequest, "Image size exceeds the maximum allowed limit")
//	//}
//
//	// Generate a unique post ID
//	blockTime := ctx.BlockTime().Unix()
//	data := fmt.Sprintf("%s|%s|%d", msg.Creator, msg.Content, blockTime)
//	postId := ms.k.sha256Generate(data)
//
//	// Create the post
//	post := types.Post{
//		Id:              postId,
//		PostType:        types.PostType_ORIGINAL,
//		Content:         msg.Content,
//		Creator:         msg.Creator,
//		Timestamp:       blockTime,
//		ImagesUrl:       msg.ImagesUrl,
//		VideosUrl:       msg.VideosUrl,
//		HomePostsUpdate: blockTime,
//		Poll:            msg.Poll,
//	}
//	if msg.Poll != nil {
//		post.PostType = types.PostType_POLL
//	}
//
//	// Store the post in the state
//	ms.k.SetPost(ctx, post)
//	// post reward
//	ms.k.PostReward(ctx, post)
//	// add to home posts
//	ms.addToHomePosts(ctx, post)
//	// add to user created posts
//	ms.addToUserCreatedPosts(ctx, msg.Creator, post)
//
//	ms.k.ProfileKeeper.CheckAndCreateUserHandle(ctx, msg.Creator)
//
//	// mentions add to activitiesReceived
//	userHandleList := msg.Mention
//	if len(userHandleList) > 0 {
//		if len(userHandleList) > 10 {
//			return nil, fmt.Errorf("cannot mention more than 10 users")
//		}
//		for _, userHandle := range userHandleList {
//			address := ms.k.ProfileKeeper.GetAddressByUserHandle(ctx, userHandle)
//			if address != "" {
//				ms.addActivitiesReceived(ctx, post, "", "", msg.Creator, address, profiletypes.ActivitiesType_ACTIVITIES_MENTION)
//			}
//		}
//	}
//
//	topicList := msg.Topic
//	category := msg.Category
//	err = ms.handleCategoryTopicPost(ctx, msg.Creator, topicList, category, blockTime, postId)
//	if err != nil {
//		return nil, err
//	}
//
//	//Emit an event for the creation
//	ctx.EventManager().EmitEvents(sdk.Events{
//		sdk.NewEvent(
//			types.EventTypeCreateFreePost,
//			sdk.NewAttribute(types.AttributeKeyCreator, msg.Creator),
//			sdk.NewAttribute(types.AttributeKeyPostID, postId),
//			sdk.NewAttribute(types.AttributeKeyTimestamp, fmt.Sprintf("%d", blockTime)),
//		),
//	})
//	//ms.k.Logger().Warn("============= create free post end ==============")
//	return &types.MsgCreateFreePostResponse{PostId: postId}, nil
//}

// CreateFreePostImagePayable implements types.MsgServer.
//func (ms msgServer) CreateFreePostImagePayable(goCtx context.Context, msg *types.MsgCreateFreePostImagePayable) (*types.MsgCreateFreePostImagePayableResponse, error) {
//	ctx := sdk.UnwrapSDKContext(goCtx)
//
//	// Validate the message
//	if len(msg.Content) == 0 {
//		return nil, errors.Wrapf(types.ErrInvalidRequest, "Content cannot be empty")
//	}
//	if err := types.ValidatePostContent(msg.Content); err != nil {
//		return nil, errors.Wrap(types.ErrInvalidRequest, err.Error())
//	}
//
//	// Validate sender address
//	//sender, err := sdk.AccAddressFromBech32(msg.Sender)
//	_, err := sdk.AccAddressFromBech32(msg.Creator)
//	if err != nil {
//		return nil, errors.Wrapf(types.ErrInvalidAddress, "Invalid sender address: %s", err)
//	}
//
//	//if len(msg.Image) > MaxImageSize {
//	//	return nil, errors.Wrap(types.ErrInvalidRequest, "Image size exceeds the maximum allowed limit")
//	//}
//
//	blockTime := ctx.BlockTime().Unix()
//	data := fmt.Sprintf("%s|%s|%d", msg.Creator, msg.Content, blockTime)
//	postId := ms.k.sha256Generate(data)
//
//	// Create the post
//	post := types.Post{
//		Id:              postId,
//		PostType:        types.PostType_ADVERTISEMENT,
//		Content:         msg.Content,
//		Creator:         msg.Creator,
//		Timestamp:       blockTime,
//		ImagesUrl:       msg.ImagesUrl,
//		VideosUrl:       msg.VideosUrl,
//		HomePostsUpdate: blockTime,
//	}
//	imagesBase64 := msg.ImagesBase64
//	if len(imagesBase64) > 0 {
//		if len(imagesBase64) > 1 {
//			return nil, fmt.Errorf("only one paid image is allowed to be uploaded")
//		}
//		paidImageBase64 := imagesBase64[0]
//		imageHash := ms.k.sha256Generate(paidImageBase64)
//		setPaidPostImageErr := ms.k.SetPaidPostImage(ctx, imageHash, msg.ImagesBase64[0])
//		if setPaidPostImageErr != nil {
//			return nil, setPaidPostImageErr
//		}
//		post.ImageIds = []string{imageHash}
//	}
//
//	// post payment
//	ms.k.postPayment(ctx, post)
//	// Store the post in the state
//	ms.k.SetPost(ctx, post)
//	// add home posts
//	ms.addToHomePosts(ctx, post)
//	// add to user created posts
//	ms.addToUserCreatedPosts(ctx, msg.Creator, post)
//
//	ms.k.ProfileKeeper.CheckAndCreateUserHandle(ctx, msg.Creator)
//
//	// mentions add to activitiesReceived
//	userHandleList := msg.Mention
//	if len(userHandleList) > 0 {
//		if len(userHandleList) > 10 {
//			return nil, fmt.Errorf("cannot mention more than 10 users")
//		}
//		for _, userHandle := range userHandleList {
//			address := ms.k.ProfileKeeper.GetAddressByUserHandle(ctx, userHandle)
//			if address != "" {
//				ms.addActivitiesReceived(ctx, post, "", "", msg.Creator, address, profiletypes.ActivitiesType_ACTIVITIES_MENTION)
//			}
//		}
//	}
//
//	topicList := msg.Topic
//	category := msg.Category
//	err = ms.handleCategoryTopicPost(ctx, msg.Creator, topicList, category, blockTime, postId)
//	if err != nil {
//		return nil, err
//	}
//
//	//Emit an event for the creation
//	ctx.EventManager().EmitEvents(sdk.Events{
//		sdk.NewEvent(
//			types.EventTypeCreatePaidPost,
//			sdk.NewAttribute(types.AttributeKeyCreator, msg.Creator),
//			sdk.NewAttribute(types.AttributeKeyPostID, postId),
//			sdk.NewAttribute(types.AttributeKeyTitle, msg.Title),
//			sdk.NewAttribute(types.AttributeKeyTimestamp, fmt.Sprintf("%d", blockTime)),
//		),
//	})
//	return &types.MsgCreateFreePostImagePayableResponse{}, nil
//}

//func (ms msgServer) CreatePaidPost(goCtx context.Context, msg *types.MsgCreatePaidPost) (*types.MsgCreatePaidPostResponse, error) {
//	ctx := sdk.UnwrapSDKContext(goCtx)
//
//	// Validate the message
//	if len(msg.Content) == 0 {
//		return nil, errors.Wrapf(types.ErrInvalidRequest, "Content cannot be empty")
//	}
//	if err := types.ValidatePostContent(msg.Content); err != nil {
//		return nil, errors.Wrap(types.ErrInvalidRequest, err.Error())
//	}
//
//	// Validate sender address
//	//sender, err := sdk.AccAddressFromBech32(msg.Sender)
//	_, err := sdk.AccAddressFromBech32(msg.Creator)
//	if err != nil {
//		return nil, errors.Wrapf(types.ErrInvalidAddress, "Invalid sender address: %s", err)
//	}
//
//	//if len(msg.Image) > MaxImageSize {
//	//	return nil, errors.Wrap(types.ErrInvalidRequest, "Image size exceeds the maximum allowed limit")
//	//}
//
//	blockTime := ctx.BlockTime().Unix()
//	data := fmt.Sprintf("%s|%s|%d", msg.Creator, msg.Content, blockTime)
//	postId := ms.k.sha256Generate(data)
//
//	// Create the post
//	post := types.Post{
//		Id:              postId,
//		PostType:        types.PostType_ADVERTISEMENT,
//		Content:         msg.Content,
//		Creator:         msg.Creator,
//		Timestamp:       blockTime,
//		ImagesUrl:       msg.ImagesUrl,
//		VideosUrl:       msg.VideosUrl,
//		HomePostsUpdate: blockTime,
//	}
//
//	imagesBase64 := msg.ImagesBase64
//	if len(imagesBase64) > 0 {
//		if len(imagesBase64) > 1 {
//			return nil, fmt.Errorf("only one paid image is allowed to be uploaded")
//		}
//		paidImageBase64 := imagesBase64[0]
//		imageHash := ms.k.sha256Generate(paidImageBase64)
//		setPaidPostImageErr := ms.k.SetPaidPostImage(ctx, imageHash, msg.ImagesBase64[0])
//		if setPaidPostImageErr != nil {
//			return nil, setPaidPostImageErr
//		}
//		post.ImageIds = []string{imageHash}
//	}
//	// post payment
//	ms.k.postPayment(ctx, post)
//	// Store the post in the state
//	ms.k.SetPost(ctx, post)
//	// add home posts
//	ms.addToHomePosts(ctx, post)
//	// add to user created posts
//	ms.addToUserCreatedPosts(ctx, msg.Creator, post)
//
//	ms.k.ProfileKeeper.CheckAndCreateUserHandle(ctx, msg.Creator)
//
//	// mentions add to activitiesReceived
//	userHandleList := msg.Mention
//	if len(userHandleList) > 0 {
//		if len(userHandleList) > 10 {
//			return nil, fmt.Errorf("cannot mention more than 10 users")
//		}
//		for _, userHandle := range userHandleList {
//			address := ms.k.ProfileKeeper.GetAddressByUserHandle(ctx, userHandle)
//			if address != "" {
//				ms.addActivitiesReceived(ctx, post, "", "", msg.Creator, address, profiletypes.ActivitiesType_ACTIVITIES_MENTION)
//			}
//		}
//	}
//
//	topicList := msg.Topic
//	category := msg.Category
//	err = ms.handleCategoryTopicPost(ctx, msg.Creator, topicList, category, blockTime, postId)
//	if err != nil {
//		return nil, err
//	}
//
//	//Emit an event for the creation
//	ctx.EventManager().EmitEvents(sdk.Events{
//		sdk.NewEvent(
//			types.EventTypeCreatePaidPost,
//			sdk.NewAttribute(types.AttributeKeyCreator, msg.Creator),
//			sdk.NewAttribute(types.AttributeKeyPostID, postId),
//			sdk.NewAttribute(types.AttributeKeyTitle, msg.Title),
//			sdk.NewAttribute(types.AttributeKeyTimestamp, fmt.Sprintf("%d", blockTime)),
//		),
//	})
//
//	return &types.MsgCreatePaidPostResponse{PostId: postId}, nil
//}

// QuotePost implements types.MsgServer.
func (ms msgServer) QuotePost(goCtx context.Context, msg *types.MsgQuotePostRequest) (*types.MsgQuotePostResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate the message
	if len(msg.Comment) == 0 {
		return nil, types.NewInvalidRequestError("comment cannot be empty")
	}
	if len(msg.Quote) == 0 {
		return nil, types.NewInvalidRequestError("quote cannot be empty")
	}
	if err := types.ValidatePostContent(msg.Comment); err != nil {
		return nil, err
	}

	// Validate sender address
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, types.NewInvalidAddressErrorf("invalid sender address: %s", err)
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

	parentPost, found := ms.k.GetPost(ctx, msg.Quote)
	if !found {
		return nil, types.NewPostNotFoundError(msg.Quote)
	}
	parentPost.RepostCount += 1
	ms.k.SetPost(ctx, parentPost)

	// mentions add to activitiesReceived
	userHandleList := msg.Mention
	if len(userHandleList) > 0 {
		if err := types.ValidateMentions(userHandleList); err != nil {
			return nil, err
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
	err = ms.handleCategoryTopicPost(ctx, msg.Creator, topicList, category, blockTime, postId)
	if err != nil {
		return nil, err
	}

	//Emit an event for the creation
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeQuotePost,
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
		return nil, types.NewInvalidRequestError("quote cannot be empty")
	}

	// Validate sender address
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, types.NewInvalidAddressErrorf("invalid sender address: %s", err)
	}

	parentPost, found := ms.k.GetPost(ctx, msg.Quote)
	if !found {
		return nil, types.NewPostNotFoundError(msg.Quote)
	}
	// add home posts
	//ms.addToHomePosts(ctx, parentPost)
	// add to user created posts
	ms.addToUserCreatedPosts(ctx, msg.Creator, parentPost)
	ms.k.ProfileKeeper.CheckAndCreateUserHandle(ctx, msg.Creator)

	parentPost.RepostCount += 1
	ms.k.SetPost(ctx, parentPost)

	blockTime := ctx.BlockTime().Unix()
	//Emit an event for the repost
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeRepost,
			sdk.NewAttribute(types.AttributeKeyCreator, msg.Creator),
			sdk.NewAttribute(types.AttributeKeyPostID, msg.Quote),
			sdk.NewAttribute(types.AttributeKeyTimestamp, fmt.Sprintf("%d", blockTime)),
		),
	})

	return &types.MsgRepostResponse{}, nil
}

// LikePost implements types.MsgServer.
func (ms msgServer) Like(goCtx context.Context, msg *types.MsgLikeRequest) (*types.MsgLikeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check if user has already liked this post
	hasLiked := ms.k.HasUserLikedPost(ctx, msg.Sender, msg.Id)
	if hasLiked {
		return nil, types.ErrAlreadyLiked
	}

	ms.k.MarkUserLikedPost(ctx, msg.Sender, msg.Id)

	post, err := ms.getPostWithValidation(ctx, msg.Id)
	if err != nil {
		ms.k.UnmarkUserLikedPost(ctx, msg.Sender, msg.Id)
		return nil, err
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

	// post reward
	ms.k.PostReward(ctx, post)
	// set likes I made
	likesIMade := types.LikesIMade{
		PostId:    post.Id,
		Timestamp: blockTime,
	}
	ms.k.SetLikesIMade(ctx, likesIMade, msg.Sender)

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

// Unlike implements types.MsgServer.
func (ms msgServer) Unlike(goCtx context.Context, msg *types.MsgUnlikeRequest) (*types.MsgUnlikeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	hasLiked := ms.k.HasUserLikedPost(ctx, msg.Sender, msg.Id)
	if !hasLiked {
		return nil, types.NewInvalidRequestError("user has not liked this post")
	}

	post, err := ms.getPostWithValidation(ctx, msg.Id)
	if err != nil {
		return nil, err
	}

	if post.LikeCount > 0 {
		post.LikeCount -= 1
	} else {
		return nil, types.ErrInvalidLikeCount
	}

	ms.k.UnmarkUserLikedPost(ctx, msg.Sender, msg.Id)

	ms.k.SetPost(ctx, post)

	// remove from likes I made
	removeErr := ms.k.RemoveFromLikesIMade(ctx, msg.Sender, msg.Id)
	if removeErr != nil {
		ms.k.MarkUserLikedPost(ctx, msg.Sender, msg.Id)
		return nil, types.WrapErrorf(types.ErrLikesIMadeRemove, "failed to remove likes I made for post %s", msg.Id)
	}

	err2 := ms.k.RemoveFromLikesReceived(ctx, post.Creator, msg.Sender, msg.Id)
	if err2 != nil {
		ms.k.MarkUserLikedPost(ctx, msg.Sender, msg.Id)
		ms.k.SetLikesIMade(ctx, types.LikesIMade{PostId: msg.Id, Timestamp: ctx.BlockTime().Unix()}, msg.Sender)
		//if restoreErr := ms.k.SetLikesIMade(ctx, types.LikesIMade{PostId: msg.Id, Timestamp: ctx.BlockTime().Unix()}, msg.Sender); restoreErr != nil {
		//	types.LogError(ms.k.logger, "restore_likes_i_made", restoreErr, "post_id", msg.Id, "sender", msg.Sender)
		//}
		return nil, types.WrapErrorf(types.ErrLikesReceivedRemove, "failed to remove likes received for post %s", msg.Id)
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
		return nil, types.ErrAlreadySaved
	}
	// Retrieve the post by ID
	post, err := ms.getPostWithValidation(ctx, msg.Id)
	if err != nil {
		return nil, err
	}

	// Check if the post type is COMMENT
	if post.PostType == types.PostType_COMMENT {
		return nil, types.NewInvalidRequestError("cannot save a comment type post")
	}

	blockTime := ctx.BlockTime().Unix()
	post.SaveCount += 1
	post.Timestamp = blockTime
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

	post, err := ms.getPostWithValidation(ctx, msg.Id)
	if err != nil {
		return nil, err
	}

	// remove from saves I made
	err = ms.k.RemoveFromSavesIMade(ctx, msg.Sender, msg.Id)
	if err != nil {
		return nil, types.WrapErrorf(types.ErrSavesIMadeRemove, "failed to remove saves I made for post %s", msg.Id)
	}

	err2 := ms.k.RemoveFromLikesReceived(ctx, post.Creator, msg.Sender, msg.Id)
	if err2 != nil {
		return nil, types.WrapErrorf(types.ErrLikesReceivedRemove, "failed to remove likes received for post %s", msg.Id)
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
		return nil, types.NewInvalidRequestError("parent id cannot be empty")
	}
	if len(msg.Comment) == 0 {
		return nil, types.NewInvalidRequestError("comment cannot be empty")
	}

	// Validate sender address
	if err := ms.validateAddress(msg.Creator); err != nil {
		return nil, err
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

	post, err := ms.getPostWithValidation(ctx, msg.ParentId)
	if err != nil {
		return nil, err
	}

	oldScore := post.Score
	// Score Accumulation
	uintExponent := ms.ScoreAccumulation(ctx, msg.Creator, post, 1)
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
		if err := types.ValidateMentions(userHandleList); err != nil {
			return nil, err
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
			types.EventTypeComment,
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
		types.LogError(ms.k.logger, "updateHomePosts", types.ErrDatabaseOperation, "operation", "GetHomePostsCount")
		return
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
		types.LogError(ms.k.logger, "addToHomePosts", types.ErrDatabaseOperation, "operation", "GetHomePostsCount")
		return
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
		types.LogError(ms.k.logger, "addToUserCreatedPosts", types.ErrDatabaseOperation, "operation", "GetUserCreatedPostsCount", "creator", creator)
		return
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
		types.LogError(ms.k.logger, "addToTopicPosts", types.ErrDatabaseOperation, "operation", "GetTopicPostsCount", "topicHash", topicHash)
		return
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
	if len(topics) > 0 {
		for _, topicHash := range topics {
			ms.k.DeleteFromTopicPostsByTopicAndPostId(ctx, topicHash, post.Id, post.HomePostsUpdate)
			ms.k.SetTopicPosts(ctx, topicHash, post.Id)
			count, b := ms.k.GetTopicPostsCount(ctx, topicHash)
			if !b {
				types.LogError(ms.k.logger, "updateTopicPosts", types.ErrDatabaseOperation, "operation", "GetTopicPostsCount", "topicHash", topicHash)
				return
			}
			if count > types.TopicPostsCount {
				ms.k.DeleteLastPostFromTopicPosts(ctx, topicHash)
				count -= 1
				ms.k.SetTopicPostsCount(ctx, topicHash, count)
			}

			//update topic
			topic, _ := ms.k.GetTopic(ctx, topicHash)

			if topic.Id != "" {
				oldScore := topic.Score
				newScore := oldScore + uintExponent
				oldKeywordsScore := topic.GetTrendingKeywordsScore()
				newKeywordsScore := oldKeywordsScore + uintExponent
				topic.Score = newScore
				topic.LikeCount += 1

				//createTime := topic.CreateTime
				topic = ms.trendingKeywordsUpdate(ctx, topic, oldKeywordsScore, newKeywordsScore)
				//blockTime := ctx.BlockTime().Unix()
				//trendingKeywordsTime := topic.GetTrendingKeywordsTime()
				//hotKeywordsCount, b := ms.k.GetTrendingKeywordsCount(ctx)
				//if !b {
				//	ms.k.Logger().Error("GetTrendingKeywordsCount error")
				//}
				//if trendingKeywordsTime > 0 {
				//	ms.k.deleteFromTrendingKeywords(ctx, topicHash, oldKeywordsScore)
				//	isWithin72 := isWithin72Hours(trendingKeywordsTime, blockTime)
				//	if isWithin72 {
				//		ms.k.addToTrendingKeywords(ctx, topicHash, newKeywordsScore)
				//		topic.TrendingKeywordsScore = newKeywordsScore
				//	} else {
				//		hotKeywordsCount -= 1
				//		topic.TrendingKeywordsScore = 0
				//		topic.TrendingKeywordsTime = 0
				//	}
				//} else {
				//	ms.k.addToTrendingKeywords(ctx, topicHash, newKeywordsScore)
				//	hotKeywordsCount += 1
				//	topic.TrendingKeywordsScore = newKeywordsScore
				//	topic.TrendingKeywordsTime = blockTime
				//}
				//
				//if hotKeywordsCount > types.TrendingKeywordsCount {
				//	earliestTopicHash := ms.k.DeleteLastFromTrendingKeywords(ctx)
				//	hotKeywordsCount -= 1
				//	earliestTopic, _ := ms.k.GetTopic(ctx, earliestTopicHash)
				//
				//	earliestTopic.TrendingKeywordsScore = 0
				//	earliestTopic.TrendingKeywordsTime = 0
				//	ms.k.AddTopic(ctx, earliestTopic)
				//} else {
				//	ms.k.SetTrendingKeywordsCount(ctx, hotKeywordsCount)
				//}
				ms.trendingTopicsUpdate(ctx, topic, oldScore, newScore)

				categoryHash := ms.k.getCategoryByTopicHash(ctx, topicHash)
				if categoryHash != "" {
					ms.k.DeleteFromCategoryTopicsByCategoryAndTopicId(ctx, categoryHash, topicHash, oldScore)
					ms.k.SetCategoryTopics(ctx, categoryHash, topicHash, newScore)
				}

				ms.k.AddTopic(ctx, topic)
			}
		}
	}
}

func (ms msgServer) trendingKeywordsUpdate(ctx sdk.Context, topic types.Topic, oldKeywordsScore uint64, newKeywordsScore uint64) types.Topic {
	blockTime := ctx.BlockTime().Unix()
	trendingKeywordsTime := topic.GetTrendingKeywordsTime()
	trendingKeywordsCount, b := ms.k.GetTrendingKeywordsCount(ctx)
	if !b {
		types.LogError(ms.k.logger, "trendingKeywordsUpdate", types.ErrDatabaseOperation, "operation", "GetTrendingKeywordsCount")
	}
	if trendingKeywordsTime > 0 {
		ms.k.deleteFromTrendingKeywords(ctx, topic.Id, oldKeywordsScore)
		withinSpecifiedHours := isWithinHours(trendingKeywordsTime, blockTime, types.TrendingKeyWordsRetentionHours)
		if withinSpecifiedHours {
			ms.k.addToTrendingKeywords(ctx, topic.Id, newKeywordsScore)
			topic.TrendingKeywordsScore = newKeywordsScore
		} else {
			trendingKeywordsCount -= 1
			topic.TrendingKeywordsScore = 0
			topic.TrendingKeywordsTime = 0
		}
	} else {
		ms.k.addToTrendingKeywords(ctx, topic.Id, newKeywordsScore)
		trendingKeywordsCount += 1
		topic.TrendingKeywordsScore = newKeywordsScore
		topic.TrendingKeywordsTime = blockTime
	}

	if trendingKeywordsCount > types.TrendingKeywordsCount {
		earliestTopicHash := ms.k.DeleteLastFromTrendingKeywords(ctx)
		trendingKeywordsCount -= 1
		earliestTopic, _ := ms.k.GetTopic(ctx, earliestTopicHash)

		earliestTopic.TrendingKeywordsScore = 0
		earliestTopic.TrendingKeywordsTime = 0
		ms.k.AddTopic(ctx, earliestTopic)
	} else {
		ms.k.SetTrendingKeywordsCount(ctx, trendingKeywordsCount)
	}
	return topic
}

func (ms msgServer) trendingTopicsUpdate(ctx sdk.Context, topic types.Topic, oldTopicScore uint64, newTopicScore uint64) {
	blockTime := ctx.BlockTime().Unix()
	createTime := topic.GetCreateTime()
	trendingTopicsCount, b := ms.k.GetTrendingTopicsCount(ctx)
	if !b {
		types.LogError(ms.k.logger, "trending_topics_update", types.ErrDatabaseOperation, "operation", "GetTrendingTopicsCount")
	}

	withinSpecifiedHours := isWithinHours(createTime, blockTime, types.TrendingTopicsRetentionHours)
	isWithinTrendingTopics := ms.k.isWithinTrendingTopics(ctx, topic.Id, oldTopicScore)

	if withinSpecifiedHours {
		if isWithinTrendingTopics {
			ms.k.deleteFromTrendingTopics(ctx, topic.Id, oldTopicScore)
			trendingTopicsCount -= 1
		}
		ms.k.addToTrendingTopics(ctx, topic.Id, newTopicScore)
		trendingTopicsCount += 1
	} else {
		if isWithinTrendingTopics {
			ms.k.deleteFromTrendingTopics(ctx, topic.Id, oldTopicScore)
			trendingTopicsCount -= 1
		}
	}

	if trendingTopicsCount > types.TrendingTopicsCount {
		ms.k.DeleteLastFromTrendingTopics(ctx)
		trendingTopicsCount -= 1
	} else {
		ms.k.SetTrendingTopicsCount(ctx, trendingTopicsCount)
	}
}

func isWithinHours(createTime int64, blockTime int64, hours int64) bool {
	timeDifference := blockTime - createTime
	if timeDifference < 0 {
		return false
	}
	return timeDifference <= hours*3600
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
//			ms.k.Logger().Error("GetActivitiesReceivedCount error")
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
		types.LogError(ms.k.logger, "addActivitiesReceived", types.ErrDatabaseOperation, "operation", "GetActivitiesReceivedCount", "target", target)
		return
	}
	count += 1
	if count > profiletypes.ActivitiesReceivedCount {
		ms.k.ProfileKeeper.DeleteLastActivitiesReceived(sdkCtx, target)
	}
	ms.k.ProfileKeeper.SetActivitiesReceivedCount(sdkCtx, target, count)
}

func (ms msgServer) handleCategoryTopicPost(ctx sdk.Context, creator string, topicList []string, category string, blockTime int64, postId string) error {
	if len(topicList) > 0 {
		if len(topicList) > 10 {
			return types.NewInvalidRequestError("topic list length cannot be larger than 10")
		}
		var hashTopics []string
		for _, topicName := range topicList {
			topicHash := ms.k.sha256Generate(strings.ToLower(topicName))
			categoryDb := ms.k.getCategoryByTopicHash(ctx, topicHash)
			// add topic
			topicExists := ms.k.TopicExists(ctx, topicHash)
			if !topicExists {
				topic := types.Topic{
					Id:         topicHash,
					Name:       topicName,
					Creator:    creator,
					CreateTime: blockTime,
					UpdateTime: blockTime,
				}
				if categoryDb != "" {
					topic.CategoryId = categoryDb
				} else {
					if category != "" {
						categoryHash := ms.k.sha256Generate(category)
						categoryExists := ms.k.CategoryExists(ctx, categoryHash)
						if categoryExists {
							topic.CategoryId = categoryHash
						}
					}
				}
				ms.k.AddTopic(ctx, topic)
			}

			// add topic search
			ms.k.SetTopicSearch(ctx, topicName)
			ms.addToTopicPosts(ctx, topicHash, postId)
			hashTopics = append(hashTopics, topicHash)

			// add category posts
			if categoryDb != "" {
				//ms.k.SetCategorySearch(ctx, categoryDb)
				//categoryHash := ms.k.sha256Generate(categoryDb)
				ms.addToCategoryPosts(ctx, categoryDb, postId)
				ms.k.SetPostCategoryMapping(ctx, categoryDb, postId)
			} else {
				if category != "" {
					categoryHash := ms.k.sha256Generate(category)
					categoryExists := ms.k.CategoryExists(ctx, categoryHash)
					if categoryExists {
						//ms.k.SetCategorySearch(ctx, category)
						ms.addToCategoryPosts(ctx, categoryHash, postId)
						topic, _ := ms.k.GetTopic(ctx, topicHash)
						ms.addToCategoryTopics(ctx, categoryHash, topic)

						ms.k.SetPostCategoryMapping(ctx, categoryHash, postId)
						//ms.k.SetTopicCategoryMapping(ctx, topicHash, category)
						ms.k.SetTopicCategoryMapping(ctx, topicHash, categoryHash)
					} else {
						isUncategorizedTopic := ms.k.IsUncategorizedTopic(ctx, topicHash)
						if !isUncategorizedTopic {
							ms.k.SetUncategorizedTopics(ctx, topicHash)
							count, _ := ms.k.GetUncategorizedTopicsCount(ctx)
							count += 1
							ms.k.SetUncategorizedTopicsCount(ctx, uint64(count))
						}
					}
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
		ms.k.SetPostTopicsMapping(ctx, hashTopics, postId)
	} else {
		// connect category and post
		if category != "" {
			categoryHash := ms.k.sha256Generate(category)
			categoryExists := ms.k.CategoryExists(ctx, categoryHash)
			if categoryExists {
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
		types.LogError(ms.k.logger, "addToCategoryPosts", types.ErrDatabaseOperation, "operation", "GetCategoryPostsCount", "categoryHash", categoryHash)
		return
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
		types.LogError(ms.k.logger, "updateCategoryPosts", types.ErrDatabaseOperation, "operation", "GetCategoryPostsCount", "category", category)
		return
	}
	if count > types.CategoryPostsCount {
		ms.k.DeleteLastPostFromCategoryPosts(ctx, category)
		count -= 1
		ms.k.SetCategoryPostsCount(ctx, category, count)
	}
}
func (ms msgServer) addToCategoryTopics(ctx sdk.Context, categoryHash string, topic types.Topic) {
	ms.k.SetCategoryTopics(ctx, categoryHash, topic.Id, topic.Score)
	count, b := ms.k.GetCategoryTopicsCount(ctx, categoryHash)
	if !b {
		types.LogError(ms.k.logger, "addToCategoryTopics", types.ErrDatabaseOperation, "operation", "GetCategoryTopicsCount", "categoryHash", categoryHash)
		return
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

	// Validate and get parent post
	parentPost, err := ms.getPostWithValidation(ctx, msg.Id)
	if err != nil {
		return nil, err
	}

	// Validate poll exists
	if parentPost.Poll == nil {
		return nil, types.NewInvalidRequestError("post does not contain a poll")
	}

	if blockTime < parentPost.Poll.VotingStart {
		return nil, types.ErrVotingNotStarted
	}
	if blockTime > parentPost.Poll.VotingEnd {
		return nil, types.ErrVotingEnded
	}

	// 原子性检查：在同一个事务中检查和设置投票状态
	_, alreadyVoted := ms.k.GetPoll(ctx, msg.Id, msg.Creator)
	if alreadyVoted {
		return nil, types.ErrAlreadyVoted
	}

	// 立即设置投票状态，防止竞态条件
	ms.k.SetPoll(ctx, msg.Id, msg.Creator, msg.OptionId)

	// 验证选项ID是否有效
	validOptionId := false
	for _, option := range parentPost.Poll.Vote {
		if option.Id == msg.OptionId {
			validOptionId = true
			break
		}
	}
	if !validOptionId {
		// 如果选项ID无效，回滚投票状态
		ms.k.RemovePoll(ctx, msg.Id, msg.Creator)
		return nil, types.NewInvalidRequestError("invalid option ID")
	}

	// 更新投票数量
	parentPost.Poll.TotalVotes += 1
	voteList := parentPost.Poll.Vote
	for i := range voteList {
		if parentPost.Poll.Vote[i].Id == msg.OptionId {
			parentPost.Poll.Vote[i].Count += 1
			break
		}
	}

	// 保存更新后的帖子
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
		}, types.ErrRequestDenied
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
		}, types.ErrRequestDenied
	}
}

// UpdateTopic implements types.MsgServer.
func (ms msgServer) UpdateTopic(goCtx context.Context, msg *types.UpdateTopicRequest) (*types.UpdateTopicResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	//creator := msg.Creator
	//isAdmin := ms.k.ProfileKeeper.IsAdmin(ctx, creator)
	//if isAdmin {
	//	json := msg.TopicJson
	//	id := json.Id
	//	topic, _ := ms.k.GetTopic(ctx, id)
	//	//ms.k.deleteFromHotTopics72(ctx, id, topic.Score)
	//	topic.Score = json.Score
	//	if json.Avatar != "" {
	//		//topic.Avatar = json.Avatar
	//		ms.k.SetTopicAvatar(ctx, id, json.Avatar)
	//	}
	//	ms.k.AddTopic(ctx, topic)
	//	//ms.k.addToHotTopics72(ctx, id, topic.Score)
	//}
	json := msg.TopicJson
	id := json.Id
	topic, _ := ms.k.GetTopic(ctx, id)
	if json.Title != "" {
		topic.Title = json.Title
	}
	if json.Summary != "" {
		topic.Summary = json.Summary
	}
	if json.Score > 0 {
		topic.Score = json.Score
	}
	categoryId := json.CategoryId
	if categoryId != "" && topic.CategoryId == "" {
		topic.CategoryId = categoryId

		ms.k.SetCategoryTopics(ctx, categoryId, id, topic.Score)
		ms.k.SetTopicCategoryMapping(ctx, id, categoryId)
	}
	ms.k.AddTopic(ctx, topic)

	if json.Image != "" {
		ms.k.SetTopicImage(ctx, id, json.Image)
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
				return nil, types.ErrRequestDenied
			}
			topic, _ := ms.k.GetTopic(ctx, msg.TopicId)
			ms.addToCategoryTopics(ctx, msg.CategoryId, topic)
			ms.k.RemoveFromUncategorizedTopics(ctx, msg.TopicId)
			ms.k.SetCategoryOperator(ctx, creator)
			ms.k.SetTopicCategoryMapping(ctx, msg.TopicId, msg.CategoryId)
		} else {
			topic, _ := ms.k.GetTopic(ctx, msg.TopicId)
			ms.addToCategoryTopics(ctx, msg.CategoryId, topic)
			ms.k.RemoveFromUncategorizedTopics(ctx, msg.TopicId)
			ms.k.SetCategoryOperator(ctx, creator)
			ms.k.SetTopicCategoryMapping(ctx, msg.TopicId, msg.CategoryId)
		}
	}
	return &types.ClassifyUncategorizedTopicResponse{Status: true}, nil
}
