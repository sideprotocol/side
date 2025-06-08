package keeper

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	btcschnorr "github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil/psbt"
	"github.com/btcsuite/btcd/txscript"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/bitcoin/crypto/schnorr"
	dlctypes "github.com/sideprotocol/side/x/dlc/types"
	"github.com/sideprotocol/side/x/lending/types"
	tsstypes "github.com/sideprotocol/side/x/tss/types"
)

// Apply implements types.MsgServer.
func (m msgServer) Apply(goCtx context.Context, msg *types.MsgApply) (*types.MsgApplyResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !m.HasPool(ctx, msg.PoolId) {
		return nil, types.ErrPoolDoesNotExist
	}

	pool := m.GetPool(ctx, msg.PoolId)
	if pool.Status != types.PoolStatus_ACTIVE {
		return nil, types.ErrPoolNotActive
	}

	poolConfig := m.GetPool(ctx, msg.PoolId).Config

	if msg.BorrowAmount.Denom != pool.Supply.Denom {
		return nil, errorsmod.Wrap(types.ErrInvalidAmount, "mismatched denom")
	}

	if err := types.CheckBorrowAmountLimit(pool, msg.BorrowAmount.Amount); err != nil {
		return nil, err
	}

	if err := types.CheckBorrowCap(pool, msg.BorrowAmount.Amount); err != nil {
		return nil, err
	}

	if msg.BorrowAmount.Amount.GT(pool.AvailableAmount) {
		return nil, types.ErrInsufficientLiquidity
	}

	trancheConfig, found := types.GetTrancheConfig(poolConfig.Tranches, msg.Maturity)
	if !found {
		return nil, errorsmod.Wrap(types.ErrInvalidMaturity, "maturity does not exist")
	}

	if types.HasRequestFee(pool) {
		if err := m.bankKeeper.SendCoins(ctx, sdk.MustAccAddressFromBech32(msg.Borrower), sdk.MustAccAddressFromBech32(m.RequestFeeCollector(ctx)), sdk.NewCoins(poolConfig.RequestFee)); err != nil {
			return nil, err
		}
	}

	if !m.dlcKeeper.HasDCM(ctx, msg.DCMId) {
		return nil, errorsmod.Wrap(types.ErrInvalidDCM, "dcm does not exist")
	}

	dcm := m.dlcKeeper.GetDCM(ctx, msg.DCMId)

	originMaturityTime := ctx.BlockTime().Add(time.Duration(trancheConfig.Maturity) * time.Second).Unix()
	maturityTime := types.GetMaturityTime(originMaturityTime)
	finalTimeout := originMaturityTime + m.FinalTimeoutDuration(ctx)

	vault, err := types.CreateVaultAddress(msg.BorrowerPubkey, dcm.Pubkey, finalTimeout)
	if err != nil {
		return nil, err
	}

	if m.HasLoan(ctx, vault) {
		return nil, types.ErrDuplicatedVault
	}

	if !m.dlcKeeper.HasEventByDate(ctx, maturityTime) {
		return nil, errorsmod.Wrap(types.ErrInvalidEvent, "default liquidation event does not exist")
	}

	defaultLiquidationEvent := m.dlcKeeper.GetEventByDate(ctx, maturityTime)

	repaymentEvent := m.dlcKeeper.GetAvailableLendingEvent(ctx)
	if repaymentEvent == nil {
		return nil, errorsmod.Wrap(types.ErrInvalidEvent, "no available event for repayment")
	}

	// update repayment event
	repaymentEvent.Description = fmt.Sprintf("repayment event for loan %s", vault)
	repaymentEvent.Outcomes = []string{vault}
	m.dlcKeeper.SetEvent(ctx, repaymentEvent)

	interest := types.GetTotalInterest(msg.BorrowAmount.Amount, trancheConfig.Maturity, trancheConfig.BorrowAPR, m.GetBlocksPerYear(ctx))
	protocolFee := types.GetProtocolFee(interest, poolConfig.ReserveFactor)

	loan := &types.Loan{
		VaultAddress:              vault,
		Borrower:                  msg.Borrower,
		BorrowerPubKey:            msg.BorrowerPubkey,
		DCM:                       dcm.Pubkey,
		MaturityTime:              maturityTime,
		FinalTimeout:              finalTimeout,
		PoolId:                    msg.PoolId,
		BorrowAmount:              msg.BorrowAmount,
		RequestFee:                poolConfig.RequestFee,
		OriginationFee:            poolConfig.OriginationFee,
		Interest:                  interest,
		ProtocolFee:               protocolFee,
		Maturity:                  trancheConfig.Maturity,
		BorrowAPR:                 trancheConfig.BorrowAPR,
		MinMaturity:               trancheConfig.Maturity * int64(trancheConfig.MinMaturityFactor) / 1000,
		DefaultLiquidationEventId: defaultLiquidationEvent.Id,
		RepaymentEventId:          repaymentEvent.Id,
		Referrer:                  msg.Referrer,
		CreateAt:                  ctx.BlockTime(),
		Status:                    types.LoanStatus_Requested,
	}

	m.SetLoan(ctx, loan)
	m.SetLoanByAddress(ctx, loan)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.EventTypeApply,
			sdk.NewAttribute(types.AttributeKeyVault, loan.VaultAddress),
			sdk.NewAttribute(types.AttributeKeyBorrower, loan.Borrower),
			sdk.NewAttribute(types.AttributeKeyDCMPubKey, loan.DCM),
			sdk.NewAttribute(types.AttributeKeyMuturityTime, fmt.Sprint(loan.MaturityTime)),
			sdk.NewAttribute(types.AttributeKeyFinalTimeout, fmt.Sprint(loan.FinalTimeout)),
			sdk.NewAttribute(types.AttributeKeyPoolId, loan.PoolId),
			sdk.NewAttribute(types.AttributeKeyBorrowAmount, loan.BorrowAmount.String()),
		))

	return &types.MsgApplyResponse{}, nil
}

