package keeper

import (
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

// SetPost stores a post in the state.
func (k Keeper) SetPost(ctx sdk.Context, post types.Post) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.PostKeyPrefix))
	bz := k.cdc.MustMarshal(&post)
	store.Set([]byte(post.Id), bz)

}

func (k Keeper) SetHomePosts(ctx sdk.Context, postId string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.HomePostKeyPrefix))

	blockTime := ctx.BlockTime().UnixNano()
	bzBlockTime := make([]byte, 8)
	binary.BigEndian.PutUint64(bzBlockTime, uint64(blockTime))

	key := append(bzBlockTime, []byte(postId)...)
	store.Set([]byte(key), []byte(postId))

	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	var count int
	var keysToDelete [][]byte
	for ; iterator.Valid(); iterator.Next() {
		count++
		if count > 100 {
			keysToDelete = append(keysToDelete, iterator.Key())
		}
	}

	for _, keyToDelete := range keysToDelete {
		store.Delete(keyToDelete)
	}

}

func (k Keeper) GetHomePosts(ctx sdk.Context, pagination *query.PageRequest) ([]string, *query.PageResponse, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.HomePostKeyPrefix))

	//var postIDs []string
	//
	//pageRes, err := query.Paginate(store, pagination, func(key, value []byte) error {
	//	postIDs = append(postIDs, string(value))
	//	return nil
	//})
	//
	//if err != nil {
	//	return nil, nil, err
	//}
	//return postIDs, pageRes, nil

	iterator := store.ReverseIterator(nil, nil)
	defer iterator.Close()
	var postIDs []string
	for ; iterator.Valid(); iterator.Next() {
		postID := string(iterator.Value())
		postIDs = append(postIDs, postID)
	}
	return postIDs, nil, nil
}

func (k Keeper) SetCategoryPosts(ctx sdk.Context, postId string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.CategoryPostKeyPrefix))
	blockTime := ctx.BlockTime().UnixNano()
	bzBlockTime := make([]byte, 8)
	binary.BigEndian.PutUint64(bzBlockTime, uint64(blockTime))

	key := append(bzBlockTime, []byte(postId)...)
	store.Set([]byte(key), []byte(postId))
}
func (k Keeper) GetCategoryPosts(ctx sdk.Context, pagination *query.PageRequest) ([]string, *query.PageResponse, error) {
	return nil, nil, nil
}
func (k Keeper) SetFollowedPosts(ctx sdk.Context, postId string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.FollowedPostKeyPrefix))
	blockTime := ctx.BlockTime().UnixNano()
	bzBlockTime := make([]byte, 8)
	binary.BigEndian.PutUint64(bzBlockTime, uint64(blockTime))

	key := append(bzBlockTime, []byte(postId)...)
	store.Set([]byte(key), []byte(postId))
}
func (k Keeper) GetFollowedPosts(ctx sdk.Context, pagination *query.PageRequest) ([]string, *query.PageResponse, error) {
	return nil, nil, nil
}
func (k Keeper) SetLikesIMade(ctx sdk.Context, likesIMade types.LikesIMade, sender string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.LikesIMadePrefix+sender+"/"))
	blockTime := ctx.BlockTime().Unix()
	//key := fmt.Sprintf("%s%s/%d%s", types.LikesIMadePrefix, sender, blockTime, likesIMade.PostId)
	key := fmt.Sprintf("%d-%s", blockTime, likesIMade.PostId)
	bz := k.cdc.MustMarshal(&likesIMade)
	store.Set([]byte(key), bz)
	k.Logger().Warn("=============SetLikesIMade called", "sender", sender, "post_id", likesIMade.PostId, "block_time", blockTime)
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

func (k Keeper) RemoveLikesIMade(ctx sdk.Context, sender string, postId string) error {
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
	blockTime := ctx.BlockTime().Unix()
	key := fmt.Sprintf("%d-%s", blockTime, likesIMade.PostId)
	bz := k.cdc.MustMarshal(&likesIMade)
	store.Set([]byte(key), bz)
}

func (k Keeper) RemoveSavesIMade(ctx sdk.Context, sender string, postId string) error {
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
	blockTime := ctx.BlockTime().Unix()
	key := fmt.Sprintf("%d", blockTime)
	bz := k.cdc.MustMarshal(&likesReceived)
	store.Set([]byte(key), bz)
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

func (k Keeper) RemoveLikesReceived(ctx sdk.Context, creator string, sender string, postId string) error {
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
//	// 使用 RFC3339Nano 确保时间的高精度
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
	moduleAddress := k.AccountKeeper.GetModuleAddress(types.ModuleName)
	fmt.Printf("==============types.ModuleName: [%s], =moduleAddress: [%s]", types.ModuleName, moduleAddress)
	return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, userAddr, amount)
}

func (k Keeper) SendCoinsFromAccountToModule(ctx sdk.Context, userAddr sdk.AccAddress, amount sdk.Coins) error {
	moduleAddress := k.AccountKeeper.GetModuleAddress(types.ModuleName)
	fmt.Printf("==============types.ModuleName: [%s], =moduleAddress: [%s]", types.ModuleName, moduleAddress)
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
