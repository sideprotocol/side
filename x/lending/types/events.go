package types

// Lending module event types
const (
	EventTypeApply                                 = "apply"
	EventTypeCancel                                = "cancel"
	EventTypeRepay                                 = "repay"
	EventTypeDefault                               = "default"
	EventTypeLiquidate                             = "liquidate"
	EventTypeSignRepaymentCet                      = "sign_repayment_cet"
	EventTypeGenerateSignedLiquidationCet          = "generate_signed_liquidation_cet"
	EventTypeGenerateSignedCancellationTransaction = "generate_signed_cancellation_transaction"

	AttributeKeyVault            = "vault"
	AttributeKeyBorrower         = "borrower"
	AttributeKeyAgencyPubKey     = "agency_pub_key"
	AttributeKeyMuturityTime     = "muturity_time"
	AttributeKeyFinalTimeout     = "final_timeout"
	AttributeKeyCollateralAmount = "collateral_amount"
	AttributeKeyBorrowAmount     = "borrow_amount"
	AttributeKeyPoolId           = "pool_id"
	AttributeKeyEventId          = "event_id"

	AttributeKeyLoanId = "loan_id"

	AttributeKeyAdaptorPoint    = "adaptor_point"
	AtrtibuteKeyRepaymentTxHash = "repayment_tx_hash"

	AttributeKeySigHashes = "sig_hashes"

	AttributeKeyTxHash = "tx_hash"
)

const (
	AttributeValueSeparator = ","
)
