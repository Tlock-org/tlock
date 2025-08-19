package types

import (
	"encoding/json"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"strings"
)

const (
	KeySomeValue      = "someValue"
	KeyAdminAddress   = "adminAddress"
	KeyChiefModerator = "chiefModerator"
)

// DefaultParams returns default module parameters.
func DefaultParams() Params {
	return Params{
		SomeValue:      true,
		AdminAddress:   "tlock1hj5fveer5cjtn4wd6wstzugjfdxzl0xp5u7j9p",
		ChiefModerator: "tlock1wfvjqmkekyuy59r535nm2ca3yjkf706nu8x49r",
	}
}

// Stringer method for Params.
func (p Params) String() string {
	bz, err := json.Marshal(p)
	if err != nil {
		// Return a descriptive error string instead of panic
		return "failed to marshal params to string"
	}

	return string(bz)
}

// Validate does the sanity check on the params.
func (p Params) Validate() error {
	if err := validateSomeValue(p.SomeValue); err != nil {
		return WrapErrorf(ErrInvalidParameter, "invalid SomeValue: %v", err)
	}

	if err := validateAddress(p.AdminAddress); err != nil {
		return WrapErrorf(ErrInvalidParameter, "invalid AdminAddress: %v", err)
	}

	if err := validateAddress(p.ChiefModerator); err != nil {
		return WrapErrorf(ErrInvalidParameter, "invalid ChiefModerator: %v", err)
	}

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
		return NewInvalidParameterErrorf("invalid parameter type: %T, expected bool", i)
	}
	return nil
}

func validateAddress(i interface{}) error {
	addr, ok := i.(string)
	if !ok {
		return NewInvalidParameterErrorf("invalid parameter type: %T, expected string", i)
	}

	addr = strings.TrimSpace(addr)
	if addr == "" {
		return NewInvalidParameterErrorf("address cannot be empty")
	}

	// Validate bech32 address format
	_, err := sdk.AccAddressFromBech32(addr)
	if err != nil {
		return NewInvalidParameterErrorf("invalid bech32 address format: %v", err)
	}

	return nil
}

// NewInvalidParameterErrorf creates a new invalid parameter error with formatting
func NewInvalidParameterErrorf(format string, args ...interface{}) error {
	return WrapErrorf(ErrInvalidParameter, format, args...)
}
