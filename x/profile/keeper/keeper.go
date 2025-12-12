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
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"strings"

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
	accountKeeper authkeeper.AccountKeeper
}

// NewKeeper creates a new Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey kvtypes.StoreKey,
	storeService storetypes.KVStoreService,
	logger log.Logger,
	authority string,
	paramSpace paramtypes.Subspace,
	accountKeeper authkeeper.AccountKeeper,
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
		logger.Error("Failed to create ORM module database", "error", err)
		panic(err) // Keep panic for initialization failure as it's critical
	}

	store, err := apiv1.NewStateStore(db)
	if err != nil {
		logger.Error("Failed to create state store", "error", err)
		panic(err) // Keep panic for initialization failure as it's critical
	}

	k := Keeper{
		cdc:    cdc,
		logger: logger,

		Params: collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),
		OrmDB:  store,

		storeKey:      storeKey,
		authority:     authority,
		paramSubspace: paramSpace,
		accountKeeper: accountKeeper,
	}

	schema, err := sb.Build()
	if err != nil {
		logger.Error("Failed to build schema", "error", err)
		return Keeper{}
	}

	k.Schema = schema

	return k
}

func (k Keeper) Logger() log.Logger {
	return k.logger
}

// InitGenesis initializes the module's state from a genesis state.
func (k *Keeper) InitGenesis(ctx sdk.Context, data *types.GenesisState) error {
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
func (k Keeper) GetProfileForUpdate(ctx sdk.Context, address string) (types.Profile, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.ProfileKeyPrefix))
	bz := store.Get([]byte(address))
	if bz == nil {
		return types.Profile{
			WalletAddress: address,
		}, false
	}

	var profile types.Profile
	k.cdc.MustUnmarshal(bz, &profile)
	return profile, true
}
func (k Keeper) SetProfileAvatar(ctx sdk.Context, address string, avatar string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.ProfileAvatarPrefix))
	key := []byte(address)
	store.Set(key, []byte(avatar))
}
func (k Keeper) GetAvatarByAddress(ctx sdk.Context, address string) string {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.ProfileAvatarPrefix))
	key := []byte(address)
	bz := store.Get(key)
	if bz == nil {
		return ""
	}
	return string(bz)
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
func padToFixedLength(s string, length int, padChar byte) string {
	if len(s) >= length {
		return s[:length]
	}
	padding := strings.Repeat(string(padChar), length-len(s))
	return s + padding
}

//	func (k Keeper) AddToUserSearchList(ctx sdk.Context, uhnn string, userSearch types.UserSearch) {
//		store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.ProfileUserSearchKeyPrefix))
//		key := append([]byte(strings.ToLower(uhnn)), []byte(userSearch.WalletAddress)...)
//		bz := k.cdc.MustMarshal(&userSearch)
//		store.Set(key, bz)
//	}
func (k Keeper) AddToUserSearchList(ctx sdk.Context, uhnn string, userSearch types.UserSearch) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.ProfileUserSearchKeyPrefix))
	uhnnLower := strings.ToLower(uhnn)
	paddedUhnn := padToFixedLength(uhnnLower, types.UserSearchFixedLength, types.UserSearchPaddingChar)
	key := append([]byte(paddedUhnn+":"), []byte(userSearch.WalletAddress)...)
	bz := k.cdc.MustMarshal(&userSearch)
	store.Set(key, bz)
}
func (k Keeper) DeleteFromUserSearchList(ctx sdk.Context, uhnn string, address string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.ProfileUserSearchKeyPrefix))
	uhnnLower := strings.ToLower(uhnn)
	paddedUhnn := padToFixedLength(uhnnLower, types.UserSearchFixedLength, types.UserSearchPaddingChar)
	key := append([]byte(paddedUhnn+":"), []byte(address)...)
	store.Delete(key)
}
func (k Keeper) SearchUsersByMatching(ctx sdk.Context, matching string) ([]types.UserSearch, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.ProfileUserSearchKeyPrefix+strings.ToLower(matching)))
	iterator := store.ReverseIterator(nil, nil)
	defer iterator.Close()
	var users []types.UserSearch
	seenAddresses := make(map[string]bool)
	const limit = 10
	count := 0
	for ; iterator.Valid() && count < limit; iterator.Next() {
		val := iterator.Value()
		var user types.UserSearch
		k.cdc.MustUnmarshal(val, &user)
		if !seenAddresses[user.WalletAddress] {
			users = append(users, user)
			seenAddresses[user.WalletAddress] = true
			count++
		}
	}
	return users, true
}

