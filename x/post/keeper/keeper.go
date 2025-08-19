package keeper

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"

	"cosmossdk.io/store/prefix"
	"cosmossdk.io/x/feegrant"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	feegrantkeeper "cosmossdk.io/x/feegrant/keeper"
	profileKeeper "github.com/rollchains/tlock/x/profile/keeper"

	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"cosmossdk.io/orm/model/ormdb"

	kvtypes "cosmossdk.io/store/types"
	apiv1 "github.com/rollchains/tlock/api/post/v1"
	"github.com/rollchains/tlock/x/post/types"

	"encoding/binary"

	sdkmath "cosmossdk.io/math"
)

type Keeper struct {
	cdc codec.BinaryCodec

	logger log.Logger

	// state management
	Schema collections.Schema
	Params collections.Item[types.Params]
	OrmDB  apiv1.StateStore

	NameMapping    collections.Map[string, string]
	storeKey       kvtypes.StoreKey
	AccountKeeper  authkeeper.AccountKeeper
	bankKeeper     bankkeeper.Keeper
	FeeGrantKeeper feegrantkeeper.Keeper
	ProfileKeeper  profileKeeper.Keeper

	authority string
}

// NewKeeper creates a new Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey kvtypes.StoreKey,
	storeService storetypes.KVStoreService,
	logger log.Logger,
	authority string,
	ak authkeeper.AccountKeeper,
	bk bankkeeper.Keeper,
	fk feegrantkeeper.Keeper,
	pk profileKeeper.Keeper,
) Keeper {
	logger = logger.With(log.ModuleKey, "x/"+types.ModuleName)

	sb := collections.NewSchemaBuilder(storeService)

	if authority == "" {
		authority = authtypes.NewModuleAddress(govtypes.ModuleName).String()
	}

	db, err := ormdb.NewModuleDB(&types.ORMModuleSchema, ormdb.ModuleDBOptions{KVStoreService: storeService})
	if err != nil {
		panic(err)
	}

	store, err := apiv1.NewStateStore(db)
	if err != nil {
		panic(err)
	}

	k := Keeper{
		cdc:    cdc,
		logger: logger,

		Params: collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),
		OrmDB:  store,

		NameMapping:    collections.NewMap(sb, collections.NewPrefix(1), "name_mapping", collections.StringKey, collections.StringValue),
		storeKey:       storeKey,
		AccountKeeper:  ak,
		bankKeeper:     bk,
		FeeGrantKeeper: fk,
		ProfileKeeper:  pk,
		authority:      authority,
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}

	k.Schema = schema

	return k
}

func (k Keeper) Logger() log.Logger {
	return k.logger
}

// InitGenesis initializes the module's state from a genesis state.
func (k *Keeper) InitGenesis(ctx context.Context, data *types.GenesisState) error {
	if err := data.Params.Validate(); err != nil {
		return err
	}

	return k.Params.Set(ctx, data.Params)
}

// ExportGenesis exports the module's state to a genesis state.
func (k *Keeper) ExportGenesis(ctx context.Context) *types.GenesisState {
	params, err := k.Params.Get(ctx)
	if err != nil {
		panic(err)
	}

	return &types.GenesisState{
		Params: params,
	}
}

func (k Keeper) EncodeBlockTime(ctx sdk.Context) []byte {
	blockTime := ctx.BlockTime().Unix()
	bzBlockTime := make([]byte, 8)
	binary.BigEndian.PutUint64(bzBlockTime, uint64(blockTime))
	return bzBlockTime
}

func (k Keeper) EncodeScore(score uint64) []byte {
	bzScore := make([]byte, 8)
	binary.BigEndian.PutUint64(bzScore, score)
	return bzScore
}

// SetPost stores a post in the state.
func (k Keeper) SetPost(ctx sdk.Context, post types.Post) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.PostKeyPrefix))
	bz := k.cdc.MustMarshal(&post)
	store.Set([]byte(post.Id), bz)

}

func (k Keeper) SetHomePostsCount(ctx sdk.Context, count int64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.HomePostsCountKeyPrefix))

	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, uint64(count))
	store.Set([]byte("count"), bz)
}

func (k Keeper) GetHomePostsCount(ctx sdk.Context) (int64, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.HomePostsCountKeyPrefix))

	bz := store.Get([]byte("count"))
	if bz == nil {
		return 0, false // 返回false表示数据不存在
	}

	count := int64(binary.BigEndian.Uint64(bz))
	return count, true // 返回true表示数据存在
}

func (k Keeper) SetHomePosts(ctx sdk.Context, postId string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.HomePostsKeyPrefix))
	blockTime := k.EncodeBlockTime(ctx)
	key := append(blockTime, []byte(postId)...)
	store.Set(key, []byte(postId))
}

func (k Keeper) IsPostInHomePosts(ctx sdk.Context, postId string, homePostsUpdate int64) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.HomePostsKeyPrefix))
	bzBlockTime := make([]byte, 8)
	binary.BigEndian.PutUint64(bzBlockTime, uint64(homePostsUpdate))
	key := append(bzBlockTime, []byte(postId)...)
	return store.Has(key)
}

func (k Keeper) GetHomePosts(ctx sdk.Context, req *types.QueryHomePostsRequest) ([]string, *query.PageResponse, uint64, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.HomePostsKeyPrefix))
	homePostsCount, _ := k.GetHomePostsCount(ctx)
	if homePostsCount == 0 {
		return []string{}, &query.PageResponse{
			NextKey: nil,
			Total:   uint64(homePostsCount),
		}, uint64(0), nil
	}

	//pageSize := req.PageSize

	var pageSize int64 = types.HomePostsPageSize
	if req.PageSize > 0 {
		pageSize = int64(req.PageSize)
	}
	totalPages := homePostsCount / pageSize
	if homePostsCount%pageSize != 0 {
		totalPages += 1
	}

	if totalPages == 0 {
		totalPages = 1
	}

	currentTime := time.Now()
	unixMilli := currentTime.Unix()
	lastTwoDigits := unixMilli % 100

	pageIndex := lastTwoDigits % totalPages
	// 00-99  10000 00:1-100 01:101-200 02:201-300
	//first := lastTwoDigits * 100
	first := pageIndex * pageSize

	const totalPosts = types.HomePostsCount
	if first >= totalPosts {
		return nil, nil, uint64(0), types.NewInvalidRequestErrorf("offset %d exceeds total number of home posts %d", first, totalPosts)
	}

	pagination := &query.PageRequest{}

	pagination.Limit = uint64(pageSize)
	pagination.Offset = uint64(first)
	pagination.Reverse = true

	var postIDs []string

	pageRes, err := query.Paginate(store, pagination, func(key, value []byte) error {
		postIDs = append(postIDs, string(value))
		return nil
	})

	if err != nil {
		types.LogError(k.logger, "get_home_posts", err, "page_index", pageIndex, "page_size", pageSize)
		return nil, nil, uint64(0), types.WrapError(types.ErrDatabaseOperation, "failed to paginate home posts")
	}
	return postIDs, pageRes, uint64(pageIndex + 1), nil
}

func (k Keeper) GetFirstPageHomePosts(ctx sdk.Context, req *types.QueryFirstPageHomePostsRequest) ([]string, *query.PageResponse, uint64, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.HomePostsKeyPrefix))
	homePostsCount, _ := k.GetHomePostsCount(ctx)
	if homePostsCount == 0 {
		return []string{}, &query.PageResponse{
			NextKey: nil,
			Total:   uint64(homePostsCount),
		}, uint64(0), nil
	}
	page := req.GetPage()
	if page < 1 {
		page = 1
	}

	//const pageSize = types.HomePostsPageSize
	var pageSize int64 = types.HomePostsPageSize
	if req.PageSize > 0 {
		pageSize = int64(req.PageSize)
	}
	totalPages := homePostsCount / pageSize
	if homePostsCount%pageSize != 0 {
		totalPages += 1
	}

	if totalPages == 0 {
		totalPages = 1
	}
	pagination := &query.PageRequest{}

	first := (page - 1) * uint64(pageSize)

	pagination.Limit = uint64(pageSize)
	pagination.Offset = first
	pagination.Reverse = true

	var postIDs []string

	pageRes, err := query.Paginate(store, pagination, func(key, value []byte) error {
		postIDs = append(postIDs, string(value))
		return nil
	})

	if err != nil {
		types.LogError(k.logger, "get_first_page_home_posts", err, "page", page, "page_size", pageSize)
		return nil, nil, uint64(0), types.WrapError(types.ErrDatabaseOperation, "failed to paginate first page home posts")
	}
	return postIDs, pageRes, page, nil
}

