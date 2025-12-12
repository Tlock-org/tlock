package module

import (
	kvtypes "cosmossdk.io/store/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"os"

	"github.com/cosmos/cosmos-sdk/codec"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

	"cosmossdk.io/core/address"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/store"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"

	modulev1 "github.com/rollchains/tlock/api/profile/module/v1"
	"github.com/rollchains/tlock/x/profile/keeper"
)

var _ appmodule.AppModule = AppModule{}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() {}

// IsAppModule implements the appmodule.AppModule interface.
func (am AppModule) IsAppModule() {}

func init() {
	appmodule.Register(
		&modulev1.Module{},
		appmodule.Provide(ProvideModule),
	)
}

type ModuleInputs struct {
	depinject.In

	Cdc          codec.Codec
	storeKey     kvtypes.StoreKey
	StoreService store.KVStoreService
	AddressCodec address.Codec

	AccountKeeper  authkeeper.AccountKeeper
	StakingKeeper  stakingkeeper.Keeper
	SlashingKeeper slashingkeeper.Keeper
	paramSpace     paramstypes.Subspace
}

type ModuleOutputs struct {
	depinject.Out

	Module appmodule.AppModule
	Keeper keeper.Keeper
}

func ProvideModule(in ModuleInputs) ModuleOutputs {
	govAddr := authtypes.NewModuleAddress(govtypes.ModuleName).String()

	k := keeper.NewKeeper(in.Cdc, in.storeKey, in.StoreService, log.NewLogger(os.Stderr), govAddr, in.paramSpace, in.AccountKeeper)
	m := NewAppModule(in.Cdc, k)

	return ModuleOutputs{Module: m, Keeper: k, Out: depinject.Out{}}
}
