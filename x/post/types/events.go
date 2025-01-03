package types

const (
	EventTypeCreateFreePostWithTitle = "create_free_post_with_title"
	EventTypeCreateFreePost          = "create_free_post"
	EventTypeCreatePaidPost          = "create_paid_post"

	AttributeKeyCreator   = "creator"
	AttributeKeyPostID    = "post_id"
	AttributeKeyCommentID = "comment_id"
	AttributeKeyTitle     = "title"
	AttributeKeyTimestamp = "timestamp"

	EventTypeLikePost   = "like_post"
	EventTypeUnlikePost = "unlike_post"
	EventTypeSavePost   = "save_post"
	EventTypeUnSavePost = "un_save_post"
	AttributeKeySender  = "sender"
)