func (k Keeper) DeleteLastPostFromHomePosts(ctx sdk.Context) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.HomePostsKeyPrefix))

	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	if !iterator.Valid() {
		types.LogError(k.logger, "delete_last_home_post", types.ErrResourceNotFound, "message", "No posts found to delete from home posts")
		return
	}

	earliestKey := iterator.Key()
	earliestPostID := string(iterator.Value())

	store.Delete(earliestKey)
	k.logger.Info("Deleted earliest home post", "post_id", earliestPostID)
}

func (k Keeper) DeleteFromHomePostsByPostId(ctx sdk.Context, postId string, homePostsUpdate int64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.HomePostsKeyPrefix))
	bzBlockTime := make([]byte, 8)
	binary.BigEndian.PutUint64(bzBlockTime, uint64(homePostsUpdate))
	key := append(bzBlockTime, []byte(postId)...)
	store.Delete(key)
}

func (k Keeper) SetPostTopicsMapping(ctx sdk.Context, topics []string, postId string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.PostTopicsMappingKeyPrefix))
	key := append([]byte(postId))
	value, _ := json.Marshal(topics)
	store.Set(key, value)
}
func (k Keeper) GetTopicsByPostId(ctx sdk.Context, postId string) []string {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.PostTopicsMappingKeyPrefix))
	key := append([]byte(postId))
	bz := store.Get(key)
	if bz == nil {
		return nil
	}
	var topics []string
	err := json.Unmarshal(bz, &topics)
	if err != nil {
		return nil
	}
	return topics
}
func (k Keeper) IsPostInTopics(ctx sdk.Context, postId string) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.PostTopicsMappingKeyPrefix))
	key := append([]byte(postId))
	return store.Has(key)
}

func (k Keeper) SetTopicSearch(ctx sdk.Context, topic string) {
	topicLower := strings.ToLower(topic)
	topicHash := k.sha256Generate(topicLower)
	key := fmt.Sprintf("%s%s%s%s", types.TopicSearchKeyPrefix, topicLower, ":", topicHash)
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte{})
	if !store.Has([]byte(key)) {
		store.Set([]byte(key), []byte(topic))
	}
}

func (k Keeper) SearchTopicMatches(ctx sdk.Context, query string) ([]string, error) {
	queryLower := strings.ToLower(query)
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte{})

	searchPrefix := fmt.Sprintf("%s%s", types.TopicSearchKeyPrefix, queryLower)

	startKey := []byte(searchPrefix)
	endKey := []byte(fmt.Sprintf("%s\xFF", searchPrefix))

	iterator := store.Iterator(startKey, endKey)
	defer iterator.Close()

	var matchedTopics []string
	limit := 10
	for ; iterator.Valid() && len(matchedTopics) < limit; iterator.Next() {
		key := string(iterator.Key())
		topicWithHash := strings.TrimPrefix(key, types.TopicSearchKeyPrefix)
		parts := strings.SplitN(topicWithHash, ":", 2)
		if len(parts) != 2 {
			continue
		}
		topic := parts[0]
		if strings.HasPrefix(topic, queryLower) {
			originalTopic := string(iterator.Value())
			matchedTopics = append(matchedTopics, originalTopic)
		}
	}

	return matchedTopics, nil
}

func (k Keeper) GetTopicPosts(ctx sdk.Context, topic string, page uint64) ([]string, *query.PageResponse, uint64, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.TopicPostsKeyPrefix+topic))

	topicPostsCount, _ := k.GetTopicPostsCount(ctx, topic)
	if topicPostsCount == 0 {
		return []string{}, &query.PageResponse{
			NextKey: nil,
			Total:   uint64(topicPostsCount),
		}, uint64(0), nil
	}

	if page < 1 {
		page = 1
	}

	const pageSize = types.TopicPostsPageSize
	pagination := &query.PageRequest{}

	first := (page - 1) * pageSize

	pagination.Limit = pageSize
	pagination.Offset = first
	pagination.Reverse = true

	var postIDs []string

	pageRes, err := query.Paginate(store, pagination, func(key, value []byte) error {
		postIDs = append(postIDs, string(value))
		return nil
	})

	if err != nil {
		types.LogError(k.logger, "get_topic_posts", err, "topic", topic, "page", page, "page_size", pageSize)
		return nil, nil, uint64(0), types.WrapError(types.ErrDatabaseOperation, "failed to paginate topic posts")
	}
	return postIDs, pageRes, page, nil
}

func (k Keeper) SetTopicPosts(ctx sdk.Context, topic string, postId string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.TopicPostsKeyPrefix+topic))
	blockTime := k.EncodeBlockTime(ctx)
	key := append(blockTime, []byte(postId)...)
	store.Set(key, []byte(postId))
}

func (k Keeper) DeleteFromTopicPostsByTopicAndPostId(ctx sdk.Context, topic string, postId string, homePostsUpdate int64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.TopicPostsKeyPrefix+topic))
	bzBlockTime := make([]byte, 8)
	binary.BigEndian.PutUint64(bzBlockTime, uint64(homePostsUpdate))
	key := append(bzBlockTime, []byte(postId)...)
	store.Delete(key)
}
func (k Keeper) DeleteLastPostFromTopicPosts(ctx sdk.Context, topic string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.TopicPostsKeyPrefix+topic))
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	if !iterator.Valid() {
		types.LogError(k.logger, "delete_last_topic_post", types.ErrResourceNotFound, "topic", topic, "message", "topicPostsCount exceeds 1000 but no posts found")
		return
	}
	earliestKey := iterator.Key()
	earliestPostID := string(iterator.Value())

	store.Delete(earliestKey)
	k.logger.Info("Deleted earliest topic posts", "post_id", earliestPostID)
}
func (k Keeper) SetTopicPostsCount(ctx sdk.Context, topic string, count int64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.TopicPostsCountKeyPrefix))
	key := append([]byte(topic))
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, uint64(count))
	store.Set(key, bz)
}

func (k Keeper) GetTopicPostsCount(ctx sdk.Context, topic string) (int64, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.TopicPostsCountKeyPrefix))
	key := append([]byte(topic))
	bz := store.Get(key)
	if bz == nil {
		return 0, true
	}

	count := int64(binary.BigEndian.Uint64(bz))
	return count, true
}

func (k Keeper) SetUserCreatedPostsCount(ctx sdk.Context, creator string, count int64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.UserCreatedPostsCountKeyPrefix))
	key := append([]byte(creator))
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, uint64(count))
	store.Set(key, bz)
}

func (k Keeper) GetUserCreatedPostsCount(ctx sdk.Context, creator string) (int64, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.UserCreatedPostsCountKeyPrefix))
	key := append([]byte(creator))
	bz := store.Get(key)
	if bz == nil {
		return 0, true
	}
	count := int64(binary.BigEndian.Uint64(bz))
	return count, true
}

func (k Keeper) AddToUserCreatedPosts(ctx sdk.Context, creator string, postId string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.UserCreatedPostsKeyPrefix+creator))
	blockTime := k.EncodeBlockTime(ctx)
	key := append(blockTime, []byte(postId)...)
	store.Set(key, []byte(postId))
}

func (k Keeper) DeleteLastPostFromUserCreated(ctx sdk.Context, creator string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.UserCreatedPostsKeyPrefix+creator))
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()
	if !iterator.Valid() {
		types.LogError(k.logger, "delete_last_user_post", types.ErrResourceNotFound, "creator", creator, "message", "userCreatedPostsCount exceeds 1000 but no posts found")
		return
	}
	earliestKey := iterator.Key()
	earliestPostID := string(iterator.Value())
	store.Delete(earliestKey)
	k.logger.Info("Deleted earliest user created post", "post_id", earliestPostID)
}

