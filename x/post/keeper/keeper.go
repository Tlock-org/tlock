package keeper

import (
	"bytes"
	"context"
	"cosmossdk.io/store/prefix"
	"cosmossdk.io/x/feegrant"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"strings"
	"time"

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

	sdkmath "cosmossdk.io/math"
	"encoding/binary"
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

		authority: authority,
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
		return 0, true
	}

	count := int64(binary.BigEndian.Uint64(bz))
	return count, true
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

func (k Keeper) GetHomePosts(ctx sdk.Context, req *types.QueryHomePostsRequest) ([]string, *query.PageResponse, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.HomePostsKeyPrefix))

	homePostsCount, _ := k.GetHomePostsCount(ctx)
	if homePostsCount == 0 {
		return []string{}, &query.PageResponse{
			NextKey: nil,
			Total:   uint64(homePostsCount),
		}, nil
	}

	const pageSize = types.HomePostsPageSize
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
	fmt.Sprintf("===============lastTwoDigits: %d", lastTwoDigits)

	pageIndex := lastTwoDigits % totalPages
	// 00-99  10000 00:1-100 01:101-200 02:201-300
	//first := lastTwoDigits * 100
	first := pageIndex * pageSize

	const totalPosts = types.HomePostsCount
	if first >= totalPosts {
		return nil, nil, fmt.Errorf("offset exceeds total number of home posts")
	}

	pagination := &query.PageRequest{}

	pagination.Limit = pageSize
	pagination.Offset = uint64(first)
	pagination.Reverse = true

	var postIDs []string

	pageRes, err := query.Paginate(store, pagination, func(key, value []byte) error {
		postIDs = append(postIDs, string(value))
		return nil
	})

	if err != nil {
		return nil, nil, err
	}
	return postIDs, pageRes, nil
}

func (k Keeper) GetFirstPageHomePosts(ctx sdk.Context) ([]string, *query.PageResponse, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.HomePostsKeyPrefix))
	homePostsCount, _ := k.GetHomePostsCount(ctx)
	if homePostsCount == 0 {
		return []string{}, &query.PageResponse{
			NextKey: nil,
			Total:   uint64(homePostsCount),
		}, nil
	}

	const pageSize = types.HomePostsPageSize
	totalPages := homePostsCount / pageSize
	if homePostsCount%pageSize != 0 {
		totalPages += 1
	}

	if totalPages == 0 {
		totalPages = 1
	}

	pagination := &query.PageRequest{}

	pagination.Limit = pageSize
	pagination.Offset = uint64(0)
	pagination.Reverse = true

	var postIDs []string

	pageRes, err := query.Paginate(store, pagination, func(key, value []byte) error {
		postIDs = append(postIDs, string(value))
		return nil
	})

	if err != nil {
		return nil, nil, err
	}
	return postIDs, pageRes, nil
}

func (k Keeper) DeleteLastPostFromHomePosts(ctx sdk.Context) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.HomePostsKeyPrefix))

	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	if !iterator.Valid() {
		panic("homePostsCount exceeds 1000 but no posts found")
	}
	earliestKey := iterator.Key()
	earliestPostID := string(iterator.Value())

	store.Delete(earliestKey)
	k.Logger().Info("Deleted earliest home post", "post_id", earliestPostID)
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

func (k Keeper) SetTopicSearch(ctx sdk.Context, topic string) {
	topicLower := strings.ToLower(topic)
	topicHash := k.sha256Generate(topic)
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

func (k Keeper) GetTopicPosts(ctx sdk.Context, topic string) ([]string, *query.PageResponse, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.TopicPostsKeyPrefix+topic))

	topicPostsCount, _ := k.GetTopicPostsCount(ctx, topic)
	if topicPostsCount == 0 {
		return []string{}, &query.PageResponse{
			NextKey: nil,
			Total:   uint64(topicPostsCount),
		}, nil
	}

	const pageSize = types.TopicPostsPageSize
	totalPages := topicPostsCount / pageSize
	if topicPostsCount%pageSize != 0 {
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

	const totalPosts = types.TopicPostsCount
	if first >= totalPosts {
		return nil, nil, fmt.Errorf("offset exceeds total number of topic posts")
	}

	pagination := &query.PageRequest{}

	pagination.Limit = pageSize
	pagination.Offset = uint64(first)
	pagination.Reverse = true

	var postIDs []string

	pageRes, err := query.Paginate(store, pagination, func(key, value []byte) error {
		postIDs = append(postIDs, string(value))
		return nil
	})

	if err != nil {
		return nil, nil, err
	}
	return postIDs, pageRes, nil
}

