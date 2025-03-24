package keeper

import (
	"context"
	"cosmossdk.io/store/prefix"
	kvtypes "cosmossdk.io/store/types"
	"encoding/binary"
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"cosmossdk.io/orm/model/ormdb"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	apiv1 "github.com/rollchains/tlock/api/profile/v1"
	"github.com/rollchains/tlock/x/profile/types"
)

type Keeper struct {
	cdc codec.BinaryCodec

	logger log.Logger

	// state management
	Schema collections.Schema
	Params collections.Item[types.Params]
	OrmDB  apiv1.StateStore

	authority     string
	storeKey      kvtypes.StoreKey
	paramSubspace paramtypes.Subspace
}

// NewKeeper creates a new Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey kvtypes.StoreKey,
	storeService storetypes.KVStoreService,
	logger log.Logger,
	authority string,
	paramSpace paramtypes.Subspace,
) Keeper {
	logger = logger.With(log.ModuleKey, "x/"+types.ModuleName)
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

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

		storeKey:      storeKey,
		authority:     authority,
		paramSubspace: paramSpace,
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
// func (k *Keeper) InitGenesis(ctx context.Context, data *types.GenesisState) error {
//
//		if err := data.Params.Validate(); err != nil {
//			return err
//		}
//		return k.Params.Set(ctx, data.Params)
//	}
func (k *Keeper) InitGenesis(ctx sdk.Context, data *types.GenesisState) error {
	if err := data.Params.Validate(); err != nil {
		return err
	}
	err := k.SetParams(ctx, data.Params)
	if err != nil {
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

func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	k.paramSubspace.SetParamSet(ctx, &params)
	return nil
}

//	func (k Keeper) GetParamsSubspace(ctx sdk.Context) types.Params {
//		var params types.Params
//		k.paramSubspace.GetParamSet(ctx, &params)
//		return params
//	}
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	params, _ := k.Params.Get(ctx)
	return params
}
func (k Keeper) GetAdminAddress(ctx sdk.Context) (string, error) {
	params := k.GetParams(ctx)
	adminAddr := params.AdminAddress
	if adminAddr == "" {
		return "", types.ErrInvalidAdminAddress
	}
	return adminAddr, nil
}
func (k Keeper) GetChiefModerator(ctx sdk.Context) (string, error) {
	params := k.GetParams(ctx)
	chiefModerator := params.ChiefModerator
	if chiefModerator == "" {
		return "", types.ErrInvalidChiefModerator
	}
	return chiefModerator, nil
}
func (k Keeper) SetChiefModerator(ctx sdk.Context, address string) (bool, error) {
	params := k.GetParams(ctx)
	params.ChiefModerator = address
	err := k.SetParams(ctx, params)
	if err != nil {
		return false, err
	}
	return true, nil
}
func (k Keeper) SetProfile(ctx sdk.Context, profile types.Profile) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.ProfileKeyPrefix))
	bz := k.cdc.MustMarshal(&profile)
	store.Set([]byte(profile.WalletAddress), bz)
}

// GetProfile
func (k Keeper) GetProfile(ctx sdk.Context, address string) (types.Profile, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.ProfileKeyPrefix))
	bz := store.Get([]byte(address))
	if bz == nil {
		return types.Profile{
			WalletAddress: address,
			UserHandle:    k.TruncateAddressSuffix(address),
		}, true
	}

	var profile types.Profile
	k.cdc.MustUnmarshal(bz, &profile)
	return profile, true
}

func (k Keeper) AddToUserHandleList(ctx sdk.Context, userHandle string, address string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.ProfileUserHandleKeyPrefix))
	key := append([]byte(userHandle))
	store.Set(key, []byte(address))
}
func (k Keeper) HasUserHandle(ctx sdk.Context, userHandle string) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.ProfileUserHandleKeyPrefix))
	key := append([]byte(userHandle))
	return store.Has(key)
}
func (k Keeper) DeleteFromUserHandleList(ctx sdk.Context, userHandle string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.ProfileUserHandleKeyPrefix))
	key := append([]byte(userHandle))
	store.Delete(key)
}
func (k Keeper) GetAddressByUserHandle(ctx sdk.Context, userHandle string) string {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.ProfileUserHandleKeyPrefix))
	key := append([]byte(userHandle))
	bz := store.Get(key)
	if bz == nil {
		return ""
	}
	return string(bz)
}

