package types

// Liquidation module event types
const (
	EventTypeLiquidate                           = "liquidate"
	EventTypeSignSettlementTransaction           = "sign_settlement_transaction"
	EventTypeGenerateSignedSettlementTransaction = "generate_signed_settlement_transaction"

	AttributeKeyLiquidator          = "liquidator"
	AttributeKeyLiquidationId       = "liquidation_id"
	AttributeKeyLiquidationRecordId = "liquidation_record_id"
	AttributeKeyDebtAmount          = "debt_amount"
	AttributeKeyCollateralAmount    = "collateral_amount"
	AttributeKeyBonusAmount         = "bonus_amount"

	AttributeKeyDCMPubKey = "dcm_pub_key"
	AttributeKeySigHashes = "sig_hashes"

	AttributeKeyTxHash = "tx_hash"
)

const (
	AttributeValueSeparator = ","
)