// SubmitCets implements types.MsgServer.
func (m msgServer) SubmitCets(goCtx context.Context, msg *types.MsgSubmitCets) (*types.MsgSubmitCetsResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !m.HasLoan(ctx, msg.LoanId) {
		return nil, types.ErrLoanDoesNotExist
	}

	loan := m.GetLoan(ctx, msg.LoanId)
	if msg.Borrower != loan.Borrower {
		return nil, types.ErrMismatchedBorrower
	}

	// NOTE: only can be authorized once for now
	if loan.Status != types.LoanStatus_Requested {
		return nil, errorsmod.Wrap(types.ErrInvalidLoanStatus, "loan non requested")
	}

	pool := m.GetPool(ctx, loan.PoolId)
	poolConfig := pool.Config

	vaultPkScript, _ := types.GetPkScriptFromAddress(loan.VaultAddress)

	depositTxs := []*psbt.Packet{}
	depositTxHashes := []string{}
	collateralAmount := sdkmath.ZeroInt()

	for _, depositTx := range msg.DepositTxs {
		p, _ := psbt.NewFromRawBytes(bytes.NewReader([]byte(depositTx)), true)

		depositTxs = append(depositTxs, p)
		depositTxHashes = append(depositTxHashes, p.UnsignedTx.TxHash().String())

		for _, out := range p.UnsignedTx.TxOut {
			if bytes.Equal(out.PkScript, vaultPkScript) {
				collateralAmount = collateralAmount.Add(sdkmath.NewInt(out.Value))
			}
		}
	}

	if collateralAmount.IsZero() {
		return nil, errorsmod.Wrap(types.ErrInsufficientCollateral, "collateral amount can not be zero")
	}

	// build DLC metadata
	dlcMeta, err := types.BuildDLCMeta(depositTxs, vaultPkScript, msg.LiquidationCet, msg.LiquidationAdaptorSignatures, msg.DefaultLiquidationAdaptorSignatures, msg.RepaymentCet, msg.RepaymentSignatures, loan.BorrowerPubKey, loan.DCM, loan.MaturityTime, loan.FinalTimeout)
	if err != nil {
		return nil, err
	}

	// create authorization
	authorization := m.CreateAuthorization(ctx, msg.LoanId, depositTxHashes)

	for i, depositTx := range msg.DepositTxs {
		if !m.HasDepositLog(ctx, depositTxHashes[i]) {
			depositLog := &types.DepositLog{
				Txid:            depositTxHashes[i],
				VaultAddress:    loan.VaultAddress,
				AuthorizationId: authorization.Id,
				DepositTx:       depositTx,
			}

			m.SetDepositLog(ctx, depositLog)
		}
	}

	m.SetDLCMeta(ctx, loan.VaultAddress, dlcMeta)

	loan.CollateralAmount = collateralAmount
	loan.Authorizations = append(loan.Authorizations, *authorization)

	var errRejected error

	defer func() {
		if errRejected != nil {
			loan.Authorizations[authorization.Id-1].Status = types.AuthorizationStatus_AUTHORIZATION_STATUS_REJECTED
			loan.Status = types.LoanStatus_Rejected
			m.SetLoan(ctx, loan)

			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeReject,
					sdk.NewAttribute(types.AttributeKeyLoanId, msg.LoanId),
					sdk.NewAttribute(types.AttributeKeyAuthorizationId, fmt.Sprintf("%d", authorization.Id)),
					sdk.NewAttribute(types.AttributeKeyReason, errRejected.Error()),
				),
			)
		}
	}()

	pricePair := types.GetPricePair(poolConfig)
	collateralDecimals := int(poolConfig.CollateralAsset.Decimals)
	borrowDecimals := int(poolConfig.LendingAsset.Decimals)
	collateralIsBaseAsset := poolConfig.CollateralAsset.IsBasePriceAsset

	dlcPricePair, found := m.dlcKeeper.PricePair(ctx, pricePair)
	if !found {
		errRejected = errorsmod.Wrap(types.ErrInvalidPricePair, "price pair does not exist in dlc")
		return nil, nil
	}

	liquidationPrice := types.GetLiquidationPrice(collateralAmount, collateralDecimals, loan.BorrowAmount.Amount, borrowDecimals, loan.Maturity, loan.BorrowAPR, m.GetBlocksPerYear(ctx), poolConfig.LiquidationThreshold, int(dlcPricePair.Decimals), dlcPricePair.Interval, collateralIsBaseAsset)
	normalizedLiquidationPrice := dlctypes.NormalizePrice(liquidationPrice, int(dlcPricePair.Decimals))

	if !m.dlcKeeper.HasEventByPrice(ctx, pricePair, normalizedLiquidationPrice) {
		errRejected = errorsmod.Wrap(types.ErrInvalidEvent, "liquidation event does not exist")
		return nil, nil
	}

	liquidationEvent := m.dlcKeeper.GetEventByPrice(ctx, pricePair, normalizedLiquidationPrice)
	if liquidationEvent.HasTriggered {
		errRejected = errorsmod.Wrap(types.ErrInvalidEvent, "liquidation event has triggered")
		return nil, nil
	}

	defaultLiquidationEvent := m.dlcKeeper.GetEvent(ctx, loan.DefaultLiquidationEventId)

	if err := types.VerifyCets(depositTxs, vaultPkScript, loan.BorrowerPubKey, loan.DCM, liquidationEvent, defaultLiquidationEvent, msg.LiquidationCet, msg.LiquidationAdaptorSignatures, msg.DefaultLiquidationAdaptorSignatures, msg.RepaymentCet, msg.RepaymentSignatures); err != nil {
		return nil, err
	}

	loan.LiquidationPrice = liquidationPrice
	loan.LiquidationEventId = liquidationEvent.Id

	loan.Status = types.LoanStatus_Authorized
	m.SetLoan(ctx, loan)

	return &types.MsgSubmitCetsResponse{}, nil
}