func (k Keeper) GetUserCreatedPosts(ctx sdk.Context, creator string, page uint64) ([]string, *query.PageResponse, uint64, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.UserCreatedPostsKeyPrefix+creator))

	userCreatedPostsCount, _ := k.GetUserCreatedPostsCount(ctx, creator)
	if userCreatedPostsCount == 0 {
		return []string{}, &query.PageResponse{
			NextKey: nil,
			Total:   uint64(userCreatedPostsCount),
		}, uint64(0), nil
	}

	const pageSize = types.UserCreatedPostsPageSize
	totalPages := userCreatedPostsCount / pageSize
	if userCreatedPostsCount%pageSize != 0 {
		totalPages += 1
	}

	if totalPages == 0 {
		totalPages = 1
	}

	if page < 1 {
		page = 1
	}
	//currentTime := time.Now()
	//unixMilli := currentTime.Unix()
	//lastTwoDigits := unixMilli % 10
	//
	//pageIndex := lastTwoDigits % totalPages
	//first := pageIndex * pageSize
	first := (page - 1) * pageSize

	const totalPosts = types.UserCreatedPostsCount
	if first >= totalPosts {
		return nil, nil, uint64(0), types.NewInvalidRequestErrorf("offset %d exceeds total number of home posts %d", first, totalPosts)
	}

	pagination := &query.PageRequest{}

	pagination.Limit = pageSize
	pagination.Offset = first
	pagination.Reverse = true

	var postIDs []string

	pageRes, err := query.Paginate(store, pagination, func(key, value []byte) error {
		postIDs = append(postIDs, string(value))
		return nil
	})

	if err != nil {
		types.LogError(k.logger, "get_user_created_posts", err, "creator", creator, "page", page, "page_size", pageSize)
		return nil, nil, uint64(0), types.WrapError(types.ErrDatabaseOperation, "failed to paginate user created posts")
	}
	return postIDs, pageRes, page, nil
}

func (k Keeper) GetLastPostsByAddress(ctx sdk.Context, address string, limit int) ([]string, *query.PageResponse, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.UserCreatedPostsKeyPrefix+address))
	userCreatedPostsCount, _ := k.GetUserCreatedPostsCount(ctx, address)
	if userCreatedPostsCount == 0 {
		return []string{}, &query.PageResponse{
			NextKey: nil,
			Total:   uint64(userCreatedPostsCount),
		}, nil
	}
	pagination := &query.PageRequest{}
	pagination.Limit = uint64(limit)
	pagination.Offset = uint64(0)
	pagination.Reverse = true
	var postIDs []string
	pageRes, err := query.Paginate(store, pagination, func(key, value []byte) error {
		postIDs = append(postIDs, string(value))
		return nil
	})
	if err != nil {
		types.LogError(k.logger, "get_last_posts_by_address", err, "address", address, "limit", limit)
		return nil, nil, types.WrapError(types.ErrDatabaseOperation, "failed to get last posts by address")
	}
	return postIDs, pageRes, nil
}

func (k Keeper) AddToCommentList(ctx sdk.Context, postId string, commentId string, score uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.CommentListKeyPrefix))
	bzScore := k.EncodeScore(score)
	//key := append(append([]byte(postId), bzScore...), []byte(commentId)...)
	var buffer bytes.Buffer
	buffer.WriteString(postId)
	buffer.Write(bzScore)
	buffer.WriteString(commentId)
	key := buffer.Bytes()
	store.Set(key, []byte(commentId))
}

func (k Keeper) GetCommentsByParentId(ctx sdk.Context, parentId string, page uint64) ([]string, *query.PageResponse, uint64, error) {
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * types.PageSize
	pageRequest := &query.PageRequest{
		Offset:     offset,
		Limit:      types.PageSize,
		CountTotal: true,
		Reverse:    true,
	}
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.CommentListKeyPrefix))
	commentStore := prefix.NewStore(store, []byte(parentId))

	var ids []string

	pageResponse, err := query.Paginate(commentStore, pageRequest, func(key []byte, value []byte) error {
		id := string(value)
		ids = append(ids, id)
		return nil
	})

	if err != nil {
		types.LogError(k.logger, "get_comments_by_parent_id", err, "parent_id", parentId, "page", page, "page_size", types.PageSize)
		return nil, nil, uint64(0), types.WrapError(types.ErrDatabaseOperation, "failed to paginate comments by parent id")
	}
	return ids, pageResponse, page, nil
}

func (k Keeper) DeleteFromCommentList(ctx sdk.Context, postId string, commentId string, score uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.CommentListKeyPrefix))
	bzScore := k.EncodeScore(score)
	var buffer bytes.Buffer
	buffer.WriteString(postId)
	buffer.Write(bzScore)
	buffer.WriteString(commentId)
	key := buffer.Bytes()
	store.Delete(key)
}

func (k Keeper) SetCommentsReceived(ctx sdk.Context, creator string, commentId string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.CommentsReceivedPrefix+creator+"/"))
	blockTime := k.EncodeBlockTime(ctx)
	key := append(blockTime, []byte(commentId)...)
	store.Set(key, []byte(commentId))
}

func (k Keeper) GetCommentsReceived(ctx sdk.Context, creator string, page uint64) ([]string, *query.PageResponse, uint64, error) {
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * types.PageSize
	pageRequest := &query.PageRequest{
		Offset:     offset,
		Limit:      types.PageSize,
		CountTotal: true,
		Reverse:    true,
	}
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.CommentsReceivedPrefix+creator+"/"))

	var list []string

	pageResponse, err := query.Paginate(store, pageRequest, func(key []byte, value []byte) error {
		id := string(value)
		list = append(list, id)
		return nil
	})

	if err != nil {
		types.LogError(k.logger, "get_comments_received", err, "creator", creator, "page", page, "page_size", types.PageSize)
		return nil, nil, uint64(0), types.WrapError(types.ErrDatabaseOperation, "failed to paginate comments received")
	}
	return list, pageResponse, page, nil
}

func (k Keeper) SetFollowedPosts(ctx sdk.Context, postId string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.FollowedPostsKeyPrefix))
	blockTime := k.EncodeBlockTime(ctx)
	key := append(blockTime, []byte(postId)...)
	store.Set([]byte(key), []byte(postId))
}
func (k Keeper) GetFollowedPosts(ctx sdk.Context, pagination *query.PageRequest) ([]string, *query.PageResponse, error) {
	return nil, nil, nil
}

func (k Keeper) MarkUserLikedPost(ctx sdk.Context, sender, postId string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.UserLikesPrefix+sender+"/"))
	blockTime := k.EncodeBlockTime(ctx)
	key := []byte(postId)
	store.Set(key, blockTime)
}

// UnmarkUserLikedPost removes the user's like record for a specific post
func (k Keeper) UnmarkUserLikedPost(ctx sdk.Context, sender, postId string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.UserLikesPrefix+sender+"/"))
	key := []byte(postId)
	store.Delete(key)
}

func (k Keeper) HasUserLikedPost(ctx sdk.Context, sender, postId string) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.UserLikesPrefix+sender+"/"))
	key := []byte(postId)
	return store.Has(key)
}

func (k Keeper) SetLikesIMade(ctx sdk.Context, likesIMade types.LikesIMade, sender string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.LikesIMadePrefix+sender+"/"))
	blockTime := k.EncodeBlockTime(ctx)
	key := append(blockTime, []byte(likesIMade.PostId)...)
	bz := k.cdc.MustMarshal(&likesIMade)
	store.Set(key, bz)
}

