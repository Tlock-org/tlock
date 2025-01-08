package types

import (
	"encoding/json"
	"fmt"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// ParamSetPairs defines the parameter key-value pairs and their validators.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair([]byte(KeySomeValue), &p.SomeValue, validateSomeValue),
	}
}
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

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

// Validate does the sanity check on the params.
func (p Params) Validate() error {
	// TODO:
	return nil
}

// validateSomeValue validates the SomeValue parameter.
func validateSomeValue(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type for SomeValue: expected bool")
	}
	return nil
}