func (k Keeper) SetTopicPosts(ctx sdk.Context, topic string, postId string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.TopicPostsKeyPrefix+topic))
	k.Logger().Warn("===SetTopicPosts time:", "blocktime", ctx.BlockTime().Unix())
	blockTime := k.EncodeBlockTime(ctx)
	key := append(blockTime, []byte(postId)...)
	k.Logger().Warn("===SetTopicPosts key:", "key", key)
	store.Set(key, []byte(postId))
}
func (k Keeper) IsPostInTopics(ctx sdk.Context, postId string) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.PostTopicsMappingKeyPrefix))
	key := append([]byte(postId))
	k.Logger().Warn("===IsPostInTopics key:", "key", key)
	return store.Has(key)
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
func (k Keeper) DeleteFromTopicPostsByTopicAndPostId(ctx sdk.Context, topic string, postId string, homePostsUpdate int64) {
	k.Logger().Warn("===DeleteFromTopicPostsByTopicAndPostId time:", "time", uint64(homePostsUpdate))
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.TopicPostsKeyPrefix+topic))
	bzBlockTime := make([]byte, 8)
	binary.BigEndian.PutUint64(bzBlockTime, uint64(homePostsUpdate))
	key := append(bzBlockTime, []byte(postId)...)
	k.Logger().Warn("===DeleteFromTopicPostsByTopicAndPostId key:", "key", key)
	store.Delete(key)
}
func (k Keeper) DeleteLastPostFromTopicPosts(ctx sdk.Context, topic string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.TopicPostsKeyPrefix+topic))
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	if !iterator.Valid() {
		panic("topicPostsCount exceeds 1000 but no posts found")
	}
	earliestKey := iterator.Key()
	earliestPostID := string(iterator.Value())

	store.Delete(earliestKey)
	k.Logger().Info("Deleted earliest topic posts", "post_id", earliestPostID)
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
		panic("userCreatedPostsCount exceeds 1000 but no posts found")
	}
	earliestKey := iterator.Key()
	earliestPostID := string(iterator.Value())
	store.Delete(earliestKey)
	k.Logger().Info("Deleted earliest user created post", "post_id", earliestPostID)
}

func (k Keeper) GetUserCreatedPosts(ctx sdk.Context, creator string) ([]string, *query.PageResponse, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.UserCreatedPostsKeyPrefix+creator))

	userCreatedPostsCount, _ := k.GetUserCreatedPostsCount(ctx, creator)
	if userCreatedPostsCount == 0 {
		return []string{}, &query.PageResponse{
			NextKey: nil,
			Total:   uint64(userCreatedPostsCount),
		}, nil
	}

	const pageSize = types.UserCreatedPostsPageSize
	totalPages := userCreatedPostsCount / pageSize
	if userCreatedPostsCount%pageSize != 0 {
		totalPages += 1
	}

	if totalPages == 0 {
		totalPages = 1
	}

	currentTime := time.Now()
	unixMilli := currentTime.Unix()
	lastTwoDigits := unixMilli % 10

	pageIndex := lastTwoDigits % totalPages
	first := pageIndex * pageSize

	const totalPosts = types.UserCreatedPostsCount
	if first >= totalPosts {
		return nil, nil, fmt.Errorf("offset exceeds total number of home posts")
	}

	pagination := &query.PageRequest{}

	pagination.Limit = pageSize
	pagination.Offset = uint64(first)
	pagination.Reverse = true

	var postIDs []string

	pageRes, err := query.Paginate(store, pagination, func(key, value []byte) error {
		postIDs = append(postIDs, string(value))
		return nil
	})

	if err != nil {
		return nil, nil, err
	}
	return postIDs, pageRes, nil
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
		return nil, nil, err
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

func (k Keeper) GetCommentsByParentId(ctx sdk.Context, parentId string) []string {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.CommentListKeyPrefix))
	commentStore := prefix.NewStore(store, []byte(parentId))
	iterator := commentStore.ReverseIterator(nil, nil)
	defer func(iterator kvtypes.Iterator) {
		err := iterator.Close()
		if err != nil {
			k.logger.Error("GetCommentsByParentId iterator error")
		}
	}(iterator)
	var ids []string
	for ; iterator.Valid(); iterator.Next() {
		id := string(iterator.Value())
		ids = append(ids, id)
	}
	return ids
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

