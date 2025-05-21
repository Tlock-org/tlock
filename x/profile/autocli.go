package module

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	modulev1 "github.com/rollchains/tlock/api/profile/v1"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: modulev1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Query the current consensus parameters",
				},
				{
					RpcMethod: "QueryProfile",
					Use:       "get [wallet_address]",
					Short:     "get the profile by wallet_address",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "wallet_address"},
					},
				},
				{
					RpcMethod: "QueryProfileAvatar",
					Use:       "avatar [address]",
					Short:     "get the avatar by address",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "address"},
					},
				},
				{
					RpcMethod: "QueryFollowRelationship",
					Use:       "follow-relationship [addressA] [addressB]",
					Short:     "get follow relationship",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{
							ProtoField: "addressA",
						},
						{
							ProtoField: "addressB",
						},
					},
				},
				{
					RpcMethod: "QueryFollowing",
					Use:       "following [wallet_address]",
					Short:     "get list of following",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{
							ProtoField: "wallet_address",
						},
						{
							ProtoField: "page",
						},
						{
							ProtoField: "limit",
						},
					},
				},
				{
					RpcMethod: "QueryFollowers",
					Use:       "followers [wallet_address]",
					Short:     "get list of followers",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{
							ProtoField: "wallet_address",
						},
					},
				},
				{
					RpcMethod: "GetMentionSuggestions",
					Use:       "get-mention-suggestions [address] [matching]",
					Short:     "get list of mention suggestions",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{
							ProtoField: "address",
						},
						{
							ProtoField: "matching",
						},
					},
				},
				//{
				//	RpcMethod: "QueryActivitiesReceived",
				//	Use:       "activitiesReceived [wallet_address]",
				//	Short:     "get list of activitiesReceived",
				//	PositionalArgs: []*autocliv1.PositionalArgDescriptor{
				//		{ProtoField: "wallet_address"},
				//	},
				//},
				{
					RpcMethod: "QueryActivitiesReceivedCount",
					Use:       "activitiesReceivedCount [wallet_address]",
					Short:     "get list of activitiesReceivedCount",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "wallet_address"},
					},
				},
				{
					RpcMethod: "SearchUsers",
					Use:       "searchUsers [matching]",
					Short:     "get users by matching",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "matching"},
					},
				},
				{
					RpcMethod: "IsAdmin",
					Use:       "is-admin [address]",
					Short:     "is admin",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "address"},
					},
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: modulev1.Msg_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "UpdateParams",
					Skip:      false, // set to true if authority gated
				},
				{
					RpcMethod: "AddProfile",
					Use:       "add-profile [creator] [profile_json]",
					Short:     "Add a new profile with optional JSON-formatted options",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{
							ProtoField: "creator",
							Optional:   false,
						},
						{
							ProtoField: "profile_json",
						},
					},
					//FlagOptions: map[string]*autocliv1.FlagOptions{
					//	"nickname": {
					//		Shorthand: "n",
					//	},
					//	"user_handle": {
					//		Shorthand: "u",
					//	},
					//	"avatar": {
					//		Shorthand: "t",
					//	},
					//},
				},
				{
					RpcMethod: "Follow",
					Use:       "follow [creator] [targetAddr]",
					Short:     "follow someone",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{
							ProtoField: "creator",
							Optional:   false,
						},
						{
							ProtoField: "targetAddr",
						},
					},
				},
				{
					RpcMethod: "Unfollow",
					Use:       "unfollow [creator] [targetAddr]",
					Short:     "unfollow someone",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{
							ProtoField: "creator",
							Optional:   false,
						},
						{
							ProtoField: "targetAddr",
						},
					},
				},
				{
					RpcMethod: "AddAdmin",
					Use:       "add-admin [creator] [address]",
					Short:     "add admin",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{
							ProtoField: "creator",
							Optional:   false,
						},
						{
							ProtoField: "address",
						},
					},
				},
				{
					RpcMethod: "RemoveAdmin",
					Use:       "remove-admin [creator] [address]",
					Short:     "remove admin",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{
							ProtoField: "creator",
							Optional:   false,
						},
						{
							ProtoField: "address",
						},
					},
				},
				{
					RpcMethod: "ManageAdmin",
					Use:       "manage-admin [creator] [action] [manage_json]",
					Short:     "manage admin",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{
							ProtoField: "creator",
							Optional:   false,
						},
						{
							ProtoField: "action",
						},
						{
							ProtoField: "manage_json",
						},
					},
				},
			},
		},
	}
}