func (k Keeper) SearchUsersLimitByMatching(ctx sdk.Context, matching string, limit int) ([]string, bool) {
	matchingLower := strings.ToLower(matching)
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte{})
	searchPrefix := fmt.Sprintf("%s%s", types.ProfileUserSearchKeyPrefix, matchingLower)
	startKey := []byte(searchPrefix)
	endKey := []byte(fmt.Sprintf("%s\xFF", searchPrefix))
	iterator := store.Iterator(startKey, endKey)
	defer iterator.Close()

	var users []string
	count := 0
	foundWallets := make(map[string]struct{})
	for ; iterator.Valid() && count < limit; iterator.Next() {
		val := iterator.Value()
		var user types.UserSearch
		k.cdc.MustUnmarshal(val, &user)
		if _, exists := foundWallets[user.WalletAddress]; exists {
			continue
		}
		users = append(users, user.WalletAddress)
		count++
		foundWallets[user.WalletAddress] = struct{}{}
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
		types.LogError(k.logger, "get_following_pagination", types.ErrDatabaseOperation, "address", address, "page", page, "limit", limit)
		return nil, nil, types.WrapError(types.ErrDatabaseOperation, "failed to paginate following list")
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

func (k Keeper) AddToFollowingSearch(ctx sdk.Context, followerAddr string, profile types.Profile) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.ProfileFollowingSearchPrefix+followerAddr+"/"))
	targetAddress := profile.WalletAddress
	nicknameKey := append([]byte(strings.ToLower(profile.Nickname)), []byte(targetAddress)...)
	store.Set(nicknameKey, []byte(targetAddress))

	userHandleKey := append([]byte(strings.ToLower(profile.UserHandle)), []byte(targetAddress)...)
	store.Set(userHandleKey, []byte(targetAddress))
}
func (k Keeper) DeleteFromFollowingSearch(ctx sdk.Context, followerAddr string, profile types.Profile) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.ProfileFollowingSearchPrefix+followerAddr+"/"))
	targetAddress := profile.WalletAddress
	nicknameKey := append([]byte(strings.ToLower(profile.Nickname)), []byte(targetAddress)...)
	store.Delete(nicknameKey)

	userHandleKey := append([]byte(strings.ToLower(profile.UserHandle)), []byte(targetAddress)...)
	store.Delete(userHandleKey)
}
func (k Keeper) GetFollowingSearch(ctx sdk.Context, address string, matching string) ([]string, error) {
	matchingLower := strings.ToLower(matching)
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte{})
	searchPrefix := fmt.Sprintf("%s%s/%s", types.ProfileFollowingSearchPrefix, address, matchingLower)
	startKey := []byte(searchPrefix)
	endKey := []byte(fmt.Sprintf("%s\xFF", searchPrefix))

	iterator := store.Iterator(startKey, endKey)
	defer iterator.Close()

	var matchedProfiles []string
	foundWallets := make(map[string]struct{})
	limit := 10
	for ; iterator.Valid() && len(matchedProfiles) < limit; iterator.Next() {
		walletAddress := string(iterator.Value())
		if _, exists := foundWallets[walletAddress]; exists {
			continue
		}
		matchedProfiles = append(matchedProfiles, walletAddress)
		foundWallets[walletAddress] = struct{}{}
	}
	return matchedProfiles, nil
}

func (k Keeper) Unfollow(ctx sdk.Context, followerAddr string, time uint64, targetAddr string) {
	storeFollowing := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.ProfileFollowingPrefix+followerAddr+"/"))
	keyFollowing := append(itob(int64(time)), []byte(targetAddr)...)
	storeFollowing.Delete(keyFollowing)

	storeFollower := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.ProfileFollowersPrefix+targetAddr+"/"))
	keyFollower := append(itob(int64(time)), []byte(followerAddr)...)
	storeFollower.Delete(keyFollower)
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

