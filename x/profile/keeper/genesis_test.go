package keeper_test

import (
	"testing"

	"github.com/rollchains/tlock/x/profile/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	f := SetupTest(t)

	genesisState := &types.GenesisState{
		Params: types.DefaultParams(),
	}

	f.k.InitGenesis(f.ctx, genesisState)

	got := f.k.ExportGenesis(f.ctx)
	require.NotNil(t, got)

}