// GetLikesIMade retrieves the list of likes made by a specific sender, ordered by blockTime in descending order.
func (k Keeper) GetLikesIMade(ctx sdk.Context, sender string, page uint64) ([]*types.LikesIMade, *query.PageResponse, uint64, error) {
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * types.PageSize
	pageRequest := &query.PageRequest{
		Offset:     offset,
		Limit:      types.PageSize,
		CountTotal: true,
		Reverse:    true,
	}
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.LikesIMadePrefix+sender+"/"))
	var likesList []*types.LikesIMade

	pageResponse, err := query.Paginate(store, pageRequest, func(key []byte, value []byte) error {
		var like types.LikesIMade
		if err := k.cdc.Unmarshal(value, &like); err != nil {
			types.LogError(k.logger, "unmarshal_likes_i_made", err, "sender", sender, "post_id", like.PostId)
			return types.WrapError(types.ErrDatabaseOperation, "failed to unmarshal LikesIMade")
		}
		likesList = append(likesList, &like)
		return nil
	})

	if err != nil {
		types.LogError(k.logger, "get_likes_i_made", err, "sender", sender, "page", page, "page_size", types.PageSize)
		return nil, nil, uint64(0), types.WrapError(types.ErrDatabaseOperation, "failed to paginate likes I made")
	}
	return likesList, pageResponse, page, nil
}

func (k Keeper) GetLikesIMadePaginated(ctx sdk.Context, sender string, offset, limit int) ([]types.LikesIMade, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.LikesIMadePrefix+sender+"/"))
	iterator := store.ReverseIterator(nil, nil)
	defer iterator.Close()

	var likesList []types.LikesIMade
	currentIndex := 0

	for ; iterator.Valid() && len(likesList) < limit; iterator.Next() {
		if currentIndex < offset {
			currentIndex++
			continue
		}

		var like types.LikesIMade
		err := k.cdc.Unmarshal(iterator.Value(), &like)
		if err != nil {
			types.LogError(k.logger, "unmarshal_likes_i_made", err, "sender", sender, "post_id", like.PostId)
			return nil, types.WrapError(types.ErrDatabaseOperation, "failed to unmarshal LikesIMade")
		}
		likesList = append(likesList, like)
		currentIndex++
	}

	return likesList, nil
}

func (k Keeper) RemoveFromLikesIMade(ctx sdk.Context, sender string, postId string) error {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.LikesIMadePrefix+sender+"/"))
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var like types.LikesIMade
		err := k.cdc.Unmarshal(iterator.Value(), &like)
		if err != nil {
			types.LogError(k.logger, "unmarshal_likes_i_made", err, "sender", sender, "post_id", postId)
			return types.WrapError(types.ErrDatabaseOperation, "failed to unmarshal LikesIMade")
		}

		if like.PostId == postId {
			key := iterator.Key()
			store.Delete(key)
			return nil
		}
	}
	return types.NewPostNotFoundError(postId)
}

func (k Keeper) MarkUserSavedPost(ctx sdk.Context, sender, postId string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.UserSavesPrefix+sender+"/"))
	blockTime := k.EncodeBlockTime(ctx)
	key := []byte(postId)
	store.Set(key, blockTime)
}
func (k Keeper) HasUserSavedPost(ctx sdk.Context, sender, postId string) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.UserSavesPrefix+sender+"/"))
	key := []byte(postId)
	return store.Has(key)
}

func (k Keeper) SetSavesIMade(ctx sdk.Context, likesIMade types.LikesIMade, sender string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.SavesIMadePrefix+sender+"/"))
	//blockTime := ctx.BlockTime().Unix()
	//key := fmt.Sprintf("%d-%s", blockTime, likesIMade.PostId)
	blockTime := k.EncodeBlockTime(ctx)
	key := append(blockTime, []byte(likesIMade.PostId)...)
	bz := k.cdc.MustMarshal(&likesIMade)
	store.Set(key, bz)
}

func (k Keeper) RemoveFromSavesIMade(ctx sdk.Context, sender string, postId string) error {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.SavesIMadePrefix+sender+"/"))
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var like types.LikesIMade
		err := k.cdc.Unmarshal(iterator.Value(), &like)
		if err != nil {
			types.LogError(k.logger, "unmarshal_saves_i_made", err, "sender", sender, "post_id", postId)
			return types.WrapError(types.ErrDatabaseOperation, "failed to unmarshal SavesIMade")
		}

		if like.PostId == postId {
			key := iterator.Key()
			store.Delete(key)
			return nil
		}
	}
	return types.NewPostNotFoundError(postId)
}

func (k Keeper) GetSavesIMade(ctx sdk.Context, sender string, page uint64) ([]*types.LikesIMade, *query.PageResponse, uint64, error) {

	if page < 1 {
		page = 1
	}
	offset := (page - 1) * types.PageSize
	pageRequest := &query.PageRequest{
		Offset:     offset,
		Limit:      types.PageSize,
		CountTotal: true,
		Reverse:    true,
	}
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.SavesIMadePrefix+sender+"/"))
	var savesList []*types.LikesIMade

	pageResponse, err := query.Paginate(store, pageRequest, func(key []byte, value []byte) error {
		var save types.LikesIMade
		if err := k.cdc.Unmarshal(value, &save); err != nil {
			return err
		}
		savesList = append(savesList, &save)
		return nil
	})

	if err != nil {
		return nil, nil, uint64(0), err
	}
	return savesList, pageResponse, page, nil
}

func (k Keeper) SetLikesReceived(ctx sdk.Context, likesReceived types.LikesReceived, creator string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.LikesReceivedPrefix+creator+"/"))
	//blockTime := ctx.BlockTime().Unix()
	//key := fmt.Sprintf("%d", blockTime)
	blockTime := k.EncodeBlockTime(ctx)
	key := append(blockTime)
	bz := k.cdc.MustMarshal(&likesReceived)
	store.Set(key, bz)
}

func (k Keeper) GetLikesReceived(ctx sdk.Context, creator string, page uint64) ([]*types.LikesReceived, *query.PageResponse, uint64, error) {
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * types.PageSize
	pageRequest := &query.PageRequest{
		Offset:     offset,
		Limit:      types.PageSize,
		CountTotal: true,
		Reverse:    true,
	}

	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.LikesReceivedPrefix+creator+"/"))
	var list []*types.LikesReceived

	pageResponse, err := query.Paginate(store, pageRequest, func(key []byte, value []byte) error {
		var received types.LikesReceived
		if err := k.cdc.Unmarshal(value, &received); err != nil {
			return err
		}
		list = append(list, &received)
		return nil
	})

	if err != nil {
		return nil, nil, uint64(0), err
	}
	return list, pageResponse, page, nil
}

func (k Keeper) RemoveFromLikesReceived(ctx sdk.Context, creator string, sender string, postId string) error {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.LikesReceivedPrefix+creator+"/"))
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var received types.LikesReceived
		err := k.cdc.Unmarshal(iterator.Value(), &received)
		if err != nil {
			types.LogError(k.logger, "unmarshal_likes_received", err, "creator", creator, "sender", sender, "post_id", postId)
			return types.WrapError(types.ErrDatabaseOperation, "failed to unmarshal LikesReceived")
		}

		if received.PostId == postId && received.LikerAddress == sender {
			key := iterator.Key()
			store.Delete(key)
			return nil
		}
	}
	return types.NewPostNotFoundError(postId)
}

func (k Keeper) PostReward(ctx sdk.Context, post types.Post) {
	metadata, found := k.bankKeeper.GetDenomMetaData(ctx, types.DenomBase)
	if !found {
		return
	}
	var exponent uint32
	for _, unit := range metadata.DenomUnits {
		if unit.Denom == metadata.Display {
			exponent = unit.Exponent
			break
		}
	}
	if exponent == 0 {
		exponent = 10000000
	}
	baseBig := big.NewInt(10)
	expBig := big.NewInt(int64(exponent))
	powerBig := new(big.Int).Exp(baseBig, expBig, nil)
	exchangeRate := sdkmath.NewIntFromBigInt(powerBig)
	rewardAmount := sdkmath.NewInt(10).Mul(exchangeRate)
	// send post reward
	amount := sdk.NewCoins(sdk.NewCoin(types.DenomBase, rewardAmount))
	userAddr, err := sdk.AccAddressFromBech32(post.Creator)

	if err != nil {
		return
	}

	err1 := k.SendCoinsFromModuleToAccount(ctx, userAddr, amount)
	if err1 != nil {
		return
	}
}

