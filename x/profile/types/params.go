package types

import (
	"encoding/json"
	"fmt"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

const (
	KeySomeValue      = "someValue"
	KeyAdminAddress   = "adminAddress"
	KeyChiefModerator = "chiefModerator"
)

// DefaultParams returns default module parameters.
func DefaultParams() Params {
	// TODO:
	return Params{
		SomeValue:      true,
		AdminAddress:   "tlock1hj5fveer5cjtn4wd6wstzugjfdxzl0xp5u7j9p",
		ChiefModerator: "tlock1hj5fveer5cjtn4wd6wstzugjfdxzl0xp5u7j9p",
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

func ParamKeyTable() paramstypes.KeyTable {
	return paramstypes.NewKeyTable().RegisterParamSet(&Params{})
}

func (p *Params) ParamSetPairs() paramstypes.ParamSetPairs {
	return paramstypes.ParamSetPairs{
		paramstypes.NewParamSetPair([]byte(KeySomeValue), &p.SomeValue, validateSomeValue),
		paramstypes.NewParamSetPair([]byte(KeyAdminAddress), &p.AdminAddress, validateAddress),
		paramstypes.NewParamSetPair([]byte(KeyChiefModerator), &p.ChiefModerator, validateAddress),
	}
}

func validateSomeValue(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T, expected bool", i)
	}
	return nil
}

func validateAddress(i interface{}) error {
	_, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T, expected string", i)
	}
	return nil
}