func (k Keeper) GetCommentsReceived(ctx sdk.Context, creator string) ([]string, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.CommentsReceivedPrefix+creator+"/"))
	iterator := store.ReverseIterator(nil, nil)
	defer iterator.Close()
	var list []string
	for ; iterator.Valid(); iterator.Next() {
		id := string(iterator.Value())
		list = append(list, id)
	}
	return list, nil
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
func (k Keeper) GetLikesIMade(ctx sdk.Context, sender string) ([]*types.LikesIMade, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.LikesIMadePrefix+sender+"/"))

	iterator := store.ReverseIterator(nil, nil)
	defer iterator.Close()

	var likesList []*types.LikesIMade

	for ; iterator.Valid(); iterator.Next() {
		var like types.LikesIMade
		err := k.cdc.Unmarshal(iterator.Value(), &like)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal LikesIMade: %w", err)
		}
		likeCopy := like
		likesList = append(likesList, &likeCopy)
	}
	k.Logger().Warn("===========GetLikesIMade retrieved likes", "count", len(likesList))
	return likesList, nil
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
			return nil, fmt.Errorf("failed to unmarshal LikesIMade: %w", err)
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
			return fmt.Errorf("failed to unmarshal LikesImade: %w", err)
		}

		if like.PostId == postId {
			key := iterator.Key()
			store.Delete(key)
			k.Logger().Warn("UnlikePost removed like", "sender", sender, "post_id", postId, "block_time", like.Timestamp)
			return nil
		}
	}
	return fmt.Errorf("like not found for sender %s and postId %s", sender, postId)
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
			return fmt.Errorf("failed to unmarshal LikesImade: %w", err)
		}

		if like.PostId == postId {
			key := iterator.Key()
			store.Delete(key)
			k.Logger().Warn("UnlikePost removed like", "sender", sender, "post_id", postId, "block_time", like.Timestamp)
			return nil
		}
	}
	return fmt.Errorf("like not found for sender %s and postId %s", sender, postId)
}

func (k Keeper) GetSavesIMade(ctx sdk.Context, sender string) ([]*types.LikesIMade, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.SavesIMadePrefix+sender+"/"))

	iterator := store.ReverseIterator(nil, nil)
	defer iterator.Close()

	var savesList []*types.LikesIMade

	for ; iterator.Valid(); iterator.Next() {
		var save types.LikesIMade
		err := k.cdc.Unmarshal(iterator.Value(), &save)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal SavesIMade: %w", err)
		}
		saveCopy := save
		savesList = append(savesList, &saveCopy)
	}
	k.Logger().Warn("===========GetSavesIMade retrieved saves", "count", len(savesList))
	return savesList, nil
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

func (k Keeper) GetLikesReceived(ctx sdk.Context, creator string) ([]*types.LikesReceived, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.LikesReceivedPrefix+creator+"/"))

	iterator := store.ReverseIterator(nil, nil)
	defer iterator.Close()

	var list []*types.LikesReceived

	for ; iterator.Valid(); iterator.Next() {
		var received types.LikesReceived
		err := k.cdc.Unmarshal(iterator.Value(), &received)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal LikesReceived: %w", err)
		}
		receivedCopy := received
		list = append(list, &receivedCopy)
	}
	k.Logger().Warn("===========Get LikesReceived retrieved likes", "count", len(list))
	return list, nil
}

func (k Keeper) RemoveFromLikesReceived(ctx sdk.Context, creator string, sender string, postId string) error {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.LikesReceivedPrefix+creator+"/"))
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var received types.LikesReceived
		err := k.cdc.Unmarshal(iterator.Value(), &received)
		if err != nil {
			return fmt.Errorf("failed to unmarshal LikesReceived: %w", err)
		}

		if received.PostId == postId && received.LikerAddress == sender {
			key := iterator.Key()
			store.Delete(key)
			k.Logger().Warn("UnlikePost removed LikesReceived", "creator", creator, "sender", sender, "post_id", postId, "block_time", received.Timestamp)
			return nil
		}
	}
	return fmt.Errorf("like not found for creator %s sender %s and postId %s", creator, sender, postId)
}

