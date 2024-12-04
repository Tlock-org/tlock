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
			},
		},
	}
}
