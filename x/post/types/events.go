package types

const (
	EventTypeCreateFreePostWithTitle = "create_free_post_with_title"
	EventTypeCreateFreePost          = "create_free_post"
	EventTypeCreatePaidPost          = "create_paid_post"
	EventTypeQuotePost               = "quote_post"
	EventTypeRepost                  = "repost"
	EventTypeComment                 = "comment"
	EventTypeLikePost                = "like_post"
	EventTypeUnlikePost              = "unlike_post"
	EventTypeSavePost                = "save_post"
	EventTypeUnSavePost              = "un_save_post"
	EventTypeAddCategory             = "add_category"
	EventTypeDeleteCategory          = "delete_category"
	EventTypeUpdateTopicCategory     = "update_topic_category"

	AttributeKeyCreator       = "creator"
	AttributeKeyPostID        = "post_id"
	AttributeKeyCommentID     = "comment_id"
	AttributeKeyTitle         = "title"
	AttributeKeyTimestamp     = "timestamp"
	AttributeKeySender        = "sender"
	AttributeKeyCategoryID    = "category_id"
	AttributeKeyTopicID       = "topic_id"
	AttributeKeyOldCategoryID = "old_category_id"
	AttributeKeyNewCategoryID = "new_category_id"
)
