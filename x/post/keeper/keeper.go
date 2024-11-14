package keeper

import (
	"context"
	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	kvtypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/feegrant"
	feegrantkeeper "cosmossdk.io/x/feegrant/keeper"
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"time"

	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"cosmossdk.io/orm/model/ormdb"
	apiv1 "github.com/rollchains/tlock/api/post/v1"
	"github.com/rollchains/tlock/x/post/types"

	cosmossdk_io_math "cosmossdk.io/math"
)

type Keeper struct {
	cdc codec.BinaryCodec

	logger log.Logger

	// state management
	Schema collections.Schema
	Params collections.Item[types.Params]
	OrmDB  apiv1.StateStore

	NameMapping   collections.Map[string, string]
	storeKey      kvtypes.StoreKey
	AccountKeeper authkeeper.AccountKeeper
	bankKeeper    bankkeeper.Keeper
	feeKeeper     feegrantkeeper.Keeper
	authority     string
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
	feeKeeper feegrantkeeper.Keeper,
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

		NameMapping:   collections.NewMap(sb, collections.NewPrefix(1), "name_mapping", collections.StringKey, collections.StringValue),
		storeKey:      storeKey,
		AccountKeeper: ak,
		bankKeeper:    bk,

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

	fmt.Printf("==============k.storeKey %s", k.storeKey.Name())

	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.PostKeyPrefix))
	bz := k.cdc.MustMarshal(&post)
	store.Set([]byte(post.Id), bz)

}

func (k Keeper) PostReward(ctx sdk.Context, post types.Post) {
	// send post reward
	amount := sdk.NewCoins(sdk.NewCoin("TOK", cosmossdk_io_math.NewInt(10)))
	//address := sdk.AccAddress([]byte(post.Sender))
	userAddr, err := sdk.AccAddressFromBech32(post.Sender)

	if err != nil {
		return
	}

	err1 := k.SendCoinsToUser(ctx, userAddr, amount)
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
func (k Keeper) generatePostID(ctx sdk.Context) string {
	// Use the current block time and a counter or other mechanism
	timestamp := ctx.BlockTime().UnixNano()
	return fmt.Sprintf("post-%d", timestamp)

}

// get post module address
func (k Keeper) GetModuleAccountAddress() sdk.AccAddress {
	return k.AccountKeeper.GetModuleAddress(types.ModuleName)
}

// transfer from module to user
func (k Keeper) SendCoinsToUser(ctx sdk.Context, userAddr sdk.AccAddress, amount sdk.Coins) error {
	moduleAddress := k.AccountKeeper.GetModuleAddress(types.ModuleName)
	fmt.Printf("==============types.ModuleName: [%s], =moduleAddress: [%s]", types.ModuleName, moduleAddress)
	return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, userAddr, amount)
}

func (k Keeper) ApproveFeegrant(ctx sdk.Context, userAddr sdk.AccAddress) {
	now := ctx.BlockTime()
	oneHour := now.Add(1 * time.Hour)
	oneDay := now.Add(24 * time.Hour)
	period := 24 * time.Hour
	spendLimit := sdk.NewCoins(sdk.NewCoin("TOK", sdkmath.NewInt(3)))
	// create a basic allowance
	allowance := feegrant.BasicAllowance{
		spendLimit,
		&oneHour,
	}

	// create a periodic allowance
	periodicAllowance := &feegrant.PeriodicAllowance{
		allowance,
		period,
		spendLimit,
		spendLimit,
		oneDay,
	}

	k.feeKeeper.GrantAllowance(ctx, k.AccountKeeper.GetModuleAddress(types.ModuleName), userAddr, periodicAllowance)

}
