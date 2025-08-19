package types

import (
	"encoding/json"
)

// DefaultParams returns default module parameters.
func DefaultParams() Params {
	return Params{
		SomeValue: true,
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
	// Add proper validation logic for SomeValue
	// Example validation (replace with actual requirements):
	// if p.SomeValue == nil {
	//     return errors.New("SomeValue cannot be nil")
	// }

	// Return nil for now as the current param is just a placeholder
	return nil
}
