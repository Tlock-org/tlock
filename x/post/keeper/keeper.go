package keeper

import (
	"context"
	"cosmossdk.io/store/prefix"
	"cosmossdk.io/x/feegrant"
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"time"

	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

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

	fmt.Printf("==============k.storeKey %s", k.storeKey.Name())

	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.PostKeyPrefix))
	bz := k.cdc.MustMarshal(&post)
	store.Set([]byte(post.Id), bz)

}

func (k Keeper) PostReward(ctx sdk.Context, post types.Post) {
	// send post reward
	amount := sdk.NewCoins(sdk.NewCoin("TOK", sdkmath.NewInt(10)))
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
	//oneHour := now.Add(1 * time.Hour)
	oneDay := now.Add(24 * time.Hour)
	//period := 24 * time.Hour
	period := 2 * time.Minute
	totalSpendLimit := sdk.NewCoins(sdk.NewCoin("TOK", sdkmath.NewInt(5)))
	spendLimit := sdk.NewCoins(sdk.NewCoin("TOK", sdkmath.NewInt(2)))
	// create a basic allowance
	allowance := feegrant.BasicAllowance{
		SpendLimit: totalSpendLimit,
		Expiration: &oneDay,
	}

	// create a periodic allowance
	periodicAllowance := &feegrant.PeriodicAllowance{
		Basic:            allowance,
		Period:           period,
		PeriodSpendLimit: spendLimit,
		PeriodCanSpend:   spendLimit,
		PeriodReset:      now.Add(period),
	}

	granter := k.AccountKeeper.GetModuleAddress(types.ModuleName)
	//granter, _ := sdk.AccAddressFromBech32("tlock1efd63aw40lxf3n4mhf7dzhjkr453axurggdkvg")
	grantee := userAddr

	err := k.FeeGrantKeeper.GrantAllowance(ctx, granter, grantee, periodicAllowance)
	if err != nil {
		ctx.Logger().Error("Failed to grant allowance", "error", err)
		return
	}
}