func (k Keeper) postPayment(ctx sdk.Context, post types.Post) {
	// send post payment
	amount := sdk.NewCoins(sdk.NewCoin(types.DenomBase, sdkmath.NewInt(10)))
	userAddr, err := sdk.AccAddressFromBech32(post.Creator)

	if err != nil {
		return
	}

	err1 := k.SendCoinsFromAccountToModule(ctx, userAddr, amount)
	if err1 != nil {
		return
	}
}

// GetPost retrieves a post by ID from the state.
func (k Keeper) GetPost(ctx sdk.Context, id string) (types.Post, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.PostKeyPrefix))
	bz := store.Get([]byte(id))
	if bz == nil {
		return types.Post{}, false
	}

	var post types.Post
	k.cdc.MustUnmarshal(bz, &post)
	return post, true
}

// generatePostID generates a unique post ID.
// This is a simple implementation using block time and some randomness.
// Consider using a more robust method in production.
func (k Keeper) sha256Generate(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

//func GeneratePostID(data string) string {
//	//data := userAddress + content + blockTime.Format(time.RFC3339Nano)
//	hash := sha256.Sum256([]byte(data))
//	return hex.EncodeToString(hash[:])
//}

// get post module address
func (k Keeper) GetModuleAccountAddress() sdk.AccAddress {
	return k.AccountKeeper.GetModuleAddress(types.ModuleName)
}

// transfer from module to user
func (k Keeper) SendCoinsFromModuleToAccount(ctx sdk.Context, userAddr sdk.AccAddress, amount sdk.Coins) error {
	//moduleAddress := k.AccountKeeper.GetModuleAddress(types.ModuleName)
	//fmt.Printf("==============types.ModuleName: [%s], =moduleAddress: [%s]", types.ModuleName, moduleAddress)
	return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, userAddr, amount)
}

func (k Keeper) SendCoinsFromAccountToModule(ctx sdk.Context, userAddr sdk.AccAddress, amount sdk.Coins) error {
	//moduleAddress := k.AccountKeeper.GetModuleAddress(types.ModuleName)
	//fmt.Printf("==============types.ModuleName: [%s], =moduleAddress: [%s]", types.ModuleName, moduleAddress)
	return k.bankKeeper.SendCoinsFromAccountToModule(ctx, userAddr, types.ModuleName, amount)
}

func (k Keeper) GrantPeriodicAllowance(ctx sdk.Context, sender sdk.AccAddress, userAddr sdk.AccAddress) {

	// Define a specific address
	//specificAddress := sdk.MustAccAddressFromBech32("tlock1vj037yzhdslqzens3lu5vfjay9y8n956gqwyqw")
	specificAddress := sdk.MustAccAddressFromBech32("tlock1hj5fveer5cjtn4wd6wstzugjfdxzl0xp5u7j9p")

	granter := k.AccountKeeper.GetModuleAddress(types.ModuleName)
	grantee := userAddr

	if sender.Equals(specificAddress) {
		now := ctx.BlockTime()
		expiration := now.Add(5 * 365 * 24 * time.Hour)

		period := 24 * time.Hour
		totalSpendLimit := sdk.NewCoins(sdk.NewCoin("uTOK", sdkmath.NewInt(20000000000)))
		spendLimit := sdk.NewCoins(sdk.NewCoin("uTOK", sdkmath.NewInt(10000000)))
		// create a basic allowance
		basicAllowance := feegrant.BasicAllowance{
			SpendLimit: totalSpendLimit,
			Expiration: &expiration,
		}

		// create a periodic allowance
		periodicAllowance := &feegrant.PeriodicAllowance{
			Basic:            basicAllowance,
			Period:           period,
			PeriodSpendLimit: spendLimit,
			PeriodCanSpend:   spendLimit,
			PeriodReset:      now.Add(period),
		}

		err := k.FeeGrantKeeper.GrantAllowance(ctx, granter, grantee, periodicAllowance)
		if err != nil {
			types.LogError(k.logger, "grant_allowance_failed", err, "granter", granter.String(), "grantee", grantee.String())
			return
		}
	} else {
		k.logger.Error("Failed to grant allowance - unknown error", "granter", granter.String(), "grantee", grantee.String())
		return
	}

}

func (k Keeper) SetCategoryPosts(ctx sdk.Context, category string, postId string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.CategoryPostsKeyPrefix))
	blockTime := k.EncodeBlockTime(ctx)
	key := append(append([]byte(category), blockTime...), []byte(postId)...)
	store.Set(key, []byte(postId))
}
func (k Keeper) SetCategoryPostsCount(ctx sdk.Context, category string, count int64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.CategoryPostsCountKeyPrefix))
	key := append([]byte(category))
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, uint64(count))
	store.Set(key, bz)
}
func (k Keeper) DeleteFromCategoryPostsByCategoryAndPostId(ctx sdk.Context, category string, postId string, homePostsUpdate int64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.CategoryPostsKeyPrefix+category))
	bzBlockTime := make([]byte, 8)
	binary.BigEndian.PutUint64(bzBlockTime, uint64(homePostsUpdate))
	key := append(bzBlockTime, []byte(postId)...)
	store.Delete(key)
}
func (k Keeper) DeleteLastPostFromCategoryPosts(ctx sdk.Context, categoryHash string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.CategoryPostsKeyPrefix+categoryHash))
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()
	earliestKey := iterator.Key()
	store.Delete(earliestKey)
}
func (k Keeper) GetCategoryPosts(ctx sdk.Context, categoryHash string, page uint64) ([]string, *query.PageResponse, uint64, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.CategoryPostsKeyPrefix+categoryHash))

	categoryPostsCount, _ := k.GetTopicPostsCount(ctx, categoryHash)
	if categoryPostsCount == 0 {
		return []string{}, &query.PageResponse{
			NextKey: nil,
			Total:   uint64(categoryPostsCount),
		}, uint64(0), nil
	}

	const pageSize = types.CategoryPostsPageSize
	totalPages := categoryPostsCount / pageSize
	if categoryPostsCount%pageSize != 0 {
		totalPages += 1
	}

	if totalPages == 0 {
		totalPages = 1
	}

	//currentTime := time.Now()
	//unixMilli := currentTime.Unix()
	//lastTwoDigits := unixMilli % 100
	//
	//pageIndex := lastTwoDigits % totalPages
	// 00-99  10000 00:1-100 01:101-200 02:201-300
	//first := lastTwoDigits * 100
	//first := pageIndex * pageSize
	if page < 1 {
		page = 1
	}
	first := (page - 1) * pageSize
	const totalPosts = types.CategoryPostsCount
	if first >= totalPosts {
		return nil, nil, uint64(0), fmt.Errorf("offset exceeds total number of category posts")
	}

	pagination := &query.PageRequest{}

	pagination.Limit = pageSize
	pagination.Offset = first
	pagination.Reverse = true

	var postIDs []string

	pageRes, err := query.Paginate(store, pagination, func(key, value []byte) error {
		postIDs = append(postIDs, string(value))
		return nil
	})

	if err != nil {
		return nil, nil, uint64(0), err
	}
	return postIDs, pageRes, page, nil
}
func (k Keeper) GetCategoryPostsCount(ctx sdk.Context, category string) (int64, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.CategoryPostsCountKeyPrefix))
	key := append([]byte(category))
	bz := store.Get(key)
	if bz == nil {
		return 0, true
	}
	count := int64(binary.BigEndian.Uint64(bz))
	return count, true
}

