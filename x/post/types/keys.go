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

	HomePostsKeyPrefix      = "Post/posts/home/"
	FollowedPostsKeyPrefix  = "Post/posts/followed/"
	HomePostsCountKeyPrefix = "home_posts_count"

	UserCreatedPostsKeyPrefix      = "Post/posts/user/created/"
	UserCreatedPostsCountKeyPrefix = "Post/posts/user/created/count/"
	UserCreatedPostsCount          = 100
	UserCreatedPostsPageSize       = 20

	CommentListKeyPrefix = "Post/comment/list/"

	LikesIMadePrefix = "Post/likes/i/made/"

	SavesIMadePrefix = "Post/saves/i/made/"

	LikesReceivedPrefix    = "Post/likes/received/"
	CommentsReceivedPrefix = "Post/comments/received/"

	HomePostsCount    = 1000
	HomePostsPageSize = 100

	TopicSearchKeyPrefix       = "Post/topic/search/"
	TopicPostsKeyPrefix        = "Post/posts/topic/"
	TopicPostsCountKeyPrefix   = "Post/posts/topic/count/"
	TopicPostsCount            = 1000
	TopicPostsPageSize         = 100
	PostTopicsMappingKeyPrefix = "Post/topics/mapping/"

	CategoryKeyPrefix             = "Post/category/"
	TopicKeyPrefix                = "Post/topic/"
	TopicCategoryMappingKeyPrefix = "Post/topicCategoryMapping/"
	CategoryTopicsKeyPrefix       = "Post/category/topics/"
	HotTopics72KeyPrefix          = "Post/hot/topics/72/"
	HotTopics72CountKeyPrefix     = "Post/hot/topics/72/count/"
	HotTopics72Count              = 1000
	HotTopics72PageSize           = 100

	CategoryPostsKeyPrefix       = "Post/posts/category/"
	CategoryPostsCountKeyPrefix  = "Post/posts/topic/count/"
	CategoryPostsCount           = 10000
	CategoryPostsPageSize        = 100
	CategorySearchKeyPrefix      = "Post/category/search/"
	PostCategoryMappingKeyPrefix = "Post/category/mapping/"

	PollUserPrefix = "Post/poll/"
)

var ORMModuleSchema = ormv1alpha1.ModuleSchemaDescriptor{
	SchemaFile: []*ormv1alpha1.ModuleSchemaDescriptor_FileEntry{
		{Id: 1, ProtoFileName: "post/v1/state.proto"},
	},
	Prefix: []byte{0},
}
