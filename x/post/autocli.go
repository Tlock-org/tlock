package module

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	modulev1 "github.com/rollchains/tlock/api/post/v1"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: modulev1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "ResolveName",
					Use:       "resolve [wallet]",
					Short:     "Resolve the name of a wallet address",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "wallet"},
					},
				},
				{
					RpcMethod: "QueryPost",
					Use:       "get [post_id]",
					Short:     "get the post by post_id",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "post_id"},
					},
				},
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Query the current consensus parameters",
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: modulev1.Msg_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "SetServiceName",
					Use:       "set [name]",
					Short:     "Set the mapping to your wallet address",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "name"},
					},
				},
				{
					RpcMethod: "UpdateParams",
					Skip:      false, // set to true if authority gated
				},
				{
					RpcMethod: "GrantAllowanceFromModule",
					Use:       "grant-allowance-from-module [sender] [userAddress]",
					Short:     "grant allowance from module account to user account",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{
							ProtoField: "sender",
						},
						{
							ProtoField: "userAddress",
						},
					},
				},
				{
					RpcMethod: "CreateFreePostWithTitle",
					Use:       "create-free-post-with-title [creator] [postId] [title] [content] [timestamp] [imagesUrl] [videosUrl]",
					Short:     "create free post with title",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{
							ProtoField: "creator",
						},
						{
							ProtoField: "postId",
						},
						{
							ProtoField: "title",
						},
						{
							ProtoField: "content",
						},
						{
							ProtoField: "timestamp",
						},
						{
							ProtoField: "imagesUrl",
						},
						{
							ProtoField: "videosUrl",
						},
					},
				},
				{
					RpcMethod: "CreateFreePost",
					Use:       "create-free-post [creator] [postId] [content] [timestamp] [imagesUrl] [videosUrl]",
					Short:     "create free post",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{
							ProtoField: "creator",
						},
						{
							ProtoField: "postId",
						},
						{
							ProtoField: "content",
						},
						{
							ProtoField: "timestamp",
						},
						{
							ProtoField: "imagesUrl",
						},
						{
							ProtoField: "videosUrl",
						},
					},
				},
				{
					RpcMethod: "CreatePaidPost",
					Use:       "create-paid-post [creator] [postId] [content] [timestamp] [imagesBase64] [imagesUrl] [videosUrl]",
					Short:     "create paid post",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{
							ProtoField: "creator",
						},
						{
							ProtoField: "postId",
						},
						{
							ProtoField: "content",
						},
						{
							ProtoField: "timestamp",
						},
						{
							ProtoField: "imagesBase64",
						},
						{
							ProtoField: "imagesUrl",
						},
						{
							ProtoField: "videosUrl",
						},
					},
				},
				{
					RpcMethod: "LikePost",
					Use:       "like-post [postId] [sender]",
					Short:     "like post",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{
							ProtoField: "postId",
						},
						{
							ProtoField: "sender",
						},
					},
				},
				{
					RpcMethod: "UnlikePost",
					Use:       "unlike-post [postId] [sender]",
					Short:     "unlike post",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{
							ProtoField: "postId",
						},
						{
							ProtoField: "sender",
						},
					},
				},
				{
					RpcMethod: "Quote",
					Use:       "quote [creator] [postId] [quote] [comment] [timestamp]",
					Short:     "quote post",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{
							ProtoField: "creator",
						},
						{
							ProtoField: "postId",
						},
						{
							ProtoField: "quote",
						},
						{
							ProtoField: "comment",
						},
						{
							ProtoField: "timestamp",
						},
					},
				},
			},
		},
	}
}
