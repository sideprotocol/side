package types

// Lending module event types
const (
	EventTypeApply                                 = "apply"
	EventTypeApprove                               = "approve"
	EventTypeReject                                = "reject"
	EventTypeCancel                                = "cancel"
	EventTypeRepay                                 = "repay"
	EventTypeDefault                               = "default"
	EventTypeLiquidate                             = "liquidate"
	EventTypeGenerateSignedCet                     = "generate_signed_cet"
	EventTypeGenerateSignedCancellationTransaction = "generate_signed_cancellation_transaction"

	AttributeKeyVault            = "vault"
	AttributeKeyBorrower         = "borrower"
	AttributeKeyDCMPubKey        = "dcm_pub_key"
	AttributeKeyMuturityTime     = "muturity_time"
	AttributeKeyFinalTimeout     = "final_timeout"
	AttributeKeyPoolId           = "pool_id"
	AttributeKeyCollateralAmount = "collateral_amount"
	AttributeKeyBorrowAmount     = "borrow_amount"

	AttributeKeyLoanId = "loan_id"

	AttributeKeySender = "sender"
	AttributeKeyAmount = "amount"

	AttributeKeyDepositTxHash = "deposit_tx_hash"

	AttributeKeyTxHash = "tx_hash"

	AttributeKeyCetType = "cet_type"
)

const (
	AttributeValueSeparator = ","
)