// SubmitDepositTransaction implements types.MsgServer.
func (m msgServer) SubmitDepositTransaction(goCtx context.Context, msg *types.MsgSubmitDepositTransaction) (*types.MsgSubmitDepositTransactionResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !m.HasLoan(ctx, msg.Vault) {
		return nil, errorsmod.Wrap(types.ErrInvalidVault, "vault does not match any loan")
	}

	loan := m.GetLoan(ctx, msg.Vault)
	if loan.Status != types.LoanStatus_Requested && loan.Status != types.LoanStatus_Authorized && loan.Status != types.LoanStatus_Rejected {
		return nil, types.ErrInvalidLoanStatus
	}

	// validate deposit tx
	tx, _, err := m.btcbridgeKeeper.ValidateTransaction(ctx, msg.DepositTx, "", msg.BlockHash, msg.Proof, m.btcbridgeKeeper.DepositConfirmationDepth(ctx))
	if err != nil {
		return nil, errorsmod.Wrapf(types.ErrInvalidDepositTx, "failed to validate tx: %v", err)
	}

	depositTxHash := tx.Hash().String()

	var depositLog *types.DepositLog
	if m.HasDepositLog(ctx, depositTxHash) {
		depositLog = m.GetDepositLog(ctx, depositTxHash)
		if depositLog.Status != types.DepositStatus_DEPOSIT_STATUS_PENDING {
			return nil, errorsmod.Wrap(types.ErrInvalidDepositTx, "deposit tx not pending")
		}
	} else {
		depositLog = &types.DepositLog{
			Txid:            depositTxHash,
			VaultAddress:    msg.Vault,
			DepositTx:       msg.DepositTx,
			AuthorizationId: m.GetAuthorizationId(ctx, msg.Vault) + 1,
		}
	}

	depositLog.Status = types.DepositStatus_DEPOSIT_STATUS_VERIFIED
	m.SetDepositLog(ctx, depositLog)

	return &types.MsgSubmitDepositTransactionResponse{}, nil
}

