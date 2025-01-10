package keeper

import (
	"context"
	"cosmossdk.io/store/prefix"
	kvtypes "cosmossdk.io/store/types"
	"encoding/binary"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/codec"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"cosmossdk.io/orm/model/ormdb"

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

	authority string
	storeKey  kvtypes.StoreKey
}

// NewKeeper creates a new Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey kvtypes.StoreKey,
	storeService storetypes.KVStoreService,
	logger log.Logger,
	authority string,
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

		storeKey:  storeKey,
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

func (k Keeper) SetProfile(ctx sdk.Context, profile types.Profile) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.ProfileKeyPrefix))
	bz := k.cdc.MustMarshal(&profile)
	store.Set([]byte(profile.WalletAddress), bz)

}

// GetProfile
func (k Keeper) GetProfile(ctx sdk.Context, walletAddress string) (types.Profile, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.ProfileKeyPrefix))
	bz := store.Get([]byte(walletAddress))
	if bz == nil {
		return types.Profile{
			WalletAddress: walletAddress,
			UserHandle:    TruncateWalletAddressSuffix(walletAddress),
		}, true
	}

	var profile types.Profile
	k.cdc.MustUnmarshal(bz, &profile)
	return profile, true
}

func TruncateWalletAddressSuffix(walletAddress string) string {
	runes := []rune(walletAddress)
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
			UserHandle:    TruncateWalletAddressSuffix(walletAddress),
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

// AddFollowing adds the target address to the follower's following list
func (k Keeper) AddFollowing(ctx sdk.Context, followerAddr string, targetAddr string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.ProfileFollowingPrefix+followerAddr+"/"))
	blockTime := ctx.BlockTime().Unix()
	key := append(itob(blockTime), []byte(targetAddr)...)
	store.Set(key, []byte(targetAddr))
}

// GetFollowing returns all following addresses for a given address
func (k Keeper) GetFollowing(ctx sdk.Context, address string) []string {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.ProfileFollowingPrefix+address+"/"))
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	var followings []string
	for ; iterator.Valid(); iterator.Next() {
		val := iterator.Value()
		followings = append(followings, string(val))
	}
	return followings
}

// AddFollower adds the follower address to the target's followers list
func (k Keeper) AddFollower(ctx sdk.Context, targetAddr string, followerAddr string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.ProfileFollowersPrefix+targetAddr+"/"))
	blockTime := ctx.BlockTime().Unix()
	key := append(itob(blockTime), []byte(followerAddr)...)
	store.Set(key, []byte(followerAddr))
}

// GetFollowers returns all follower addresses for a given address
func (k Keeper) GetFollowers(ctx sdk.Context, address string) []string {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.ProfileFollowersPrefix+address+"/"))
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	var followers []string
	for ; iterator.Valid(); iterator.Next() {
		val := iterator.Value()
		followers = append(followers, string(val))
	}
	return followers
}