func (k Keeper) PostReward(ctx sdk.Context, post types.Post) {
	// send post reward
	amount := sdk.NewCoins(sdk.NewCoin(types.DenomBase, sdkmath.NewInt(10)))
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

	ctx.Logger().Debug("===========userAddress, amount:", userAddr, amount)
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
	specificAddress := sdk.MustAccAddressFromBech32("tlock1vj037yzhdslqzens3lu5vfjay9y8n956gqwyqw")

	if sender.Equals(specificAddress) {
		now := ctx.BlockTime()
		//oneHour := now.Add(1 * time.Hour)
		oneDay := now.Add(24 * time.Hour)
		//oneDay := now.Add(10 * time.Second)
		//period := 24 * time.Hour
		period := 10 * time.Second
		totalSpendLimit := sdk.NewCoins(sdk.NewCoin("TOK", sdkmath.NewInt(5)))
		spendLimit := sdk.NewCoins(sdk.NewCoin("TOK", sdkmath.NewInt(2)))
		// create a basic allowance
		basicAllowance := feegrant.BasicAllowance{
			SpendLimit: totalSpendLimit,
			Expiration: &oneDay,
		}

		// create a periodic allowance
		periodicAllowance := &feegrant.PeriodicAllowance{
			Basic:            basicAllowance,
			Period:           period,
			PeriodSpendLimit: spendLimit,
			PeriodCanSpend:   spendLimit,
			PeriodReset:      now.Add(period),
		}

		granter := k.AccountKeeper.GetModuleAddress(types.ModuleName)
		grantee := userAddr

		err := k.FeeGrantKeeper.GrantAllowance(ctx, granter, grantee, periodicAllowance)
		if err != nil {
			ctx.Logger().Error("Failed to grant allowance", "error", err)
			return
		}
	} else {
		ctx.Logger().Error("Failed to grant allowance", "error", "=====")
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
func (k Keeper) GetCategoryPosts(ctx sdk.Context, categoryHash string) ([]string, *query.PageResponse, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.CategoryPostsKeyPrefix+categoryHash))

	categoryPostsCount, _ := k.GetTopicPostsCount(ctx, categoryHash)
	if categoryPostsCount == 0 {
		return []string{}, &query.PageResponse{
			NextKey: nil,
			Total:   uint64(categoryPostsCount),
		}, nil
	}

	const pageSize = types.CategoryPostsPageSize
	totalPages := categoryPostsCount / pageSize
	if categoryPostsCount%pageSize != 0 {
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

	const totalPosts = types.CategoryPostsCount
	if first >= totalPosts {
		return nil, nil, fmt.Errorf("offset exceeds total number of category posts")
	}

	pagination := &query.PageRequest{}

	pagination.Limit = pageSize
	pagination.Offset = uint64(first)
	pagination.Reverse = true

	var postIDs []string

	pageRes, err := query.Paginate(store, pagination, func(key, value []byte) error {
		postIDs = append(postIDs, string(value))
		return nil
	})

	if err != nil {
		return nil, nil, err
	}
	return postIDs, pageRes, nil
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

func (k Keeper) addToHotTopics72(ctx sdk.Context, topicHash string, topicScore uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.HotTopics72KeyPrefix))
	bzScore := k.EncodeScore(topicScore)
	var buffer bytes.Buffer
	buffer.Write(bzScore)
	buffer.WriteString(topicHash)
	key := buffer.Bytes()
	k.Logger().Warn("============addToHotTopics72 key:", "key", key)
	store.Set(key, []byte(topicHash))
}
func (k Keeper) deleteFormHotTopics72(ctx sdk.Context, topicHash string, topicScore uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.HotTopics72KeyPrefix))
	bzScore := k.EncodeScore(topicScore)
	var buffer bytes.Buffer
	buffer.Write(bzScore)
	buffer.WriteString(topicHash)
	key := buffer.Bytes()
	k.Logger().Warn("============deleteFormHotTopics72 key:", "key", key)
	store.Delete(key)
}
func (k Keeper) GetHotTopics72(ctx sdk.Context) ([]string, *query.PageResponse, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.HotTopics72KeyPrefix))

	hotTopics72Count, _ := k.GetHotTopics72Count(ctx)
	if hotTopics72Count == 0 {
		return []string{}, &query.PageResponse{
			NextKey: nil,
			Total:   uint64(hotTopics72Count),
		}, nil
	}

	const pageSize = types.HotTopics72PageSize
	totalPages := hotTopics72Count / pageSize
	if hotTopics72Count%pageSize != 0 {
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

	const totalTopics = types.HotTopics72Count
	if first >= totalTopics {
		return nil, nil, fmt.Errorf("offset exceeds total number of hot topics")
	}

	pagination := &query.PageRequest{}

	pagination.Limit = pageSize
	pagination.Offset = uint64(first)
	pagination.Reverse = true

	var postIDs []string

	pageRes, err := query.Paginate(store, pagination, func(key, value []byte) error {
		postIDs = append(postIDs, string(value))
		return nil
	})
	k.Logger().Warn("======postIds:", "postIDs", postIDs)
	if err != nil {
		return nil, nil, err
	}
	return postIDs, pageRes, nil
}
func (k Keeper) SetHotTopics72Count(ctx sdk.Context, count int64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.HotTopics72CountKeyPrefix))
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, uint64(count))
	store.Set([]byte("count"), bz)
}
func (k Keeper) GetHotTopics72Count(ctx sdk.Context) (int64, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.HotTopics72CountKeyPrefix))
	bz := store.Get([]byte("count"))
	if bz == nil {
		return 0, true
	}
	count := int64(binary.BigEndian.Uint64(bz))
	return count, true
}
func (k Keeper) DeleteLastFromHotTopics72(ctx sdk.Context) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.HotTopics72CountKeyPrefix))
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()
	if !iterator.Valid() {
		panic("hotTopics72Count exceeds 1000 but no posts found")
	}
	earliestKey := iterator.Key()
	store.Delete(earliestKey)
}
func (k Keeper) SetTopicCategoryMapping(ctx sdk.Context, topicHash string, category string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.TopicCategoryMappingKeyPrefix))
	key := append([]byte(topicHash))
	value, _ := json.Marshal(category)
	store.Set(key, value)
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