// Redeem implements types.MsgServer.
func (m msgServer) Redeem(goCtx context.Context, msg *types.MsgRedeem) (*types.MsgRedeemResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !m.HasLoan(ctx, msg.LoanId) {
		return nil, types.ErrLoanDoesNotExist
	}

	loan := m.GetLoan(ctx, msg.LoanId)
	if msg.Borrower != loan.Borrower {
		return nil, types.ErrMismatchedBorrower
	}

	if loan.Status != types.LoanStatus_Rejected {
		return nil, errorsmod.Wrap(types.ErrInvalidLoanStatus, "loan not rejected")
	}

	p, _ := psbt.NewFromRawBytes(bytes.NewReader([]byte(msg.Tx)), true)

	borrowerPubKey, _ := hex.DecodeString(loan.BorrowerPubKey)
	dcmPubKey, _ := hex.DecodeString(loan.DCM)

	internalKey := types.GetInternalKey(borrowerPubKey, dcmPubKey)

	script, _ := hex.DecodeString(m.GetDLCMeta(ctx, msg.LoanId).MultisigScript)
	sigHashes := []string{}

	merkleTree := types.GetTapscriptTree(types.GetDLCTapscripts(m.GetDLCMeta(ctx, msg.LoanId)))
	scriptProof := merkleTree.LeafMerkleProofs[0]

	controlBlock, err := types.GetControlBlock(internalKey, scriptProof)
	if err != nil {
		return nil, err
	}

	for i, ti := range p.UnsignedTx.TxIn {
		prevTxHash := ti.PreviousOutPoint.Hash.String()
		if !m.HasDepositLog(ctx, prevTxHash) {
			return nil, types.ErrDepositTxDoesNotExist
		}

		depositLog := m.GetDepositLog(ctx, prevTxHash)
		if depositLog.VaultAddress != msg.LoanId {
			return nil, errorsmod.Wrap(types.ErrInvalidDepositTx, "deposit tx does not match the loan id")
		}

		// check deposit status
		if depositLog.Status != types.DepositStatus_DEPOSIT_STATUS_VERIFIED {
			return nil, errorsmod.Wrap(types.ErrInvalidDepositTx, "deposit tx non verified")
		}

		sigBytes, _ := hex.DecodeString(msg.Signatures[i])

		sigHash, err := types.CalcTapscriptSigHash(p, i, types.DefaultSigHashType, script)
		if err != nil {
			return nil, err
		}

		if !schnorr.Verify(sigBytes, sigHash, borrowerPubKey) {
			return nil, types.ErrInvalidSignature
		}

		sigHashes = append(sigHashes, base64.StdEncoding.EncodeToString(sigHash))

		p.Inputs[i].TaprootInternalKey = btcschnorr.SerializePubKey(internalKey)
		p.Inputs[i].TaprootLeafScript = []*psbt.TaprootTapLeafScript{
			{
				ControlBlock: controlBlock,
				Script:       script,
				LeafVersion:  txscript.BaseLeafVersion,
			},
		}

		// update deposit status
		depositLog.Status = types.DepositStatus_DEPOSIT_STATUS_REDEEMING
		m.SetDepositLog(ctx, depositLog)
	}

	serializedTx, err := p.B64Encode()
	if err != nil {
		return nil, err
	}

	redemption := &types.Redemption{
		Id:         m.IncrementRedemptionId(ctx),
		LoanId:     msg.LoanId,
		Txid:       p.UnsignedTx.TxHash().String(),
		Tx:         serializedTx,
		Signatures: msg.Signatures,
		CreateAt:   ctx.BlockTime(),
	}
	m.SetRedemption(ctx, redemption)

	m.tssKeeper.InitiateSigningRequest(
		ctx,
		types.ModuleName,
		types.ToScopedId(redemption.Id),
		tsstypes.SigningType_SIGNING_TYPE_SCHNORR,
		int32(types.SigningIntent_SIGNING_INTENT_REDEMPTION),
		loan.DCM,
		sigHashes,
		nil,
	)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRedeem,
			sdk.NewAttribute(types.AttributeKeyBorrower, msg.Borrower),
			sdk.NewAttribute(types.AttributeKeyLoanId, msg.LoanId),
			sdk.NewAttribute(types.AttributeKeyId, fmt.Sprintf("%d", redemption.Id)),
		),
	)

	return &types.MsgRedeemResponse{}, nil
}

