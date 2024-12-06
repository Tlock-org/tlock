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
					Use:       "create-free-post-with-title [creator] [title] [content]",
					Short:     "create free post with title",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{
							ProtoField: "creator",
						},
						{
							ProtoField: "title",
						},
						{
							ProtoField: "content",
						},
					},
					FlagOptions: map[string]*autocliv1.FlagOptions{
						"imagesUrl": {
							//Shorthand: "i",
						},
						"videosUrl": {
							//Shorthand: "v",
						},
					},
				},
				{
					RpcMethod: "CreateFreePost",
					Use:       "create-free-post [creator] [content]",
					Short:     "create free post",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{
							ProtoField: "creator",
						},
						{
							ProtoField: "content",
						},
					},
					FlagOptions: map[string]*autocliv1.FlagOptions{
						"imagesUrl": {
							//Shorthand: "i",
						},
						"videosUrl": {
							//Shorthand: "v",
						},
					},
				},
				{
					RpcMethod: "CreateFreePostImagePayable",
					Use:       "create-free-post-image-payable [creator] [content] [imagesBase64]",
					Short:     "create free post image payable",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{
							ProtoField: "creator",
						},
						{
							ProtoField: "content",
						},
						{
							ProtoField: "imagesBase64",
						},
					},
					FlagOptions: map[string]*autocliv1.FlagOptions{
						"imagesUrl": {
							//Shorthand: "i",
						},
						"videosUrl": {
							//Shorthand: "v",
						},
					},
				},
				{
					RpcMethod: "CreatePaidPost",
					Use:       "create-paid-post [creator] [content] [imagesBase64]",
					Short:     "create paid post",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{
							ProtoField: "creator",
						},
						{
							ProtoField: "content",
						},
						{
							ProtoField: "imagesBase64",
						},
					},
					FlagOptions: map[string]*autocliv1.FlagOptions{
						"imagesUrl": {
							//Shorthand: "i",
						},
						"videosUrl": {
							//Shorthand: "v",
						},
					},
				},
				{
					RpcMethod: "QuotePost",
					Use:       "quote-post [creator] [quote] [comment]",
					Short:     "quote post",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{
							ProtoField: "creator",
						},
						{
							ProtoField: "quote",
						},
						{
							ProtoField: "comment",
						},
					},
				},
				{
					RpcMethod: "Like",
					Use:       "like [sender] [id]",
					Short:     "like",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{
							ProtoField: "sender",
						},
						{
							ProtoField: "id",
						},
					},
				},
				{
					RpcMethod: "Unlike",
					Use:       "unlike [sender] [id]",
					Short:     "unlike",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{
							ProtoField: "sender",
						},
						{
							ProtoField: "id",
						},
					},
				},
				{
					RpcMethod: "SavePost",
					Use:       "save-post [sender] [id]",
					Short:     "save post",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{
							ProtoField: "sender",
						},
						{
							ProtoField: "id",
						},
					},
				},
				{
					RpcMethod: "UnSavePost",
					Use:       "un-save-post [sender] [id]",
					Short:     "un save post",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{
							ProtoField: "sender",
						},
						{
							ProtoField: "id",
						},
					},
				},
				{
					RpcMethod: "Comment",
					Use:       "comment [creator] [parentId] [comment]",
					Short:     "comment",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{
							ProtoField: "creator",
						},
						{
							ProtoField: "parentId",
						},
						{
							ProtoField: "comment",
						},
					},
				},
			},
		},
	}
}
