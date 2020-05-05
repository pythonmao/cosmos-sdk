package types

import (
	"fmt"
	"github.com/gogo/protobuf/proto"
	"gopkg.in/yaml.v2"

	types "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Governance message types and routes
const (
	TypeMsgDeposit        = "deposit"
	TypeMsgVote           = "vote"
	TypeMsgSubmitProposal = "submit_proposal"
)

var _, _, _ sdk.Msg = MsgSubmitProposal{}, MsgDeposit{}, MsgVote{}
var _ MsgSubmitProposalI = &MsgSubmitProposal{}
var _ types.UnpackInterfacesMessage = MsgSubmitProposal{}

// MsgSubmitProposalI defines the specific interface a concrete message must
// implement in order to process governance proposals. The concrete MsgSubmitProposalLegacy
// must be defined at the application-level.
type MsgSubmitProposalI interface {
	sdk.Msg

	GetContent() Content
	SetContent(Content) error

	GetInitialDeposit() sdk.Coins
	SetInitialDeposit(sdk.Coins)

	GetProposer() sdk.AccAddress
	SetProposer(sdk.AccAddress)
}

// NewMsgSubmitProposal creates a new MsgSubmitProposal.
func NewMsgSubmitProposal(content Content, initialDeposit sdk.Coins, proposer sdk.AccAddress) (*MsgSubmitProposal, error) {
	m := &MsgSubmitProposal{
		InitialDeposit: initialDeposit,
		Proposer:       proposer,
	}
	err := m.SetContent(content)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (msg *MsgSubmitProposal) GetInitialDeposit() sdk.Coins { return msg.InitialDeposit }

func (msg *MsgSubmitProposal) GetProposer() sdk.AccAddress { return msg.Proposer }

func (msg *MsgSubmitProposal) GetContent() Content {
	content, ok := msg.Content.GetCachedValue().(Content)
	if !ok {
		return nil
	}
	return content
}

func (msg *MsgSubmitProposal) SetInitialDeposit(coins sdk.Coins) {
	msg.InitialDeposit = coins
}

func (msg *MsgSubmitProposal) SetProposer(address sdk.AccAddress) {
	msg.Proposer = address
}

func (m *MsgSubmitProposal) SetContent(content Content) error {
	msg, ok := content.(proto.Message)
	if !ok {
		return fmt.Errorf("can't proto marshal %T", msg)
	}
	any, err := types.NewAnyWithValue(msg)
	if err != nil {
		return err
	}
	m.Content = any
	return nil
}

// Route implements Msg
func (msg MsgSubmitProposal) Route() string { return RouterKey }

// Type implements Msg
func (msg MsgSubmitProposal) Type() string { return TypeMsgSubmitProposal }

// ValidateBasic implements Msg
func (msg MsgSubmitProposal) ValidateBasic() error {
	if msg.Proposer.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, msg.Proposer.String())
	}
	if !msg.InitialDeposit.IsValid() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, msg.InitialDeposit.String())
	}
	if msg.InitialDeposit.IsAnyNegative() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, msg.InitialDeposit.String())
	}

	content := msg.GetContent()
	if content == nil {
		return sdkerrors.Wrap(ErrInvalidProposalContent, "missing content")
	}
	if !IsValidProposalType(content.ProposalType()) {
		return sdkerrors.Wrap(ErrInvalidProposalType, content.ProposalType())
	}
	if err := content.ValidateBasic(); err != nil {
		return err
	}

	return nil
}

// GetSignBytes implements Msg
func (msg MsgSubmitProposal) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements Msg
func (msg MsgSubmitProposal) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Proposer}
}

// String implements the Stringer interface
func (msg MsgSubmitProposal) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

func (m MsgSubmitProposal) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	var content Content
	return unpacker.UnpackAny(m.Content, &content)
}

// NewMsgDeposit creates a new MsgDeposit instance
func NewMsgDeposit(depositor sdk.AccAddress, proposalID uint64, amount sdk.Coins) MsgDeposit {
	return MsgDeposit{proposalID, depositor, amount}
}

// Route implements Msg
func (msg MsgDeposit) Route() string { return RouterKey }

// Type implements Msg
func (msg MsgDeposit) Type() string { return TypeMsgDeposit }

