package keeper

import (
	"context"
	"cosmossdk.io/errors"
	"encoding/base64"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/rollchains/tlock/x/post/types"
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
	// Generate a unique post ID
	data := fmt.Sprintf("%s|%s|%s|%d", msg.Creator, msg.Title, msg.Content, blockTime)
	postID := ms.k.generatePostID(data)

	// Create the post
	post := types.Post{
		Id:              postID,
		PostType:        types.PostType_ARTICLE,
		Title:           msg.Title,
		Content:         msg.Content,
		Creator:         msg.Creator,
		Timestamp:       blockTime,
		ImagesUrl:       msg.ImagesUrl,
		VideosUrl:       msg.VideosUrl,
		HomePostsUpdate: blockTime,
	}

	// Store the post in the state
	ms.k.SetPost(ctx, post)
	// post reward
	ms.k.PostReward(ctx, post)
	// add home posts
	ms.addHomePosts(ctx, post)

	//Emit an event for the creation
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeCreateFreePostWithTitle,
			sdk.NewAttribute(types.AttributeKeyCreator, msg.Creator),
			sdk.NewAttribute(types.AttributeKeyPostID, postID),
			sdk.NewAttribute(types.AttributeKeyTitle, msg.Title),
			sdk.NewAttribute(types.AttributeKeyTimestamp, fmt.Sprintf("%d", blockTime)),
		),
	})

	return &types.MsgCreateFreePostWithTitleResponse{PostId: postID}, nil
}

// CreateFreePost implements types.MsgServer.
func (ms msgServer) CreateFreePost(goCtx context.Context, msg *types.MsgCreateFreePost) (*types.MsgCreateFreePostResponse, error) {
	toString := base64.StdEncoding.EncodeToString([]byte("{\"body\":{\"messages\":[{\"@type\":\"/cosmos.bank.v1beta1.MsgSend\",\"from_address\":\"tlock1hj5fveer5cjtn4wd6wstzugjfdxzl0xp5u7j9p\",\"to_address\":\"tlock1efd63aw40lxf3n4mhf7dzhjkr453axurggdkvg\",\"amount\":[{\"denom\":\"uTOK\",\"amount\":\"10\"}]}],\"memo\":\"\",\"timeout_height\":\"0\",\"extension_options\":[],\"non_critical_extension_options\":[]},\"auth_info\":{\"signer_infos\":[{\"public_key\":{\"@type\":\"/cosmos.crypto.secp256k1.PubKey\",\"key\":\"ApZa31BR3NWLylRT6Qi5+f+zXtj2OpqtC76vgkUGLyww\"},\"mode_info\":{\"single\":{\"mode\":\"SIGN_MODE_DIRECT\"}},\"sequence\":\"1\"}],\"fee\":{\"amount\":[{\"denom\":\"uTOK\",\"amount\":\"77886\"}],\"gas_limit\":\"77886\",\"payer\":\"\",\"granter\":\"\"},\"tip\":null},\"signatures\":[\"RWGJJcQ8Ioul52d6HbBW6F1FJwuNPRSTTny6xpAZCfpYgHSIGuk0+uupaC5gx0FKur8qOA9tZlhKhAfLyf9hWg==\"]}"))
	fmt.Printf("=======================toString: %s\n", toString)

	//ms.k.Logger().Error("===============toString: %s\n", toString)

	ms.k.Logger().Warn("============= to create free post ==============")
	ctx := sdk.UnwrapSDKContext(goCtx)

	//bytes := ctx.TxBytes()
	//hash := sha256.Sum256(bytes)
	//fmt.Printf("=======================hash: %s\n", hash)

	// Validate the message
	if len(msg.Content) == 0 {
		return nil, errors.Wrapf(types.ErrInvalidRequest, "Content cannot be empty")
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

	// Generate a unique post ID
	blockTime := ctx.BlockTime().Unix()
	data := fmt.Sprintf("%s|%s|%d", msg.Creator, msg.Content, blockTime)
	postID := ms.k.generatePostID(data)

	// Create the post
	post := types.Post{
		Id:              postID,
		PostType:        types.PostType_ORIGINAL,
		Content:         msg.Content,
		Creator:         msg.Creator,
		Timestamp:       blockTime,
		ImagesUrl:       msg.ImagesUrl,
		VideosUrl:       msg.VideosUrl,
		HomePostsUpdate: blockTime,
	}

	// Store the post in the state
	ms.k.SetPost(ctx, post)
	// post reward
	ms.k.PostReward(ctx, post)
	// add home posts
	ms.addHomePosts(ctx, post)

	ms.k.ProfileKeeper.CheckAndCreateUserHandle(ctx, msg.Creator)

	//Emit an event for the creation
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeCreateFreePost,
			sdk.NewAttribute(types.AttributeKeyCreator, msg.Creator),
			sdk.NewAttribute(types.AttributeKeyPostID, postID),
			sdk.NewAttribute(types.AttributeKeyTimestamp, fmt.Sprintf("%d", blockTime)),
		),
	})
	ms.k.Logger().Warn("============= create free post end ==============")
	return &types.MsgCreateFreePostResponse{PostId: postID}, nil
}

