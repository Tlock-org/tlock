package types

import (
	"cosmossdk.io/collections"

	ormv1alpha1 "cosmossdk.io/api/cosmos/orm/v1alpha1"
)

var (
	// ParamsKey saves the current module params.
	ParamsKey = collections.NewPrefix(0)
)

const (
	ModuleName = "post"

	StoreKey = ModuleName

	QuerierRoute = ModuleName

	DenomBase = "uTOK"

	PostKeyPrefix = "Post/content/"

	PageSize = 10

	MaxPostContentLength          = 300
	MaxPostWithTitleContentLength = 10000

	HomePostsKeyPrefix      = "Post/posts/home/"
	FollowedPostsKeyPrefix  = "Post/posts/followed/"
	HomePostsCountKeyPrefix = "home_posts_count"

	UserCreatedPostsKeyPrefix      = "Post/posts/user/created/"
	UserCreatedPostsCountKeyPrefix = "Post/posts/user/created/count/"
	UserCreatedPostsCount          = 100
	UserCreatedPostsPageSize       = 10

	CommentListKeyPrefix = "Post/comment/list/"

	UserLikesPrefix  = "Post/user/likes/"
	LikesIMadePrefix = "Post/likes/i/made/"

	UserSavesPrefix  = "Post/user/saves/"
	SavesIMadePrefix = "Post/saves/i/made/"

	LikesReceivedPrefix    = "Post/likes/received/"
	CommentsReceivedPrefix = "Post/comments/received/"

	HomePostsCount    = 1000
	HomePostsPageSize = 10

	TopicSearchKeyPrefix       = "Post/topic/search/"
	TopicPostsKeyPrefix        = "Post/posts/topic/"
	TopicPostsCountKeyPrefix   = "Post/posts/topic/count/"
	TopicPostsCount            = 1000
	TopicPostsPageSize         = 10
	PostTopicsMappingKeyPrefix = "Post/topics/mapping/"

	CategoryKeyPrefix                 = "Post/category/"
	CategoryWithIndexKeyPrefix        = "Post/category/index/"
	CategoryIndexKeyPrefix            = "Post/category_index"
	TopicKeyPrefix                    = "Post/topic/"
	TopicImagePrefix                  = "Post/topic/image/"
	TopicCategoryMappingKeyPrefix     = "Post/topicCategoryMapping/"
	CategoryTopicsKeyPrefix           = "Post/category/topics/"
	UncategorizedTopicsKeyPrefix      = "Post/uncategorized/topics/"
	UncategorizedTopicsCountKeyPrefix = "Post/uncategorized_topics_count/"
	HotTopics72KeyPrefix              = "Post/hot/topics/72/"
	HotTopics72CountKeyPrefix         = "Post/hot_topics_72_count/"
	HotTopics72Count                  = 1000
	HotTopics72PageSize               = 10
	CategoryTopicsCountKeyPrefix      = "Post/category/topics/count/"
	CategoryTopicsCount               = 10000
	CategoryTopicsPageSize            = 10

	CategoryPostsKeyPrefix       = "Post/posts/category/"
	CategoryPostsCountKeyPrefix  = "Post/posts/topic/count/"
	CategoryPostsCount           = 10000
	CategoryPostsPageSize        = 10
	CategorySearchKeyPrefix      = "Post/category/search/"
	PostCategoryMappingKeyPrefix = "Post/category/mapping/"
	CategoryOperatorKeyPrefix    = "Post/category/operator"

	PollUserPrefix = "Post/poll/"

	FollowTopicPrefix     = "Post/follow/topic/"
	FollowTopicTimePrefix = "Post/follow/topic/time/"
)

var ORMModuleSchema = ormv1alpha1.ModuleSchemaDescriptor{
	SchemaFile: []*ormv1alpha1.ModuleSchemaDescriptor_FileEntry{
		{Id: 1, ProtoFileName: "post/v1/state.proto"},
	},
	Prefix: []byte{0},
}
