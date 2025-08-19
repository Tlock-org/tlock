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
					Use:       "resolve [address]",
					Short:     "Resolve the name of an address",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "address"},
					},
				},
				{
					RpcMethod: "QueryHomePosts",
					Use:       "home-posts [page_size]",
					Short:     "Query home posts",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{
							ProtoField: "page_size",
						},
					},
				},
				{
					RpcMethod: "QueryFirstPageHomePosts",
					Use:       "first-home-posts [page] [page_size]",
					Short:     "Query first home posts",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{
							ProtoField: "page",
						},
						{
							ProtoField: "page_size",
						},
					},
				},
				{
					RpcMethod: "QueryTopicPosts",
					Use:       "topic-posts [topic_id] [page]",
					Short:     "Query topic posts",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{
							ProtoField: "topic_id",
						},
						{
							ProtoField: "page",
						},
					},
				},
				{
					RpcMethod: "QueryUserCreatedPosts",
					Use:       "user-created-posts [address] [page]",
					Short:     "Query user created posts",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{
							ProtoField: "address",
						},
						{
							ProtoField: "page",
						},
					},
				},
				{
					RpcMethod: "QueryPost",
					Use:       "get [post_id]",
					Short:     "Get the post by post_id",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "post_id"},
					},
				},
				{
					RpcMethod: "SearchTopics",
					Use:       "search-topics [matching]",
					Short:     "Get topics by matching",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "matching"},
					},
				},
				{
					RpcMethod: "QueryComments",
					Use:       "query-comments [id] [page]",
					Short:     "Query comments",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{
							ProtoField: "id",
						},
						{
							ProtoField: "page",
						},
					},
				},
				{
					RpcMethod: "QueryCommentsReceived",
					Use:       "query-comments-received [address] [page]",
					Short:     "Query comments received",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{
							ProtoField: "address",
						},
						{
							ProtoField: "page",
						},
					},
				},
				{
					RpcMethod: "LikesIMade",
					Use:       "likes-i-made [address] [page]",
					Short:     "Query the list of likes made by a specific address",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{
							ProtoField: "address",
						},
						{
							ProtoField: "page",
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
					Use:       "saves-i-made [address] [page]",
					Short:     "Query the list of saves made by a specific address",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{
							ProtoField: "address",
						},
						{
							ProtoField: "page",
						},
					},
				},
				{
					RpcMethod: "LikesReceived",
					Use:       "likes-received [address] [page]",
					Short:     "Query the list of likes received by a specific address",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{
							ProtoField: "address",
						},
						{
							ProtoField: "page",
						},
					},
				},
				{
					RpcMethod: "QueryActivitiesReceived",
					Use:       "activities-received [address] [page]",
					Short:     "Get list of activities received",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{
							ProtoField: "address",
						},
						{
							ProtoField: "page",
						},
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
					Short:     "Query categories",
				},
				{
					RpcMethod: "QueryTopicsByCategory",
					Use:       "category-topics [category_id] [page]",
					Short:     "Get topics by category",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{
							ProtoField: "category_id",
						},
						{
							ProtoField: "page",
						},
					},
				},
				{
					RpcMethod: "QueryCategoryByTopic",
					Use:       "topic-category [topic_id]",
					Short:     "Get category by topic",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "topic_id"},
					},
				},
				{
					RpcMethod: "QueryCategoryPosts",
					Use:       "category-posts [category_id] [page]",
					Short:     "Get posts by category",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{
							ProtoField: "category_id",
						},
						{
							ProtoField: "page",
						},
					},
				},
				{
					RpcMethod: "QueryFollowingTopics",
					Use:       "following-topics [address] [page]",
					Short:     "Get following topics",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{
							ProtoField: "address",
						},
						{
							ProtoField: "page",
						},
					},
				},
				{
					RpcMethod: "QueryIsFollowingTopic",
					Use:       "is-following-topic [address] [topic_id]",
					Short:     "Query is following topic",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{
							ProtoField: "address",
						},
						{
							ProtoField: "topic_id",
						},
					},
				},
				{
					RpcMethod: "QueryUncategorizedTopics",
					Use:       "uncategorized-topics [address] [page] [limit]",
					Short:     "Query uncategorized topics",
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
					RpcMethod: "QueryVoteOption",
					Use:       "vote-option",
					Short:     "Query vote option",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{
							ProtoField: "address",
						},
						{
							ProtoField: "post_id",
						},
					},
				},
				{
					RpcMethod: "QueryTopicImage",
					Use:       "topic-image [topic_id]",
					Short:     "Get the image by topic",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "topic_id"},
					},
				},
				{
					RpcMethod: "QueryTopic",
					Use:       "topic [topic_id]",
					Short:     "Get the topic by id",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "topic_id"},
					},
				},
				{
					RpcMethod: "QueryTrendingKeywords",
					Use:       "trending-keywords [page]",
					Short:     "Get trending keywords",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{
							ProtoField: "page",
						},
					},
				},
				{
					RpcMethod: "QueryTrendingTopics",
					Use:       "trending-topics [page]",
					Short:     "Get trending topics",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{
							ProtoField: "page",
						},
					},
				},
				{
					RpcMethod: "QueryPaidPostImage",
					Use:       "paid-post-image [image_id]",
					Short:     "Get the image by id",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "image_id"},
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
					Short:     "Set the mapping to your address",
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
					Short:     "Grant allowance from module account to user account",
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
					RpcMethod: "CreatePost",
					Use:       "create-post [creator] [post_detail]",
					Short:     "Create post",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{
							ProtoField: "creator",
						},
						{
							ProtoField: "post_detail",
						},
					},
				},
				//{
				//	RpcMethod: "CreateFreePostWithTitle",
				//	Use:       "create-free-post-with-title [creator] [title] [content]",
				//	Short:     "create free post with title",
				//	PositionalArgs: []*autocliv1.PositionalArgDescriptor{
				//		{
				//			ProtoField: "creator",
				//		},
				//		{
				//			ProtoField: "title",
				//		},
				//		{
				//			ProtoField: "content",
				//		},
				//	},
				//	FlagOptions: map[string]*autocliv1.FlagOptions{
				//		"imagesUrl": {
				//			//Shorthand: "i",
				//		},
				//		"videosUrl": {
				//			//Shorthand: "v",
				//		},
				//		"mention": {},
				//		"topic":   {},
				//	},
				//},
				//{
				//	RpcMethod: "CreateFreePost",
				//	Use:       "create-free-post [creator] [content]",
				//	Short:     "create free post",
				//	PositionalArgs: []*autocliv1.PositionalArgDescriptor{
				//		{
				//			ProtoField: "creator",
				//		},
				//		{
				//			ProtoField: "content",
				//		},
				//	},
				//	FlagOptions: map[string]*autocliv1.FlagOptions{
				//		"imagesUrl": {},
				//		"videosUrl": {},
				//		"mention":   {},
				//		"topic":     {},
				//	},
				//},
				//{
				//	RpcMethod: "CreateFreePostImagePayable",
				//	Use:       "create-free-post-image-payable [creator] [content] [imagesBase64]",
				//	Short:     "create free post image payable",
				//	PositionalArgs: []*autocliv1.PositionalArgDescriptor{
				//		{
				//			ProtoField: "creator",
				//		},
				//		{
				//			ProtoField: "content",
				//		},
				//		{
				//			ProtoField: "imagesBase64",
				//		},
				//	},
				//	FlagOptions: map[string]*autocliv1.FlagOptions{
				//		"imagesUrl": {
				//			//Shorthand: "i",
				//		},
				//		"videosUrl": {
				//			//Shorthand: "v",
				//		},
				//		"mention": {},
				//		"topic":   {},
				//	},
				//},
				//{
				//	RpcMethod: "CreatePaidPost",
				//	Use:       "create-paid-post [creator] [content] [imagesBase64]",
				//	Short:     "create paid post",
				//	PositionalArgs: []*autocliv1.PositionalArgDescriptor{
				//		{
				//			ProtoField: "creator",
				//		},
				//		{
				//			ProtoField: "content",
				//		},
				//		{
				//			ProtoField: "imagesBase64",
				//		},
				//	},
				//	FlagOptions: map[string]*autocliv1.FlagOptions{
				//		"imagesUrl": {
				//			//Shorthand: "i",
				//		},
				//		"videosUrl": {
				//			//Shorthand: "v",
				//		},
				//		"mention": {},
				//		"topic":   {},
				//	},
				//},
				{
					RpcMethod: "QuotePost",
					Use:       "quote-post [creator] [quote] [comment]",
					Short:     "Quote post",
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
					Short:     "Repost post",
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
					Short:     "Like post",
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
					Short:     "Unlike post",
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
					Short:     "Save post",
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
					Short:     "Unsave post",
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
					Short:     "Comment post",
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
					Short:     "Add category",
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
					Short:     "Delete category",
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
					Short:     "Update topic",
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
					Short:     "Follow topic",
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
					Short:     "Unfollow topic",
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
					RpcMethod: "ClassifyUncategorizedTopic",
					Use:       "classify-uncategorized-topic [creator] [topic_id] [category_id]",
					Short:     "Classify uncategorized topic",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{
							ProtoField: "creator",
							Optional:   false,
						},
						{
							ProtoField: "topic_id",
						},
						{
							ProtoField: "category_id",
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