//	func (k Keeper) SetCategorySearch(ctx sdk.Context, category string) {
//		categoryLower := strings.ToLower(category)
//		categoryHash := k.sha256Generate(category)
//		key := fmt.Sprintf("%s%s%s%s", types.CategorySearchKeyPrefix, categoryLower, ":", categoryHash)
//		store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte{})
//		if !store.Has([]byte(key)) {
//			store.Set([]byte(key), []byte(category))
//		}
//	}
func (k Keeper) SetPostCategoryMapping(ctx sdk.Context, category string, postId string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.PostCategoryMappingKeyPrefix))
	key := append([]byte(postId))
	store.Set(key, []byte(category))
}
func (k Keeper) IsPostInCategory(ctx sdk.Context, postId string) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.PostCategoryMappingKeyPrefix))
	key := append([]byte(postId))
	return store.Has(key)
}
func (k Keeper) GetCategoryByPostId(ctx sdk.Context, postId string) string {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.PostCategoryMappingKeyPrefix))
	key := append([]byte(postId))
	bz := store.Get(key)
	if bz == nil {
		return ""
	}
	return string(bz)
}

func (k Keeper) AddCategory(ctx sdk.Context, category types.Category) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.CategoryKeyPrefix))
	var buffer bytes.Buffer
	//bzIndex := k.EncodeScore(category.Index)
	//buffer.Write(bzIndex)
	buffer.WriteString(category.Id)
	key := buffer.Bytes()
	bz := k.cdc.MustMarshal(&category)
	// Set the category in the store
	store.Set(key, bz)
}
func (k Keeper) AddCategoryWithIndex(ctx sdk.Context, category types.Category) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.CategoryWithIndexKeyPrefix))
	var buffer bytes.Buffer
	bzIndex := k.EncodeScore(category.Index)
	buffer.Write(bzIndex)
	buffer.WriteString(category.Id)
	key := buffer.Bytes()
	bz := k.cdc.MustMarshal(&category)
	// Set the category in the store
	store.Set(key, bz)
}

func (k Keeper) DeleteCategory(ctx sdk.Context, categoryHash string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.CategoryKeyPrefix))
	key := []byte(categoryHash)
	store.Delete(key)
}
func (k Keeper) GetCategory(ctx sdk.Context, categoryHash string) types.Category {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.CategoryKeyPrefix))
	key := []byte(categoryHash)
	bz := store.Get(key)
	var category types.Category
	k.cdc.MustUnmarshal(bz, &category)
	return category
}
func (k Keeper) CategoryExists(ctx sdk.Context, categoryHash string) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.CategoryKeyPrefix))
	key := []byte(categoryHash)
	return store.Has(key)
}

func (k Keeper) GetAllCategories(ctx sdk.Context) []types.Category {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.CategoryWithIndexKeyPrefix))

	// Create an iterator to iterate over all keys in the category store
	iterator := store.Iterator(nil, nil) // Iterate over the entire prefix store
	defer iterator.Close()

	var categories []types.Category

	// Iterate over each key-value pair in the store
	for ; iterator.Valid(); iterator.Next() {
		var category types.Category
		// Unmarshal the binary data into the Category struct
		err := k.cdc.Unmarshal(iterator.Value(), &category)
		if err != nil {
			// If unmarshalling fails, panic with an error message
			// Consider handling the error more gracefully in production
			panic(fmt.Sprintf("failed to unmarshal category with key %s: %v", iterator.Key(), err))
		}
		// Append the unmarshalled category to the slice
		categories = append(categories, category)
	}

	return categories
}

func (k Keeper) SetCategoryIndex(ctx sdk.Context, index uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.CategoryIndexKeyPrefix))
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, index)
	store.Set([]byte("index"), bz)
}

func (k Keeper) GetCategoryIndex(ctx sdk.Context) (uint64, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.CategoryIndexKeyPrefix))

	bz := store.Get([]byte("index"))
	if bz == nil {
		return 0, true
	}

	index := binary.BigEndian.Uint64(bz)
	return index, true
}

func (k Keeper) AddTopic(ctx sdk.Context, topic types.Topic) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.TopicKeyPrefix))
	key := []byte(topic.Id)
	bz := k.cdc.MustMarshal(&topic)
	store.Set(key, bz)
}
func (k Keeper) GetTopic(ctx sdk.Context, topicHash string) (types.Topic, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.TopicKeyPrefix))
	// Use the topic ID as the key
	key := []byte(topicHash)
	// Get the value from the store
	bz := store.Get(key)
	if bz == nil {
		// Topic not found
		return types.Topic{}, false
	}
	var topic types.Topic
	// Unmarshal the binary data into the Topic struct
	err := k.cdc.Unmarshal(bz, &topic)
	if err != nil {
		// If unmarshalling fails, log an error and return false
		panic(fmt.Sprintf("Failed to unmarshal topic with topicHash %s: %v", topicHash, err))
	}
	return topic, true
}

func (k Keeper) TopicExists(ctx sdk.Context, topicHash string) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.TopicKeyPrefix))
	key := []byte(topicHash)
	return store.Has(key)
}

func (k Keeper) addToTrendingKeywords(ctx sdk.Context, topicHash string, score uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.TrendingKeywordsPrefix))
	bzScore := k.EncodeScore(score)
	var buffer bytes.Buffer
	buffer.Write(bzScore)
	buffer.WriteString(topicHash)
	key := buffer.Bytes()
	store.Set(key, []byte(topicHash))
}
func (k Keeper) deleteFromTrendingKeywords(ctx sdk.Context, topicHash string, score uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.TrendingKeywordsPrefix))
	bzScore := k.EncodeScore(score)
	var buffer bytes.Buffer
	buffer.Write(bzScore)
	buffer.WriteString(topicHash)
	key := buffer.Bytes()
	store.Delete(key)
}
func (k Keeper) DeleteLastFromTrendingKeywords(ctx sdk.Context) string {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.TrendingKeywordsPrefix))
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()
	if !iterator.Valid() {
		panic("trendingKeywords exceeds 1000 but no found")
	}
	earliestKey := iterator.Key()
	bz := store.Get(earliestKey)
	store.Delete(earliestKey)
	return string(bz)
}
func (k Keeper) GetTrendingKeywords(ctx sdk.Context, page uint64) ([]string, *query.PageResponse, uint64, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.TrendingKeywordsPrefix))

	trendingKeywordsCount, _ := k.GetTrendingKeywordsCount(ctx)
	if trendingKeywordsCount == 0 {
		return []string{}, &query.PageResponse{
			NextKey: nil,
			Total:   uint64(trendingKeywordsCount),
		}, uint64(0), nil
	}

	if page < 1 {
		page = 1
	}
	const pageSize = types.TrendingKeywordsPageSize
	totalPages := trendingKeywordsCount / pageSize
	if trendingKeywordsCount%pageSize != 0 {
		totalPages += 1
	}

	if totalPages == 0 {
		totalPages = 1
	}

	//currentTime := time.Now()
	//unixMilli := currentTime.Unix()
	//lastTwoDigits := unixMilli % 100

	//pageIndex := lastTwoDigits % totalPages
	// 00-99  10000 00:1-100 01:101-200 02:201-300
	first := (page - 1) * pageSize

	const totalTopics = types.TrendingKeywordsCount
	if first >= totalTopics {
		return nil, nil, uint64(0), fmt.Errorf("offset exceeds total number of trending keywords")
	}

	pagination := &query.PageRequest{}

	pagination.Limit = pageSize
	pagination.Offset = first
	pagination.Reverse = true

	var topicIds []string

	pageRes, err := query.Paginate(store, pagination, func(key, value []byte) error {
		topicIds = append(topicIds, string(value))
		return nil
	})
	if err != nil {
		return nil, nil, uint64(0), err
	}
	return topicIds, pageRes, page, nil
}
func (k Keeper) SetTrendingKeywordsCount(ctx sdk.Context, count int64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.TrendingKeywordsCountPrefix))
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, uint64(count))
	store.Set([]byte("count"), bz)
}
func (k Keeper) GetTrendingKeywordsCount(ctx sdk.Context) (int64, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.TrendingKeywordsCountPrefix))
	bz := store.Get([]byte("count"))
	if bz == nil {
		return 0, true
	}
	count := int64(binary.BigEndian.Uint64(bz))
	return count, true
}