func (k Keeper) AddToUserSearchList(ctx sdk.Context, uhnn string, userSearch types.UserSearch) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.ProfileUserSearchKeyPrefix))
	key := append([]byte(uhnn), []byte(userSearch.WalletAddress)...)
	bz := k.cdc.MustMarshal(&userSearch)
	store.Set(key, bz)
}
func (k Keeper) DeleteFromUserSearchList(ctx sdk.Context, uhnn string, address string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.ProfileUserSearchKeyPrefix))
	key := append([]byte(uhnn), []byte(address)...)
	store.Delete(key)
}
func (k Keeper) SearchUsersByMatching(ctx sdk.Context, matching string) ([]types.UserSearch, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.ProfileUserSearchKeyPrefix+matching))
	iterator := store.ReverseIterator(nil, nil)
	defer iterator.Close()
	var users []types.UserSearch
	const limit = 10
	count := 0
	for ; iterator.Valid() && count < limit; iterator.Next() {
		val := iterator.Value()
		var user types.UserSearch
		k.cdc.MustUnmarshal(val, &user)
		users = append(users, user)
		count++
	}
	return users, true
}

func (k Keeper) TruncateAddressSuffix(address string) string {
	runes := []rune(address)
	if len(runes) >= 10 {
		return string(runes[len(runes)-10:])
	}
	return string(runes)
}

func (k Keeper) CheckAndCreateUserHandle(ctx sdk.Context, walletAddress string) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.ProfileKeyPrefix))
	bz := store.Get([]byte(walletAddress))
	if bz == nil {
		profile := types.Profile{
			WalletAddress: walletAddress,
			UserHandle:    k.TruncateAddressSuffix(walletAddress),
		}
		k.SetProfile(ctx, profile)
	}
	return true
}

// Helper function to convert int64 to bytes
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

// AddFollowing adds the target address to the follower's following list
func (k Keeper) AddToFollowing(ctx sdk.Context, followerAddr string, targetAddr string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.ProfileFollowingPrefix+followerAddr+"/"))
	blockTime := ctx.BlockTime().Unix()
	key := append(itob(blockTime), []byte(targetAddr)...)
	store.Set(key, []byte(targetAddr))
}

// GetFollowing returns all following addresses for a given address
func (k Keeper) GetFollowing(ctx sdk.Context, address string) []string {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.ProfileFollowingPrefix+address+"/"))
	iterator := store.ReverseIterator(nil, nil)
	defer iterator.Close()

	var followings []string
	for ; iterator.Valid(); iterator.Next() {
		val := iterator.Value()
		followings = append(followings, string(val))
	}
	return followings
}
func (k Keeper) GetFollowingPagination(ctx sdk.Context, address string, page uint64, limit uint64) ([]string, *query.PageResponse, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.ProfileFollowingPrefix+address+"/"))
	first := page * limit
	pagination := &query.PageRequest{}
	pagination.Limit = limit
	pagination.Offset = first
	pagination.Reverse = true
	var followings []string
	pageRes, err := query.Paginate(store, pagination, func(key, value []byte) error {
		followings = append(followings, string(value))
		return nil
	})

	if err != nil {
		return nil, nil, err
	}
	return followings, pageRes, nil
}

// IsFollowing
func (k Keeper) IsFollowing(ctx sdk.Context, follower string, target string) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.ProfileFollowingPrefix+follower+"/"))
	iterator := store.ReverseIterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		val := iterator.Value()
		if string(val) == target {
			return true
		}
	}
	return false
}

func (k Keeper) Unfollow(ctx sdk.Context, followerAddr string, time uint64, targetAddr string) {
	storeFlowing := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.ProfileFollowingPrefix+followerAddr+"/"))
	keyFlowing := append(itob(int64(time)), []byte(targetAddr)...)
	storeFlowing.Delete(keyFlowing)

	storeFlower := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.ProfileFollowersPrefix+targetAddr+"/"))
	keyFlower := append(itob(int64(time)), []byte(followerAddr)...)
	storeFlower.Delete(keyFlower)
}

// AddFollower adds the follower address to the target's followers list
func (k Keeper) AddToFollowers(ctx sdk.Context, targetAddr string, followerAddr string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.ProfileFollowersPrefix+targetAddr+"/"))
	blockTime := ctx.BlockTime().Unix()
	key := append(itob(blockTime), []byte(followerAddr)...)
	store.Set(key, []byte(followerAddr))
}

// GetFollowers returns all follower addresses for a given address
func (k Keeper) GetFollowers(ctx sdk.Context, address string) []string {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.ProfileFollowersPrefix+address+"/"))
	iterator := store.ReverseIterator(nil, nil)
	defer iterator.Close()

	var followers []string
	for ; iterator.Valid(); iterator.Next() {
		val := iterator.Value()
		followers = append(followers, string(val))
	}
	return followers
}

// AddFollowing adds the target address to the follower's following list
func (k Keeper) SetFollowTime(ctx sdk.Context, followerAddr string, targetAddr string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.ProfileFollowTimePrefix))
	blockTime := ctx.BlockTime().Unix()
	key := []byte(fmt.Sprintf("%s:%s", followerAddr, targetAddr))
	store.Set(key, itob(blockTime))
}

