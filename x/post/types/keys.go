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
	CategoryPostsKeyPrefix  = "Post/posts/category/"
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

	TopicPostsKeyPrefix       = "Post/posts/topic/"
	TopicPostsCountKeyPrefix  = "Post/posts/topic/count/"
	TopicPostsCount           = 1000
	TopicPostsPageSize        = 100
	TopicPostsBelongKeyPrefix = "Post/posts/topic/belong/"
)

var ORMModuleSchema = ormv1alpha1.ModuleSchemaDescriptor{
	SchemaFile: []*ormv1alpha1.ModuleSchemaDescriptor_FileEntry{
		{Id: 1, ProtoFileName: "post/v1/state.proto"},
	},
	Prefix: []byte{0},
}
