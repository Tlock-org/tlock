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
					Use:       "get [address]",
					Short:     "Get the profile by address",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "address"},
					},
				},
				{
					RpcMethod: "QueryProfileAvatar",
					Use:       "avatar [address]",
					Short:     "Get the avatar by address",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "address"},
					},
				},
				{
					RpcMethod: "QueryFollowRelationship",
					Use:       "follow-relationship [addressA] [addressB]",
					Short:     "Get follow relationship",
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
					Use:       "following [address]",
					Short:     "Get list of following",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{
							ProtoField: "address",
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
					Use:       "followers [address]",
					Short:     "Get list of followers",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{
							ProtoField: "address",
						},
					},
				},
				{
					RpcMethod: "GetMentionSuggestions",
					Use:       "get-mention-suggestions [address] [matching]",
					Short:     "Get list of mention suggestions",
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
					Use:       "activities-received-count [address]",
					Short:     "Get activities received count",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "address"},
					},
				},
				{
					RpcMethod: "SearchUsers",
					Use:       "search-users [matching]",
					Short:     "Get users by matching",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "matching"},
					},
				},
				{
					RpcMethod: "IsAdmin",
					Use:       "is-admin [address]",
					Short:     "Check if address is admin",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "address"},
					},
				},
				{
					RpcMethod: "QueryMessages",
					Use:       "query-messages [receiver_addr] [sender_addr]",
					Short:     "Query messages",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{
							ProtoField: "receiver_addr",
						},
						{
							ProtoField: "sender_addr",
						},
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
					Short:     "Follow someone",
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
					Short:     "Unfollow someone",
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
					Short:     "Add admin",
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
					Short:     "Remove admin",
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
					Short:     "Manage admin",
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
				{
					RpcMethod: "SendMessage",
					Use:       "send-message [creator] [receiver] [content]",
					Short:     "Send message",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{
							ProtoField: "creator",
							Optional:   false,
						},
						{
							ProtoField: "receiver",
						},
						{
							ProtoField: "content",
						},
					},
				},
			},
		},
	}
}