// Repay implements types.MsgServer.
func (m msgServer) Repay(goCtx context.Context, msg *types.MsgRepay) (*types.MsgRepayResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !m.HasLoan(ctx, msg.LoanId) {
		return nil, types.ErrLoanDoesNotExist
	}

	loan := m.GetLoan(ctx, msg.LoanId)
	if loan.Status != types.LoanStatus_Open {
		return nil, errorsmod.Wrap(types.ErrInvalidLoanStatus, "loan not open")
	}

	if ctx.BlockTime().Unix()-loan.CreateAt.Unix() < loan.MinMaturity {
		return nil, types.ErrMinMaturityNotReached
	}

	interest := m.GetCurrentInterest(ctx, loan)
	amount := loan.BorrowAmount.Add(interest)

	// escrow repaid amount
	if err := m.bankKeeper.SendCoinsFromAccountToModule(ctx, sdk.MustAccAddressFromBech32(msg.Borrower), types.RepaymentEscrowAccount, sdk.NewCoins(amount)); err != nil {
		return nil, err
	}

	loan.Status = types.LoanStatus_Repaid
	m.SetLoan(ctx, loan)

	repayment := &types.Repayment{
		LoanId:   msg.LoanId,
		Amount:   amount,
		CreateAt: ctx.BlockTime(),
	}
	m.SetRepayment(ctx, repayment)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRepay,
			sdk.NewAttribute(types.AttributeKeyBorrower, msg.Borrower),
			sdk.NewAttribute(types.AttributeKeyLoanId, msg.LoanId),
			sdk.NewAttribute(types.AttributeKeyAmount, amount.String()),
		),
	)

	return &types.MsgRepayResponse{}, nil
}
