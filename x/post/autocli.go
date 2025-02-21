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
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Query the current consensus parameters",
				},
				{
					RpcMethod: "ResolveName",
					Use:       "resolve [wallet]",
					Short:     "Resolve the name of a wallet address",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "wallet"},
					},
				},
				{
					RpcMethod: "QueryHomePosts",
					Use:       "home-posts",
					Short:     "query home posts",
				},
				{
					RpcMethod: "QueryFirstPageHomePosts",
					Use:       "first-home-posts",
					Short:     "query first home posts",
				},
				{
					RpcMethod: "QueryTopicPosts",
					Use:       "topic-posts [topic_id]",
					Short:     "query topic posts",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "topic_id"},
					},
				},
				{
					RpcMethod: "QueryUserCreatedPosts",
					Use:       "user-created-posts [wallet]",
					Short:     "query user created posts",
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
					RpcMethod: "SearchTopics",
					Use:       "search-topics [matching]",
					Short:     "get topics by matching",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "matching"},
					},
				},
				{
					RpcMethod: "QueryComments",
					Use:       "query-comments",
					Short:     "query commends",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "id"},
					},
				},
				{
					RpcMethod: "QueryCommentsReceived",
					Use:       "query-comments-received",
					Short:     "query comments received",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "wallet"},
					},
				},
				{
					RpcMethod: "LikesIMade",
					Use:       "likes-i-made [wallet]",
					Short:     "Query the list of likes made by a specific wallet",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{
							ProtoField: "wallet",
						},
					},
					//FlagOptions: map[string]*autocliv1.FlagOptions{
					//	"limit": {
					//		Name:         "limit",
					//		Usage:        "Limit the number of results returned",
					//		DefaultValue: "10",
					//	},
					//	"offset": {
					//		Name:         "offset",
					//		Usage:        "Skip the first N results",
					//		DefaultValue: "0",
					//	},
					//	"count-total": {
					//		Name:         "count-total",
					//		Usage:        "Include the total count of likes made",
					//		DefaultValue: "false",
					//	},
					//	"reverse": {
					//		Name:         "reverse",
					//		Usage:        "Reverse the order of the results",
					//		DefaultValue: "false",
					//	},
					//},
				},
				{
					RpcMethod: "SavesIMade",
					Use:       "saves-i-made [wallet]",
					Short:     "Query the list of save made by a specific wallet",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{
							ProtoField: "wallet",
						},
					},
				},
				{
					RpcMethod: "LikesReceived",
					Use:       "likes-received [wallet]",
					Short:     "Query the list of likes received by a specific wallet",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{
							ProtoField: "wallet",
						},
					},
				},
				{
					RpcMethod: "QueryActivitiesReceived",
					Use:       "activities-received [address]",
					Short:     "get list of activitiesReceived",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "address"},
					},
				},
				//{
				//	RpcMethod: "QueryCategoryExists",
				//	Use:       "category-exists [category]",
				//	Short:     "get category exists",
				//	PositionalArgs: []*autocliv1.PositionalArgDescriptor{
				//		{ProtoField: "category"},
				//	},
				//},
				{
					RpcMethod: "QueryCategories",
					Use:       "categories",
					Short:     "query categories",
				},
				{
					RpcMethod: "QueryTopicsByCategory",
					Use:       "category-topics [category_id]",
					Short:     "get topics by category",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "category_id"},
					},
				},
				{
					RpcMethod: "QueryCategoryPosts",
					Use:       "category-posts [category_id]",
					Short:     "get posts by category",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "category_id"},
					},
				},
				{
					RpcMethod: "QueryHotTopics72",
					Use:       "hot-topics",
					Short:     "get hot topics",
				},
				{
					RpcMethod: "QueryFollowingTopics",
					Use:       "following-topics [address]",
					Short:     "get following topics",
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
						"mention": {},
						"topic":   {},
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
						"imagesUrl": {},
						"videosUrl": {},
						"mention":   {},
						"topic":     {},
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
						"mention": {},
						"topic":   {},
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
						"mention": {},
						"topic":   {},
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
					FlagOptions: map[string]*autocliv1.FlagOptions{
						"mention": {},
						"topic":   {},
					},
				},
				{
					RpcMethod: "Repost",
					Use:       "repost [creator] [quote]",
					Short:     "repost",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{
							ProtoField: "creator",
						},
						{
							ProtoField: "quote",
						},
					},
					FlagOptions: map[string]*autocliv1.FlagOptions{
						"mention": {},
						"topic":   {},
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
					RpcMethod: "UnsavePost",
					Use:       "unsave-post [sender] [id]",
					Short:     "unsave post",
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
				{
					RpcMethod: "AddCategory",
					Use:       "add-category [creator] [params]",
					Short:     "add category",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{
							ProtoField: "creator",
						},
						{
							ProtoField: "params",
						},
					},
				},
				{
					RpcMethod: "DeleteCategory",
					Use:       "delete-category [creator] [id]",
					Short:     "delete category",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{
							ProtoField: "creator",
						},
						{
							ProtoField: "id",
						},
					},
				},
				{
					RpcMethod: "UpdateTopic",
					Use:       "update-topic [creator] [topic_json]",
					Short:     "update topic",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{
							ProtoField: "creator",
							Optional:   false,
						},
						{
							ProtoField: "topic_json",
						},
					},
				},
				{
					RpcMethod: "FollowTopic",
					Use:       "follow-topic [creator] [topic_id]",
					Short:     "follow topic",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{
							ProtoField: "creator",
							Optional:   false,
						},
						{
							ProtoField: "topic_id",
						},
					},
				},
				{
					RpcMethod: "UnfollowTopic",
					Use:       "unfollow-topic [creator] [topic_id]",
					Short:     "unfollow topic",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{
							ProtoField: "creator",
							Optional:   false,
						},
						{
							ProtoField: "topic_id",
						},
					},
				},
				//{
				//	RpcMethod: "Mention",
				//	Use:       "mention [creator] [mention_json]",
				//	Short:     "mention",
				//	PositionalArgs: []*autocliv1.PositionalArgDescriptor{
				//		{
				//			ProtoField: "creator",
				//			Optional:   false,
				//		},
				//		{
				//			ProtoField: "mention_json",
				//		},
				//	},
				//},
			},
		},
	}
}