func (k Keeper) GetFollowersPagination(ctx sdk.Context, address string, page uint64, limit uint64) ([]string, *query.PageResponse, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.ProfileFollowersPrefix+address+"/"))
	first := page * limit
	pagination := &query.PageRequest{}
	pagination.Limit = limit
	pagination.Offset = first
	pagination.Reverse = true
	var followers []string
	pageRes, err := query.Paginate(store, pagination, func(key, value []byte) error {
		followers = append(followers, string(value))
		return nil
	})

	if err != nil {
		types.LogError(k.logger, "get_followers_pagination", types.ErrDatabaseOperation, "address", address, "page", page, "limit", limit)
		return nil, nil, types.WrapError(types.ErrDatabaseOperation, "failed to paginate followers list")
	}
	return followers, pageRes, nil
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
		Reverse:    true,
	}

	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.ActivitiesReceivedPrefix+address+"/"))
	var activitiesList []*types.ActivitiesReceived

	pageResponse, err := query.Paginate(store, pageRequest, func(key []byte, value []byte) error {
		var activity types.ActivitiesReceived
		if err := k.cdc.Unmarshal(value, &activity); err != nil {
			types.LogError(k.logger, "unmarshal_activities_received", err, "address", address)
			return types.WrapError(types.ErrDatabaseOperation, "failed to unmarshal activities received")
		}
		activitiesList = append(activitiesList, &activity)
		return nil
	})

	if err != nil {
		types.LogError(k.logger, "get_activities_received_paginate", err, "address", address, "page", page)
		return nil, nil, uint64(0), types.WrapError(types.ErrDatabaseOperation, "failed to paginate activities received")
	}
	return activitiesList, pageResponse, page, nil
}

func (k Keeper) DeleteLastActivitiesReceived(ctx sdk.Context, wallet string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.ActivitiesReceivedPrefix+wallet+"/"))
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()
	if !iterator.Valid() {
		return
	}
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
		types.LogError(k.logger, "add_admin", types.ErrInvalidRequest, "address", address)
		return types.NewInvalidRequestErrorf("address %s is already an admin", address)
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
		types.LogError(k.logger, "add_editable_admin", types.ErrInvalidRequest, "address", address)
		return types.NewInvalidRequestErrorf("address %s is already an editable admin", address)
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

// GetMessageCount returns the message count between two users
func (k Keeper) GetMessageCount(ctx sdk.Context, receiverAddr string, senderAddr string) int64 {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.ProfileMessageCountPrefix))
	key := []byte(fmt.Sprintf("%s/%s", receiverAddr, senderAddr))
	bz := store.Get(key)
	if bz == nil {
		return 0
	}
	return btoi(bz)
}

// SetMessageCount sets the message count between two users
func (k Keeper) SetMessageCount(ctx sdk.Context, receiverAddr string, senderAddr string, count int64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.ProfileMessageCountPrefix))
	key := []byte(fmt.Sprintf("%s/%s", receiverAddr, senderAddr))
	bz := itob(count)
	store.Set(key, bz)
}

// DeleteOldestMessage deletes the oldest message between two users
func (k Keeper) DeleteOldestMessage(ctx sdk.Context, receiverAddr string, senderAddr string) {
	storePrefix := fmt.Sprintf("%s%s/%s/", types.ProfileMessagePrefix, receiverAddr, senderAddr)
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(storePrefix))

	// Use iterator to find the first (oldest) message
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	if iterator.Valid() {
		oldestKey := iterator.Key()
		store.Delete(oldestKey)
	}
}

// StoreMessage stores a message txHash between sender and receiver
func (k Keeper) StoreMessage(ctx sdk.Context, receiverAddr string, senderAddr string, txHash string) {
	blockTime := ctx.BlockTime().Unix()
	storePrefix := fmt.Sprintf("%s%s/%s/", types.ProfileMessagePrefix, receiverAddr, senderAddr)
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(storePrefix))

	// Use timestamp as key
	messageKey := itob(blockTime)
	store.Set(messageKey, []byte(txHash))
}

// GetMessages retrieves all message txHashes between two users (for query)
func (k Keeper) GetMessages(ctx sdk.Context, receiverAddr string, senderAddr string) []string {
	storePrefix := fmt.Sprintf("%s%s/%s/", types.ProfileMessagePrefix, receiverAddr, senderAddr)
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(storePrefix))

	// Reverse iterator to get latest messages first
	iterator := store.ReverseIterator(nil, nil)
	defer iterator.Close()

	var txHashes []string
	for ; iterator.Valid(); iterator.Next() {
		txHash := string(iterator.Value())
		txHashes = append(txHashes, txHash)
	}

	return txHashes
}
