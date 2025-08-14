package ante

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/sideprotocol/side/x/btcbridge/types"
)

// FeeSponsorshipDecorator implements fee sponsorship
// The gas fee is sponsored if the tx satisfies the sponsorship requirement
// Fallback to the default DeductFeeDecorator otherwise
type FeeSponsorshipDecorator struct {
	accountKeeper   types.AccountKeeper
	bankKeeper      types.BankKeeper
	feegrantKeeper  FeeGrantKeeper
	btcbridgeKeeper BtcBridgeKeeper

	txFeeChecker authante.TxFeeChecker

	defaultFeeDecorator authante.DeductFeeDecorator
}

// NewFeeSponsorshipDecorator creates a new decorator for fee sponsorship
func NewFeeSponsorshipDecorator(authKeeper types.AccountKeeper, bankKeeper types.BankKeeper, feegrantKeeper FeeGrantKeeper, btcbridgeKeeper BtcBridgeKeeper, txFeeChecker authante.TxFeeChecker) FeeSponsorshipDecorator {
	if txFeeChecker == nil {
		txFeeChecker = checkTxFeeWithValidatorMinGasPrices
	}

	return FeeSponsorshipDecorator{
		accountKeeper:       authKeeper,
		bankKeeper:          bankKeeper,
		feegrantKeeper:      feegrantKeeper,
		btcbridgeKeeper:     btcbridgeKeeper,
		txFeeChecker:        txFeeChecker,
		defaultFeeDecorator: authante.NewDeductFeeDecorator(authKeeper, bankKeeper, feegrantKeeper, txFeeChecker),
	}
}

// AnteHandle implements sdk.AnteDecorator
func (fsd FeeSponsorshipDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, errorsmod.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	if !simulate && ctx.BlockHeight() > 0 && feeTx.GetGas() == 0 {
		return ctx, errorsmod.Wrap(sdkerrors.ErrInvalidGasLimit, "must provide positive gas")
	}

	var (
		priority int64
		err      error
	)

	fee := feeTx.GetFee()
	if !simulate {
		fee, priority, err = fsd.txFeeChecker(ctx, tx)
		if err != nil {
			return ctx, err
		}
	}

	if !fsd.checkSponsorship(ctx, tx, fee) {
		// fallback to default fee decorator
		return fsd.defaultFeeDecorator.AnteHandle(ctx, tx, simulate, next)
	}

	// sponsor fee
	if err := fsd.sponsorFee(ctx, fee); err != nil {
		return ctx, err
	}

	newCtx := ctx.WithPriority(priority)

	return next(newCtx, tx, simulate)
}

// sponsorFee performs the fee sponsorship for the given tx
func (fsd FeeSponsorshipDecorator) sponsorFee(ctx sdk.Context, fee sdk.Coins) error {
	sponsorAddress := fsd.getFeeSponsorAddress()

	sponsorAccount := fsd.accountKeeper.GetAccount(ctx, sponsorAddress)
	if sponsorAccount == nil {
		return sdkerrors.ErrUnknownAddress.Wrapf("fee sponsor address: %s does not exist", sponsorAddress)
	}

	// deduct fees from the sponsor account
	if !fee.IsZero() {
		err := authante.DeductFees(fsd.bankKeeper, ctx, sponsorAccount, fee)
		if err != nil {
			return err
		}
	}

	events := sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeTx,
			sdk.NewAttribute(sdk.AttributeKeyFee, fee.String()),
			sdk.NewAttribute(sdk.AttributeKeyFeePayer, sponsorAddress.String()),
		),
	}
	ctx.EventManager().EmitEvents(events)

	return nil
}

// checkSponsorship returns true if the given tx should be sponsored, false otherwise
// NOTE: The following requirements must be satisfied for fee sponsorship
// 1. Fee sponsorship is enabled
// 2. The tx is a BTCT transfer tx
// 3. The gas fee does not exceed the maximum sponsorship fee
// 4. The fee sponsor account has sufficient balances
func (fsd FeeSponsorshipDecorator) checkSponsorship(ctx sdk.Context, tx sdk.Tx, fee sdk.Coins) bool {
	if !fsd.btcbridgeKeeper.FeeSponsorshipEnabled(ctx) {
		return false
	}

	if !fsd.isSponsorableTx(ctx, tx) {
		return false
	}

	if !fee.IsAllLTE(fsd.btcbridgeKeeper.MaxSponsorFee(ctx)) {
		return false
	}

	if !fsd.bankKeeper.SpendableCoins(ctx, fsd.getFeeSponsorAddress()).IsAllGTE(fee) {
		return false
	}

	return true
}

// isSponsorableTx returns true if the given tx is sponsorable, false otherwise
func (fsd FeeSponsorshipDecorator) isSponsorableTx(ctx sdk.Context, tx sdk.Tx) bool {
	for _, m := range tx.GetMsgs() {
		switch msg := m.(type) {
		case *banktypes.MsgSend:
			denoms := msg.Amount.Denoms()
			if len(denoms) != 1 || denoms[0] != fsd.btcbridgeKeeper.BtcDenom(ctx) {
				return false
			}

		case *banktypes.MsgMultiSend:
			denoms := msg.Inputs[0].Coins.Denoms()
			if len(denoms) != 1 || denoms[0] != fsd.btcbridgeKeeper.BtcDenom(ctx) {
				return false
			}

		default:
			return false
		}
	}

	return true
}

// getFeeSponsorAddress gets the fee sponsor address
func (fsd FeeSponsorshipDecorator) getFeeSponsorAddress() sdk.AccAddress {
	return fsd.accountKeeper.GetModuleAddress(types.FeeSponsorName)
}