// CreateFreePostImagePayable implements types.MsgServer.
func (ms msgServer) CreateFreePostImagePayable(goCtx context.Context, msg *types.MsgCreateFreePostImagePayable) (*types.MsgCreateFreePostImagePayableResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate the message
	if len(msg.Content) == 0 {
		return nil, errors.Wrapf(types.ErrInvalidRequest, "Content cannot be empty")
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
	postID := ms.k.generatePostID(data)

	// Create the post
	post := types.Post{
		Id:              postID,
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
	ms.addHomePosts(ctx, post)

	ms.k.ProfileKeeper.CheckAndCreateUserHandle(ctx, msg.Creator)
	//Emit an event for the creation
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeCreatePaidPost,
			sdk.NewAttribute(types.AttributeKeyCreator, msg.Creator),
			sdk.NewAttribute(types.AttributeKeyPostID, postID),
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
	postID := ms.k.generatePostID(data)

	// Create the post
	post := types.Post{
		Id:              postID,
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
	ms.addHomePosts(ctx, post)

	ms.k.ProfileKeeper.CheckAndCreateUserHandle(ctx, msg.Creator)

	//Emit an event for the creation
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeCreatePaidPost,
			sdk.NewAttribute(types.AttributeKeyCreator, msg.Creator),
			sdk.NewAttribute(types.AttributeKeyPostID, postID),
			sdk.NewAttribute(types.AttributeKeyTitle, msg.Title),
			sdk.NewAttribute(types.AttributeKeyTimestamp, fmt.Sprintf("%d", blockTime)),
		),
	})

	return &types.MsgCreatePaidPostResponse{PostId: postID}, nil
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

	// Validate sender address
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errors.Wrapf(types.ErrInvalidAddress, "Invalid sender address: %s", err)
	}

	blockTime := ctx.BlockTime().Unix()
	data := fmt.Sprintf("%s|%s|%s|%d", msg.Creator, msg.Quote, msg.Comment, blockTime)
	postID := ms.k.generatePostID(data)

	// Create the post
	post := types.Post{
		Id:              postID,
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

	ms.k.ProfileKeeper.CheckAndCreateUserHandle(ctx, msg.Creator)

	//Emit an event for the creation
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeCreateFreePostWithTitle,
			sdk.NewAttribute(types.AttributeKeyCreator, msg.Creator),
			sdk.NewAttribute(types.AttributeKeyPostID, postID),
			sdk.NewAttribute(types.AttributeKeyTimestamp, fmt.Sprintf("%d", blockTime)),
		),
	})

	return &types.MsgQuotePostResponse{}, nil
}

