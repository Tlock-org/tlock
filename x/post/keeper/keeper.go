package keeper

import (
	"bytes"
	"context"
	"cosmossdk.io/store/prefix"
	"cosmossdk.io/x/feegrant"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
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
	fmt.Sprintf("===============key:%s", key)

	//iterator := store.Iterator(nil, nil)
	//defer iterator.Close()
	//
	//var count int
	//var keysToDelete [][]byte
	//for ; iterator.Valid(); iterator.Next() {
	//
	//	existingPostID := string(iterator.Value())
	//	if existingPostID == postId {
	//		return
	//	}
	//
	//	count++
	//	if count > 100 {
	//		keysToDelete = append(keysToDelete, iterator.Key())
	//	}
	//}
	//
	//for _, keyToDelete := range keysToDelete {
	//	store.Delete(keyToDelete)
	//}
	store.Set(key, []byte(postId))
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

	const pageSize = types.PageSize
	totalPages := homePostsCount / pageSize
	if homePostsCount%pageSize != 0 {
		totalPages += 1
	}

	if totalPages == 0 {
		totalPages = 1
	}

	currentTime := time.Now()
	unixMilli := currentTime.Unix()
	lastTwoDigits := unixMilli % 10
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

	//iterator := store.ReverseIterator(nil, nil)
	//defer iterator.Close()
	//var postIDs []string
	//for ; iterator.Valid(); iterator.Next() {
	//	postID := string(iterator.Value())
	//	postIDs = append(postIDs, postID)
	//}
	//return postIDs, nil, nil
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

	const pageSize = types.PageSize
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

	//iterator := store.Iterator(nil, nil)
	//defer iterator.Close()
	//
	//for ; iterator.Valid(); iterator.Next() {
	//	currentPostId := string(iterator.Value())
	//	if currentPostId == postId {
	//		earliestKey := iterator.Key()
	//		store.Delete(earliestKey)
	//		return
	//	}
	//}
}

func (k Keeper) IsPostInHomePosts(ctx sdk.Context, postId string, homePostsUpdate int64) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.HomePostsKeyPrefix))

	bzBlockTime := make([]byte, 8)
	binary.BigEndian.PutUint64(bzBlockTime, uint64(homePostsUpdate))

	key := append(bzBlockTime, []byte(postId)...)

	get := store.Get(key)
	if get != nil {
		return true
	} else {
		return false
	}
	//iterator := store.Iterator(nil, nil)
	//defer iterator.Close()
	//
	//for ; iterator.Valid(); iterator.Next() {
	//	currentPostId := string(iterator.Value())
	//	if currentPostId == postId {
	//		return true
	//	}
	//}
	//
	//return false
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

func (k Keeper) SetCategoryPosts(ctx sdk.Context, postId string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.CategoryPostsKeyPrefix))
	blockTime := k.EncodeBlockTime(ctx)
	key := append(blockTime, []byte(postId)...)
	store.Set([]byte(key), []byte(postId))
}
func (k Keeper) GetCategoryPosts(ctx sdk.Context, pagination *query.PageRequest) ([]string, *query.PageResponse, error) {
	return nil, nil, nil
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
func (k Keeper) SetLikesIMade(ctx sdk.Context, likesIMade types.LikesIMade, sender string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.LikesIMadePrefix+sender+"/"))
	//blockTime := ctx.BlockTime().Unix()
	//key := fmt.Sprintf("%d-%s", blockTime, likesIMade.PostId)
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
func (k Keeper) generatePostID(data string) string {
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
