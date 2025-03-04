package types

// Lending module event types
const (
	EventTypeApply     = "apply"
	EventTypeRepay     = "repay"
	EventTypeDefault   = "default"
	EventTypeLiquidate = "liquidate"

	AttributeKeyVault            = "vault"
	AttributeKeyBorrower         = "borrower"
	AttributeKeyAgencyPubKey     = "agency_pub_key"
	AttributeKeyLoanSecretHash   = "loan_secret_hash"
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
)

const (
	AttributeValueSeparator = ","
)
