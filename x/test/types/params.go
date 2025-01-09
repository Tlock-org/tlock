package types

import (
	"encoding/json"
	"fmt"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// DefaultParams returns default module parameters.
func DefaultParams() Params {
	// TODO:
	return Params{
		SomeValue: true,
	}
}

// Stringer method for Params.
func (p Params) String() string {
	bz, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}

	return string(bz)
}

const (
	ParamStoreKeySomeValue = "someValue"
)

// Validate does the sanity check on the params.
func (p Params) Validate() error {
	// TODO:
	return nil
}

func (p *Params) ParamSetPairs() paramstypes.ParamSetPairs {
	return paramstypes.ParamSetPairs{
		paramstypes.NewParamSetPair([]byte(ParamStoreKeySomeValue), &p.SomeValue, validateSomeValue),
	}
}

func validateSomeValue(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T, expected bool", i)
	}
	return nil
}

func ParamKeyTable() paramstypes.KeyTable {
	return paramstypes.NewKeyTable().RegisterParamSet(&Params{})
}