func (k Keeper) SetCategoryTopics(ctx sdk.Context, topicScore uint64, categoryHash string, topicHash string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.CategoryTopicsKeyPrefix+categoryHash))
	bzScore := k.EncodeScore(topicScore)
	var buffer bytes.Buffer
	//buffer.WriteString(categoryHash)
	buffer.Write(bzScore)
	buffer.WriteString(topicHash)
	key := buffer.Bytes()
	store.Set(key, []byte(topicHash))
}
func (k Keeper) DeleteLastPostFromCategoryTopics(ctx sdk.Context, categoryHash string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.CategoryTopicsKeyPrefix+categoryHash))
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()
	earliestKey := iterator.Key()
	store.Delete(earliestKey)
}
func (k Keeper) GetCategoryTopics(ctx sdk.Context, categoryHash string) ([]string, *query.PageResponse, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.CategoryTopicsKeyPrefix+categoryHash))

	categoryTopicsCount, _ := k.GetCategoryTopicsCount(ctx, categoryHash)
	if categoryTopicsCount == 0 {
		return []string{}, &query.PageResponse{
			NextKey: nil,
			Total:   uint64(categoryTopicsCount),
		}, nil
	}

	const pageSize = types.CategoryTopicsPageSize
	totalPages := categoryTopicsCount / pageSize
	if categoryTopicsCount%pageSize != 0 {
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
	const totalTopics = types.CategoryTopicsCount
	if first >= totalTopics {
		return nil, nil, fmt.Errorf("offset exceeds total number of category posts")
	}

	pagination := &query.PageRequest{}

	pagination.Limit = pageSize
	pagination.Offset = uint64(first)
	pagination.Reverse = true

	var topics []string

	pageRes, err := query.Paginate(store, pagination, func(key, value []byte) error {
		topics = append(topics, string(value))
		return nil
	})

	if err != nil {
		return nil, nil, err
	}
	return topics, pageRes, nil
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
func (k Keeper) SetPoll(ctx sdk.Context, id string, sender string, optionId int64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.PollUserPrefix+id+"/"))
	optionIdBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(optionIdBytes, uint64(optionId))
	store.Set([]byte(sender), optionIdBytes)
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

func (k Keeper) GetFollowingTopics(ctx sdk.Context, address string) []string {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.FollowTopicPrefix+address+"/"))
	iterator := store.ReverseIterator(nil, nil)
	defer iterator.Close()
	var followingTopics []string
	for ; iterator.Valid(); iterator.Next() {
		val := iterator.Value()
		followingTopics = append(followingTopics, string(val))
	}
	return followingTopics
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
