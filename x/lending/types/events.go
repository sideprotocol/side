package types

// Lending module event types
const (
	EventTypeApply                               = "apply"
	EventTypeAuthorize                           = "authorize"
	EventTypeApprove                             = "approve"
	EventTypeReject                              = "reject"
	EventTypeRedeem                              = "redeem"
	EventTypeRepay                               = "repay"
	EventTypeDefault                             = "default"
	EventTypeLiquidate                           = "liquidate"
	EventTypeGenerateSignedCet                   = "generate_signed_cet"
	EventTypeGenerateSignedRedemptionTransaction = "generate_signed_redemption_transaction"
	EventTypeRegisterReferrer                    = "register_referrer"
	EventTypeUpdateReferrer                      = "update_referrer"
	EventTypeReferral                            = "referral"

	AttributeKeyVault              = "vault"
	AttributeKeyBorrower           = "borrower"
	AttributeKeyBorrowerPubKey     = "borrower_pub_key"
	AttributeKeyBorrowerAuthPubKey = "borrower_auth_pub_key"
	AttributeKeyDCMPubKey          = "dcm_pub_key"
	AttributeKeyMaturityTime       = "maturity_time"
	AttributeKeyFinalTimeout       = "final_timeout"
	AttributeKeyPoolId             = "pool_id"
	AttributeKeyBorrowAmount       = "borrow_amount"
	AttributeKeyDLCEventId         = "dlc_event_id"
	AttributeKeyOraclePubKey       = "oracle_pub_key"

	AttributeKeyLoanId = "loan_id"
	AttributeKeyId     = "id"
	AttributeKeyAmount = "amount"

	AttributeKeyCollateralAmount = "collateral_amount"
	AttributeKeyLiquidationPrice = "liquidation_price"

	AttributeKeyAuthorizationId = "authorization_id"
	AttributeKeyReason          = "reason"

	AttributeKeyDepositTxHash = "deposit_tx_hash"

	AttributeKeyTxHash = "tx_hash"

	AttributeKeyCetType = "cet_type"

	AttributeKeyReferrerName      = "referrer_name"
	AttributeKeyReferralCode      = "referral_code"
	AttributeKeyReferrerAddress   = "referrer_address"
	AttributeKeyReferralFeeFactor = "referral_fee_factor"
	AttributeKeyReferralFee       = "referral_fee"
)

const (
	AttributeValueSeparator = ","
)
