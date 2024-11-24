package types

import (
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ sdk.Msg = &MsgUpdateParams{}
)

// NewMsgCreateFreePost creates a new MsgCreateFreePost instance
func NewMsgCreateFreePost(title, content, sender string, timestamp int64) *MsgCreateFreePost {
	return &MsgCreateFreePost{
		Title:     title,
		Content:   content,
		Sender:    sender,
		Timestamp: timestamp,
	}
}

// NewMsgCreatePaidPost creates a new MsgCreatePaidPost instance
func NewMsgCreatePaidPost(title, content, sender string, timestamp int64) *MsgCreatePaidPost {
	return &MsgCreatePaidPost{
		Title:     title,
		Content:   content,
		Sender:    sender,
		Timestamp: timestamp,
	}
}

// NewMsgUpdateParams creates new instance of MsgUpdateParams
func NewMsgUpdateParams(
	sender sdk.Address,
	someValue bool,
) *MsgUpdateParams {
	return &MsgUpdateParams{
		Authority: sender.String(),
		Params: Params{
			SomeValue: someValue,
		},
	}
}

// Route returns the name of the module
func (msg MsgUpdateParams) Route() string { return ModuleName }

// Type returns the the action
func (msg MsgUpdateParams) Type() string { return "update_params" }

// GetSignBytes implements the LegacyMsg interface.
func (msg MsgUpdateParams) GetSignBytes() []byte {
	return sdk.MustSortJSON(AminoCdc.MustMarshalJSON(&msg))
}

// GetSigners returns the expected signers for a MsgUpdateParams message.
func (msg *MsgUpdateParams) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{addr}
}

// ValidateBasic does a sanity check on the provided data.
func (msg *MsgUpdateParams) Validate() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return errors.Wrap(err, "invalid authority address")
	}
	return msg.Params.Validate()
}