// ValidateBasic implements Msg
func (msg MsgDeposit) ValidateBasic() error {
	if msg.Depositor.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, msg.Depositor.String())
	}
	if !msg.Amount.IsValid() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, msg.Amount.String())
	}
	if msg.Amount.IsAnyNegative() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, msg.Amount.String())
	}

	return nil
}

// String implements the Stringer interface
func (msg MsgDeposit) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

// GetSignBytes implements Msg
func (msg MsgDeposit) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements Msg
func (msg MsgDeposit) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Depositor}
}

// NewMsgVote creates a message to cast a vote on an active proposal
func NewMsgVote(voter sdk.AccAddress, proposalID uint64, option VoteOption) MsgVote {
	return MsgVote{proposalID, voter, option}
}

// Route implements Msg
func (msg MsgVote) Route() string { return RouterKey }

// Type implements Msg
func (msg MsgVote) Type() string { return TypeMsgVote }

// ValidateBasic implements Msg
func (msg MsgVote) ValidateBasic() error {
	if msg.Voter.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, msg.Voter.String())
	}
	if !ValidVoteOption(msg.Option) {
		return sdkerrors.Wrap(ErrInvalidVote, msg.Option.String())
	}

	return nil
}

// String implements the Stringer interface
func (msg MsgVote) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

// GetSignBytes implements Msg
func (msg MsgVote) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements Msg
func (msg MsgVote) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Voter}
}

// ---------------------------------------------------------------------------
// Deprecated
//
// TODO: Remove once client-side Protobuf migration has been completed.
// ---------------------------------------------------------------------------

// MsgSubmitProposalLegacy defines a (deprecated) message to create/submit a governance
// proposal.
//
// TODO: Remove once client-side Protobuf migration has been completed.
type MsgSubmitProposalLegacy struct {
	Content        Content        `json:"content" yaml:"content"`
	InitialDeposit sdk.Coins      `json:"initial_deposit" yaml:"initial_deposit"` //  Initial deposit paid by sender. Must be strictly positive
	Proposer       sdk.AccAddress `json:"proposer" yaml:"proposer"`               //  Address of the proposer
}

var _ MsgSubmitProposalI = &MsgSubmitProposalLegacy{}

// NewMsgSubmitProposalLegacy returns a (deprecated) MsgSubmitProposalLegacy message.
//
// TODO: Remove once client-side Protobuf migration has been completed.
func NewMsgSubmitProposalLegacy(content Content, initialDeposit sdk.Coins, proposer sdk.AccAddress) *MsgSubmitProposalLegacy {
	return &MsgSubmitProposalLegacy{content, initialDeposit, proposer}
}

// ValidateBasic implements Msg
func (msg MsgSubmitProposalLegacy) ValidateBasic() error {
	if msg.Content == nil {
		return sdkerrors.Wrap(ErrInvalidProposalContent, "missing content")
	}
	if msg.Proposer.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, msg.Proposer.String())
	}
	if !msg.InitialDeposit.IsValid() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, msg.InitialDeposit.String())
	}
	if msg.InitialDeposit.IsAnyNegative() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, msg.InitialDeposit.String())
	}
	if !IsValidProposalType(msg.Content.ProposalType()) {
		return sdkerrors.Wrap(ErrInvalidProposalType, msg.Content.ProposalType())
	}

	return msg.Content.ValidateBasic()
}

// GetSignBytes implements Msg
func (msg MsgSubmitProposalLegacy) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// nolint
func (msg MsgSubmitProposalLegacy) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Proposer}
}
func (msg MsgSubmitProposalLegacy) Route() string                { return RouterKey }
func (msg MsgSubmitProposalLegacy) Type() string                 { return TypeMsgSubmitProposal }
func (msg MsgSubmitProposalLegacy) GetContent() Content          { return msg.Content }
func (msg MsgSubmitProposalLegacy) GetInitialDeposit() sdk.Coins { return msg.InitialDeposit }
func (msg MsgSubmitProposalLegacy) GetProposer() sdk.AccAddress  { return msg.Proposer }

func (msg *MsgSubmitProposalLegacy) SetContent(content Content) error {
	msg.Content = content
	return nil
}

func (msg *MsgSubmitProposalLegacy) SetInitialDeposit(deposit sdk.Coins) {
	msg.InitialDeposit = deposit
}

func (msg *MsgSubmitProposalLegacy) SetProposer(proposer sdk.AccAddress) {
	msg.Proposer = proposer
}