func (k Keeper) addToTrendingTopics(ctx sdk.Context, topicHash string, score uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.TrendingTopicsPrefix))
	bzScore := k.EncodeScore(score)
	var buffer bytes.Buffer
	buffer.Write(bzScore)
	buffer.WriteString(topicHash)
	key := buffer.Bytes()
	store.Set(key, []byte(topicHash))
}
func (k Keeper) isWithinTrendingTopics(ctx sdk.Context, topicHash string, score uint64) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.TrendingTopicsPrefix))
	bzScore := k.EncodeScore(score)
	var buffer bytes.Buffer
	buffer.Write(bzScore)
	buffer.WriteString(topicHash)
	key := buffer.Bytes()
	return store.Has(key)
}
func (k Keeper) deleteFromTrendingTopics(ctx sdk.Context, topicHash string, score uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.TrendingTopicsPrefix))
	bzScore := k.EncodeScore(score)
	var buffer bytes.Buffer
	buffer.Write(bzScore)
	buffer.WriteString(topicHash)
	key := buffer.Bytes()
	store.Delete(key)
}
func (k Keeper) DeleteLastFromTrendingTopics(ctx sdk.Context) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.TrendingTopicsPrefix))
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()
	if !iterator.Valid() {
		panic("trendingTopicsCount exceeds 1000 but no found")
	}
	earliestKey := iterator.Key()
	store.Delete(earliestKey)
}
func (k Keeper) GetTrendingTopics(ctx sdk.Context, page uint64) ([]string, *query.PageResponse, uint64, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.TrendingTopicsPrefix))
	trendingTopicsCount, _ := k.GetTrendingTopicsCount(ctx)
	if trendingTopicsCount == 0 {
		return []string{}, &query.PageResponse{
			NextKey: nil,
			Total:   uint64(trendingTopicsCount),
		}, uint64(0), nil
	}

	if page < 1 {
		page = 1
	}
	const pageSize = types.TrendingTopicsPageSize
	totalPages := trendingTopicsCount / pageSize
	if trendingTopicsCount%pageSize != 0 {
		totalPages += 1
	}

	if totalPages == 0 {
		totalPages = 1
	}

	//currentTime := time.Now()
	//unixMilli := currentTime.Unix()
	//lastTwoDigits := unixMilli % 100

	//pageIndex := lastTwoDigits % totalPages
	// 00-99  10000 00:1-100 01:101-200 02:201-300
	first := (page - 1) * pageSize

	const totalTopics = types.TrendingTopicsCount
	if first >= totalTopics {
		return nil, nil, uint64(0), fmt.Errorf("offset exceeds total number of trending topics")
	}

	pagination := &query.PageRequest{}

	pagination.Limit = pageSize
	pagination.Offset = first
	pagination.Reverse = true

	var topicIds []string

	pageRes, err := query.Paginate(store, pagination, func(key, value []byte) error {
		topicIds = append(topicIds, string(value))
		return nil
	})
	if err != nil {
		return nil, nil, uint64(0), err
	}
	return topicIds, pageRes, page, nil
}
func (k Keeper) SetTrendingTopicsCount(ctx sdk.Context, count int64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.TrendingTopicsCountPrefix))
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, uint64(count))
	store.Set([]byte("count"), bz)
}
func (k Keeper) GetTrendingTopicsCount(ctx sdk.Context) (int64, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.TrendingTopicsCountPrefix))
	bz := store.Get([]byte("count"))
	if bz == nil {
		return 0, true
	}
	count := int64(binary.BigEndian.Uint64(bz))
	return count, true
}

func (k Keeper) SetTopicCategoryMapping(ctx sdk.Context, topicHash string, categoryHash string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.TopicCategoryMappingKeyPrefix))
	key := append([]byte(topicHash))
	store.Set(key, []byte(categoryHash))
}
func (k Keeper) getCategoryByTopicHash(ctx sdk.Context, topicHash string) string {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.TopicCategoryMappingKeyPrefix))
	key := append([]byte(topicHash))
	value := store.Get(key)
	if value == nil {
		return ""
	}
	return string(value)
}

func (k Keeper) SetCategoryTopics(ctx sdk.Context, categoryHash string, topicHash string, topicScore uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.CategoryTopicsKeyPrefix+categoryHash))
	bzScore := k.EncodeScore(topicScore)
	var buffer bytes.Buffer
	buffer.Write(bzScore)
	buffer.WriteString(topicHash)
	key := buffer.Bytes()
	store.Set(key, []byte(topicHash))
}

func (k Keeper) DeleteFromCategoryTopicsByCategoryAndTopicId(ctx sdk.Context, categoryHash string, topicHash string, topicScore uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.CategoryTopicsKeyPrefix+categoryHash))
	bzScore := k.EncodeScore(topicScore)
	var buffer bytes.Buffer
	buffer.Write(bzScore)
	buffer.WriteString(topicHash)
	key := buffer.Bytes()
	store.Delete(key)
}

func (k Keeper) DeleteLastPostFromCategoryTopics(ctx sdk.Context, categoryHash string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.CategoryTopicsKeyPrefix+categoryHash))
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()
	earliestKey := iterator.Key()
	store.Delete(earliestKey)
}
func (k Keeper) GetCategoryTopics(ctx sdk.Context, categoryHash string, page uint64) ([]string, *query.PageResponse, uint64, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.CategoryTopicsKeyPrefix+categoryHash))

	categoryTopicsCount, _ := k.GetCategoryTopicsCount(ctx, categoryHash)
	if categoryTopicsCount == 0 {
		return []string{}, &query.PageResponse{
			NextKey: nil,
			Total:   uint64(categoryTopicsCount),
		}, uint64(0), nil
	}

	if page < 1 {
		page = 1
	}
	const pageSize = types.CategoryTopicsPageSize
	totalPages := categoryTopicsCount / pageSize
	if categoryTopicsCount%pageSize != 0 {
		totalPages += 1
	}

	if totalPages == 0 {
		totalPages = 1
	}

	//currentTime := time.Now()
	//unixMilli := currentTime.Unix()
	//lastTwoDigits := unixMilli % 100

	//pageIndex := lastTwoDigits % totalPages
	// 00-99  10000 00:1-100 01:101-200 02:201-300
	//first := lastTwoDigits * 100
	//first := pageIndex * pageSize
	first := (page - 1) * pageSize
	const totalTopics = types.CategoryTopicsCount
	if first >= totalTopics {
		return nil, nil, uint64(0), fmt.Errorf("offset exceeds total number of category posts")
	}

	pagination := &query.PageRequest{}

	pagination.Limit = pageSize
	pagination.Offset = first
	pagination.Reverse = true

	var topics []string

	pageRes, err := query.Paginate(store, pagination, func(key, value []byte) error {
		topics = append(topics, string(value))
		return nil
	})

	if err != nil {
		return nil, nil, uint64(0), err
	}
	return topics, pageRes, page, nil
}

func (k Keeper) SetCategoryTopicsCount(ctx sdk.Context, categoryHash string, count int64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.CategoryTopicsCountKeyPrefix))
	key := append([]byte(categoryHash))
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, uint64(count))
	store.Set(key, bz)
}
func (k Keeper) GetCategoryTopicsCount(ctx sdk.Context, categoryHash string) (int64, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.CategoryTopicsCountKeyPrefix))
	key := append([]byte(categoryHash))
	bz := store.Get(key)
	if bz == nil {
		return 0, true
	}
	count := int64(binary.BigEndian.Uint64(bz))
	return count, true
}

