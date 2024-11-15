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
					RpcMethod: "CreatePost",
					Use:       "create [postId] [title] [content] [sender] [timestamp]",
					Short:     "create a new post",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
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
							ProtoField: "sender",
						},
						{
							ProtoField: "timestamp",
						},
					},
				},
				{
					RpcMethod: "UpdateParams",
					Skip:      false, // set to true if authority gated
				},
				{
					RpcMethod: "SetApprove",
					Use:       "set [userAddress]",
					Short:     "Set Approve",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "Address"},
					},
				},
			},
		},
	}
}
