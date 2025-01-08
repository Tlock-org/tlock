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

	KeySomeValue = "some_value"

	PostKeyPrefix = "Post/content/"

	HomePostsKeyPrefix      = "Post/posts/home/"
	CategoryPostsKeyPrefix  = "Post/posts/category/"
	FollowedPostsKeyPrefix  = "Post/posts/followed/"
	HomePostsCountKeyPrefix = "home_posts_count"

	CommentListKeyPrefix = "Post/comment/list/"

	LikesIMadePrefix = "Post/likes/i/made/"

	SavesIMadePrefix = "Post/saves/i/made/"

	LikesReceivedPrefix = "Post/likes/received/"

	HomePostsCount = 100
	PageSize       = 10
)

var ORMModuleSchema = ormv1alpha1.ModuleSchemaDescriptor{
	SchemaFile: []*ormv1alpha1.ModuleSchemaDescriptor_FileEntry{
		{Id: 1, ProtoFileName: "post/v1/state.proto"},
	},
	Prefix: []byte{0},
}
