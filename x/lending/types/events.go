package types

// Lending module event types
const (
	EventTypeApply                               = "apply"
	EventTypeApprove                             = "approve"
	EventTypeReject                              = "reject"
	EventTypeRedeem                              = "redeem"
	EventTypeRepay                               = "repay"
	EventTypeDefault                             = "default"
	EventTypeLiquidate                           = "liquidate"
	EventTypeGenerateSignedCet                   = "generate_signed_cet"
	EventTypeGenerateSignedRedemptionTransaction = "generate_signed_redemption_transaction"

	AttributeKeyVault            = "vault"
	AttributeKeyBorrower         = "borrower"
	AttributeKeyDCMPubKey        = "dcm_pub_key"
	AttributeKeyMuturityTime     = "muturity_time"
	AttributeKeyFinalTimeout     = "final_timeout"
	AttributeKeyPoolId           = "pool_id"
	AttributeKeyCollateralAmount = "collateral_amount"
	AttributeKeyBorrowAmount     = "borrow_amount"

	AttributeKeyLoanId = "loan_id"
	AttributeKeyId     = "id"
	AttributeKeyAmount = "amount"

	AttributeKeyAuthorizationId = "authorization_id"
	AttributeKeyReason          = "reason"

	AttributeKeyDepositTxHash = "deposit_tx_hash"

	AttributeKeyTxHash = "tx_hash"

	AttributeKeyCetType = "cet_type"
)

const (
	AttributeValueSeparator = ","
)