// GetFollowTime
func (k Keeper) GetFollowTime(ctx sdk.Context, followerAddr string, targetAddr string) (uint64, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.ProfileFollowTimePrefix))
	key := []byte(fmt.Sprintf("%s:%s", followerAddr, targetAddr))
	bz := store.Get(key)
	if bz == nil {
		return 0, false
	}
	followTime := btoi(bz)
	return uint64(followTime), true
}
func (k Keeper) DeleteFollowTime(ctx sdk.Context, followerAddr string, targetAddr string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.ProfileFollowTimePrefix))
	key := []byte(fmt.Sprintf("%s:%s", followerAddr, targetAddr))
	store.Delete(key)
}

func (k Keeper) EncodeBlockTime(ctx sdk.Context) []byte {
	blockTime := ctx.BlockTime().Unix()
	bzBlockTime := make([]byte, 8)
	binary.BigEndian.PutUint64(bzBlockTime, uint64(blockTime))
	return bzBlockTime
}

func (k Keeper) SetActivitiesReceived(ctx sdk.Context, activitiesReceived types.ActivitiesReceived, targetAddr string, operator string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.ActivitiesReceivedPrefix+targetAddr+"/"))
	blockTime := k.EncodeBlockTime(ctx)
	key := append(append(blockTime, []byte(activitiesReceived.ActivitiesType.String())...), []byte(operator)...)
	bz := k.cdc.MustMarshal(&activitiesReceived)
	store.Set(key, bz)
}

func (k Keeper) GetActivitiesReceived(ctx sdk.Context, address string, page uint64) ([]*types.ActivitiesReceived, *query.PageResponse, uint64, error) {
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * types.PageSize

	pageRequest := &query.PageRequest{
		Offset:     offset,
		Limit:      types.PageSize,
		CountTotal: true,
	}

	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.ActivitiesReceivedPrefix+address+"/"))
	var activitiesList []*types.ActivitiesReceived

	pageResponse, err := query.Paginate(store, pageRequest, func(key []byte, value []byte) error {
		var activity types.ActivitiesReceived
		if err := k.cdc.Unmarshal(value, &activity); err != nil {
			return err
		}
		activitiesList = append(activitiesList, &activity)
		return nil
	})

	if err != nil {
		return nil, nil, uint64(0), err
	}
	return activitiesList, pageResponse, page, nil
}

func (k Keeper) DeleteLastActivitiesReceived(ctx sdk.Context, wallet string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.ActivitiesReceivedPrefix+wallet+"/"))
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()
	earliestKey := iterator.Key()
	store.Delete(earliestKey)
}

func (k Keeper) SetActivitiesReceivedCount(ctx sdk.Context, walletAddr string, count int64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.ActivitiesReceivedCountPrefix))
	key := append([]byte(walletAddr))
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, uint64(count))
	store.Set(key, bz)
}

func (k Keeper) GetActivitiesReceivedCount(ctx sdk.Context, walletAddr string) (int64, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.ActivitiesReceivedCountPrefix))
	key := append([]byte(walletAddr))
	bz := store.Get(key)
	if bz == nil {
		return 0, true
	}
	count := int64(binary.BigEndian.Uint64(bz))
	return count, true
}

func (k Keeper) AddAdmin(ctx sdk.Context, address string) error {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.AuthorityKeyPrefix))
	key := append([]byte(address))

	if store.Has(key) {
		return fmt.Errorf("address %s is already an admin", address)
	}

	store.Set(key, []byte(address))
	return nil
}

func (k Keeper) RemoveAdmin(ctx sdk.Context, address string) error {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.AuthorityKeyPrefix))
	key := append([]byte(address))
	store.Delete(key)
	return nil
}

func (k Keeper) IsAdmin(ctx sdk.Context, address string) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.AuthorityKeyPrefix))
	key := append([]byte(address))
	return store.Has(key)
}

func (k Keeper) AddEditableAdmin(ctx sdk.Context, address string) error {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.AuthorityEditableAdminKeyPrefix))
	key := append([]byte(address))
	if store.Has(key) {
		return fmt.Errorf("address %s is already an editable admin", address)
	}
	store.Set(key, []byte(address))
	return nil
}
func (k Keeper) IsEditableAdmin(ctx sdk.Context, address string) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.AuthorityEditableAdminKeyPrefix))
	key := append([]byte(address))
	return store.Has(key)
}
func (k Keeper) RemoveEditableAdmin(ctx sdk.Context, address string) error {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.AuthorityEditableAdminKeyPrefix))
	key := append([]byte(address))
	store.Delete(key)
	return nil
}
