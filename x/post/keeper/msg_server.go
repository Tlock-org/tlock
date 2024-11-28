package keeper

import (
	"context"
	"cosmossdk.io/errors"
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

	// Generate a unique post ID
	//postID := ms.k.generatePostID(ctx)
	postID := msg.PostId

	// Create the post
	post := types.Post{
		Id:        postID,
		PostType:  types.PostType_ARTICLE,
		Title:     msg.Title,
		Content:   msg.Content,
		Creator:   msg.Creator,
		Timestamp: msg.Timestamp,
		ImagesUrl: msg.ImagesUrl,
		VideosUrl: msg.VideosUrl,
	}

	// Store the post in the state
	ms.k.SetPost(ctx, post)
	// post reward
	ms.k.PostReward(ctx, post)

	//Emit an event for the creation
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeCreateFreePostWithTitle,
			sdk.NewAttribute(types.AttributeKeyCreator, msg.Creator),
			sdk.NewAttribute(types.AttributeKeyPostID, postID),
			sdk.NewAttribute(types.AttributeKeyTitle, msg.Title),
			sdk.NewAttribute(types.AttributeKeyTimestamp, fmt.Sprintf("%d", msg.Timestamp)),
		),
	})

	return &types.MsgCreateFreePostWithTitleResponse{PostId: postID}, nil
}

// CreateFreePost implements types.MsgServer.
func (ms msgServer) CreateFreePost(goCtx context.Context, msg *types.MsgCreateFreePost) (*types.MsgCreateFreePostResponse, error) {
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

	// Generate a unique post ID
	//postID := ms.k.generatePostID(ctx)
	postID := msg.PostId

	// Create the post
	post := types.Post{
		Id:        postID,
		PostType:  types.PostType_ORIGINAL,
		Content:   msg.Content,
		Creator:   msg.Creator,
		Timestamp: msg.Timestamp,
		ImagesUrl: msg.ImagesUrl,
		VideosUrl: msg.VideosUrl,
	}

	// Store the post in the state
	ms.k.SetPost(ctx, post)
	// post reward
	ms.k.PostReward(ctx, post)

	//Emit an event for the creation
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeCreateFreePost,
			sdk.NewAttribute(types.AttributeKeyCreator, msg.Creator),
			sdk.NewAttribute(types.AttributeKeyPostID, postID),
			sdk.NewAttribute(types.AttributeKeyTitle, msg.Title),
			sdk.NewAttribute(types.AttributeKeyTimestamp, fmt.Sprintf("%d", msg.Timestamp)),
		),
	})

	return &types.MsgCreateFreePostResponse{PostId: postID}, nil
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

	// Generate a unique post ID
	//postID := ms.k.generatePostID(ctx)
	postID := msg.PostId

	// Create the post
	post := types.Post{
		Id:           postID,
		PostType:     types.PostType_ADVERTISEMENT,
		Content:      msg.Content,
		Creator:      msg.Creator,
		Timestamp:    msg.Timestamp,
		ImagesBase64: msg.ImagesBase64,
		ImagesUrl:    msg.ImagesUrl,
		VideosUrl:    msg.VideosUrl,
	}

	// post payment
	ms.k.postPayment(ctx, post)

	// Store the post in the state
	ms.k.SetPost(ctx, post)

	//Emit an event for the creation
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeCreatePaidPost,
			sdk.NewAttribute(types.AttributeKeyCreator, msg.Creator),
			sdk.NewAttribute(types.AttributeKeyPostID, postID),
			sdk.NewAttribute(types.AttributeKeyTitle, msg.Title),
			sdk.NewAttribute(types.AttributeKeyTimestamp, fmt.Sprintf("%d", msg.Timestamp)),
		),
	})

	return &types.MsgCreatePaidPostResponse{PostId: postID}, nil
}

// LikePost implements types.MsgServer.
func (ms msgServer) LikePost(goCtx context.Context, msg *types.MsgLikePostRequest) (*types.MsgLikePostResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	post, found := ms.k.GetPost(ctx, msg.PostId)
	if !found {
		return nil, errors.Wrap(types.ErrPostNotFound, msg.PostId)
	}
	post.LikeCount += 1

	// update post
	ms.k.SetPost(ctx, post)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeLikePost,
			sdk.NewAttribute(types.AttributeKeyPostID, msg.PostId),
			sdk.NewAttribute(types.AttributeKeySender, msg.Sender),
		),
	})

	return &types.MsgLikePostResponse{Status: true}, nil
}

// UnlikePost implements types.MsgServer.
func (ms msgServer) UnlikePost(goCtx context.Context, msg *types.MsgUnlikePostRequest) (*types.MsgUnlikePostResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	post, found := ms.k.GetPost(ctx, msg.PostId)
	if !found {
		return nil, errors.Wrap(types.ErrPostNotFound, msg.PostId)
	}

	if post.LikeCount > 0 {
		post.LikeCount -= 1
	} else {
		return nil, errors.Wrap(types.ErrInvalidLikeCount, "like count is already zero")
	}

	ms.k.SetPost(ctx, post)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeUnlikePost,
			sdk.NewAttribute(types.AttributeKeyPostID, msg.PostId),
			sdk.NewAttribute(types.AttributeKeySender, msg.Sender),
		),
	})

	return &types.MsgUnlikePostResponse{Status: true}, nil
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

	postID := msg.PostId

	// Create the post
	post := types.Post{
		Id:        postID,
		PostType:  types.PostType_QUOTE,
		Content:   msg.Comment,
		Creator:   msg.Creator,
		Timestamp: msg.Timestamp,
		Quote:     msg.Quote,
	}

	// Store the post in the state
	ms.k.SetPost(ctx, post)
	// post reward
	ms.k.PostReward(ctx, post)

	//Emit an event for the creation
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeCreateFreePostWithTitle,
			sdk.NewAttribute(types.AttributeKeyCreator, msg.Creator),
			sdk.NewAttribute(types.AttributeKeyPostID, postID),
			sdk.NewAttribute(types.AttributeKeyTimestamp, fmt.Sprintf("%d", msg.Timestamp)),
		),
	})

	return &types.MsgQuotePostResponse{}, nil
}
