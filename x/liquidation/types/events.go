package types

// Auctioin module event types
const (
	EventTypeLiquidate                        = "liquidate"
	EventTypeSignPaymentTransaction           = "sign_payment_transaction"
	EventTypeGenerateSignedPaymentTransaction = "generate_signed_payment_transaction"

	AttributeKeyLiquidator          = "liquidator"
	AttributeKeyLiquidationId       = "liquidation_id"
	AttributeKeyLiquidationRecordId = "liquidation_record_id"
	AttributeKeyDebtAmount          = "debt_amount"
	AttributeKeyCollateralAmount    = "collateral_amount"

	AttributeKeyAgencyPubKey = "agency_pub_key"
	AttributeKeySigHashes    = "sig_hashes"

	AttributeKeyTxHash = "tx_hash"
)

const (
	AttributeValueSeparator = ","
)