func (k Keeper) SetUncategorizedTopicsCount(ctx sdk.Context, count uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.UncategorizedTopicsCountKeyPrefix))
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, uint64(count))
	store.Set([]byte("count"), bz)
}
func (k Keeper) GetUncategorizedTopicsCount(ctx sdk.Context) (int64, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.UncategorizedTopicsCountKeyPrefix))
	bz := store.Get([]byte("count"))
	if bz == nil {
		return 0, true
	}
	count := int64(binary.BigEndian.Uint64(bz))
	return count, true
}
func (k Keeper) SetUncategorizedTopics(ctx sdk.Context, topicHash string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.UncategorizedTopicsKeyPrefix))
	blockTime := k.EncodeBlockTime(ctx)
	key := append([]byte(topicHash), blockTime...)
	store.Set(key, []byte(topicHash))
}
func (k Keeper) IsUncategorizedTopic(ctx sdk.Context, topicHash string) bool {
	prefixKey := append([]byte(types.UncategorizedTopicsKeyPrefix), []byte(topicHash)...)
	store := prefix.NewStore(ctx.KVStore(k.storeKey), prefixKey)
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()
	return iterator.Valid()
}
func (k Keeper) RemoveFromUncategorizedTopics(ctx sdk.Context, topicHash string) {
	// 1. Remove the topic from the UncategorizedTopics list
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.UncategorizedTopicsKeyPrefix+topicHash))
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		store.Delete(iterator.Key())
	}
}
func (k Keeper) GetUncategorizedTopics(ctx sdk.Context, page, limit int64) ([]string, *query.PageResponse, error) {
	if page < 1 || limit < 1 {
		return nil, nil, fmt.Errorf("invalid page (%d) or limit (%d)", page, limit)
	}
	unCategoryTopicsCount, _ := k.GetUncategorizedTopicsCount(ctx)
	if unCategoryTopicsCount == 0 {
		return []string{}, &query.PageResponse{
			NextKey: nil,
			Total:   uint64(unCategoryTopicsCount),
		}, nil
	}

	totalPages := unCategoryTopicsCount / limit
	if unCategoryTopicsCount%limit != 0 {
		totalPages += 1
	}

	if totalPages == 0 {
		totalPages = 1
	}

	currentTime := time.Now()
	unixMilli := currentTime.Unix()
	lastTwoDigits := unixMilli % 100

	pageIndex := lastTwoDigits % totalPages
	// 00-99  10000 00:1-100 01:101-200 02:201-300
	//first := lastTwoDigits * 100
	first := pageIndex * limit
	const totalTopics = types.CategoryTopicsCount
	if first >= totalTopics {
		return nil, nil, fmt.Errorf("offset exceeds total number of category posts")
	}

	pagination := &query.PageRequest{}

	pagination.Limit = uint64(limit)
	pagination.Offset = uint64(first)
	pagination.Reverse = true

	var topics []string

	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.UncategorizedTopicsKeyPrefix))
	pageRes, err := query.Paginate(store, pagination, func(key, value []byte) error {
		topics = append(topics, string(value))
		return nil
	})

	if err != nil {
		return nil, nil, err
	}
	//k.Logger().Warn("============topics:", "topics", topics)
	return topics, pageRes, nil
}

func (k Keeper) SetPoll(ctx sdk.Context, id string, sender string, optionId int64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.PollUserPrefix+id+"/"))
	optionIdBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(optionIdBytes, uint64(optionId))
	store.Set([]byte(sender), optionIdBytes)
}

// RemovePoll removes a user's vote record for a specific poll
func (k Keeper) RemovePoll(ctx sdk.Context, id string, sender string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.PollUserPrefix+id+"/"))
	store.Delete([]byte(sender))
}

func (k Keeper) GetPoll(ctx sdk.Context, id string, sender string) (string, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.PollUserPrefix+id+"/"))
	b := store.Get([]byte(sender))
	if b == nil {
		return "", false
	}
	return string(b), true
}

func itob(v int64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}
func btoi(bz []byte) int64 {
	if len(bz) != 8 {
		return 0
	}
	return int64(binary.BigEndian.Uint64(bz))
}
func (k Keeper) FollowTopic(ctx sdk.Context, address string, topicHash string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.FollowTopicPrefix+address+"/"))
	blockTime := ctx.BlockTime().Unix()
	key := append(itob(blockTime), []byte(topicHash)...)
	store.Set(key, []byte(topicHash))
}

func (k Keeper) GetFollowingTopics(ctx sdk.Context, address string, page uint64) ([]string, *query.PageResponse, uint64, error) {
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * types.PageSize
	pageRequest := &query.PageRequest{
		Offset:     offset,
		Limit:      types.PageSize,
		CountTotal: true,
		Reverse:    true,
	}

	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.FollowTopicPrefix+address+"/"))
	var followingTopics []string

	pageResponse, err := query.Paginate(store, pageRequest, func(key []byte, value []byte) error {
		var topicHash string
		topicHash = string(value)
		followingTopics = append(followingTopics, topicHash)
		return nil
	})

	if err != nil {
		return nil, nil, uint64(0), err
	}
	return followingTopics, pageResponse, page, nil
}

func (k Keeper) IsFollowingTopic(ctx sdk.Context, follower string, topicHash string) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.FollowTopicPrefix+follower+"/"))
	iterator := store.ReverseIterator(nil, nil)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		val := iterator.Value()
		if string(val) == topicHash {
			return true
		}
	}
	return false
}

func (k Keeper) SetFollowTopicTime(ctx sdk.Context, address string, topicHash string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.FollowTopicTimePrefix))
	blockTime := ctx.BlockTime().Unix()
	key := []byte(fmt.Sprintf("%s:%s", address, topicHash))
	store.Set(key, itob(blockTime))
}

func (k Keeper) GetFollowTopicTime(ctx sdk.Context, followerAddr string, targetAddr string) (uint64, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.FollowTopicTimePrefix))
	key := []byte(fmt.Sprintf("%s:%s", followerAddr, targetAddr))
	bz := store.Get(key)
	if bz == nil {
		return 0, false
	}
	followTime := btoi(bz)
	return uint64(followTime), true
}

func (k Keeper) DeleteFollowTopicTime(ctx sdk.Context, address string, topicHash string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.FollowTopicTimePrefix))
	key := []byte(fmt.Sprintf("%s:%s", address, topicHash))
	store.Delete(key)
}

func (k Keeper) UnfollowTopic(ctx sdk.Context, address string, time uint64, topicHash string) {
	storeFlowing := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.FollowTopicPrefix+address+"/"))
	keyFlowing := append(itob(int64(time)), []byte(topicHash)...)
	storeFlowing.Delete(keyFlowing)
}
func (k Keeper) SetCategoryOperator(ctx sdk.Context, address string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.CategoryOperatorKeyPrefix))
	store.Set([]byte("operator"), []byte(address))
}

func (k Keeper) GetCategoryOperator(ctx sdk.Context) (string, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.CategoryOperatorKeyPrefix))
	bz := store.Get([]byte("operator"))
	if bz == nil {
		return "", true
	}
	return string(bz), true
}

func (k Keeper) SetTopicImage(ctx sdk.Context, topicHash string, avatar string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.TopicImagePrefix))
	key := []byte(topicHash)
	//k.Logger().Warn("============avatar:", "key", key, "topicHash", topicHash)
	store.Set(key, []byte(avatar))
}
func (k Keeper) GetImageByTopic(ctx sdk.Context, topicHash string) string {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.TopicImagePrefix))
	key := []byte(topicHash)
	bz := store.Get(key)
	//k.Logger().Warn("============bz:", "key", key, "bz", string(bz))
	if bz == nil {
		return ""
	}
	return string(bz)
}

// SetPaidPostImage stores the single paid image (base64 encoded) for a given imageHash into on-chain KVStore.
func (k Keeper) SetPaidPostImage(ctx sdk.Context, imageHash string, image string) error {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.PostPaidImagePrefix))
	bz, err := json.Marshal(image)
	if err != nil {
		return fmt.Errorf("failed to marshal image: %w", err)
	}
	store.Set([]byte(imageHash), bz)
	return nil
}

// GetPaidPostImage retrieves the stored paid image (base64 encoded) for a given imageHash.
// Returns the image and a boolean indicating whether it was found.
func (k Keeper) GetPaidPostImage(ctx sdk.Context, imageHash string) (string, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.PostPaidImagePrefix))
	bz := store.Get([]byte(imageHash))
	if bz == nil {
		return "", false
	}
	var image string
	if err := json.Unmarshal(bz, &image); err != nil {
		return "", false
	}
	return image, true
}