// LikePost implements types.MsgServer.
func (ms msgServer) Like(goCtx context.Context, msg *types.MsgLikeRequest) (*types.MsgLikeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	post, found := ms.k.GetPost(ctx, msg.Id)
	if !found {
		return nil, errors.Wrap(types.ErrPostNotFound, msg.Id)
	}

	// add home posts
	if post.PostType != types.PostType_COMMENT {
		exist := ms.k.IsPostInHomePosts(ctx, post.Id, post.HomePostsUpdate)
		fmt.Sprintf("==========exist: %s\n", exist)

		if exist {
			ms.addHomePostsExist(ctx, post)
		} else {
			ms.addHomePosts(ctx, post)
		}
	}

	blockTime := ctx.BlockTime().Unix()
	post.LikeCount += 1
	post.HomePostsUpdate = blockTime

	// update post
	ms.k.SetPost(ctx, post)

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
	err := ms.k.RemoveLikesIMade(ctx, msg.Sender, msg.Id)
	if err != nil {
		return nil, errors.Wrap(types.ErrLikesIMadeRemove, msg.Id)
	}

	err2 := ms.k.RemoveLikesReceived(ctx, post.Creator, msg.Sender, msg.Id)
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
	// Retrieve the post by ID
	post, found := ms.k.GetPost(ctx, msg.Id)
	if !found {
		return nil, errors.Wrap(types.ErrPostNotFound, msg.Id)
	}

	// Check if the post type is COMMENT
	if post.PostType == types.PostType_COMMENT {
		return nil, errors.Wrap(types.ErrInvalidPostType, "cannot save a comment type post")
	}

	exist := ms.k.IsPostInHomePosts(ctx, post.Id, post.HomePostsUpdate)
	if exist {
		ms.addHomePostsExist(ctx, post)
	} else {
		ms.addHomePosts(ctx, post)
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

	// set likes received
	likesReceived := types.LikesReceived{
		LikerAddress: msg.Sender,
		PostId:       post.Id,
		LikeType:     types.LikeType_SAVE,
		Timestamp:    blockTime,
	}
	ms.k.SetLikesReceived(ctx, likesReceived, post.Creator)

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
	err := ms.k.RemoveSavesIMade(ctx, msg.Sender, msg.Id)
	if err != nil {
		return nil, errors.Wrap(types.ErrSavesIMadeRemove, msg.Id)
	}

	err2 := ms.k.RemoveLikesReceived(ctx, post.Creator, msg.Sender, msg.Id)
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
	commentID := ms.k.generatePostID(data)

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
	if post.PostType != types.PostType_COMMENT {
		exist := ms.k.IsPostInHomePosts(ctx, post.Id, post.HomePostsUpdate)
		if exist {
			ms.addHomePostsExist(ctx, post)
		} else {
			ms.addHomePosts(ctx, post)
		}
	}

	// update post commentCount
	if !found {
		return nil, errors.Wrap(types.ErrPostNotFound, msg.ParentId)
	}
	post.CommentCount += 1
	post.HomePostsUpdate = blockTime

	// update post
	ms.k.SetPost(ctx, post)

	ms.k.ProfileKeeper.CheckAndCreateUserHandle(ctx, msg.Creator)
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

func (ms msgServer) addHomePostsExist(ctx sdk.Context, post types.Post) {
	ms.k.DeleteHomePostsByPostId(ctx, post.Id, post.HomePostsUpdate)
	ms.k.SetHomePosts(ctx, post.Id)
	count, b := ms.k.GetHomePostsCount(ctx)
	if !b {
		panic("GetHomePostsCount error")
	}
	if count > types.HomePostsCount {
		ms.k.DeleteFirstHomePosts(ctx)
		count -= 1
		ms.k.SetHomePostsCount(ctx, count)
	}
}

func (ms msgServer) addHomePosts(ctx sdk.Context, post types.Post) {
	ms.k.SetHomePosts(ctx, post.Id)
	count, b := ms.k.GetHomePostsCount(ctx)
	if !b {
		panic("GetHomePostsCount error")
	}
	count += 1
	if count > types.HomePostsCount {
		ms.k.DeleteFirstHomePosts(ctx)
	} else {
		ms.k.SetHomePostsCount(ctx, count)
	}
}
