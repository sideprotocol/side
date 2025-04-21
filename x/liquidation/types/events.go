package types

// Liquidation module event types
const (
	EventTypeLiquidate                           = "liquidate"
	EventTypeGenerateSignedSettlementTransaction = "generate_signed_settlement_transaction"

	AttributeKeyLiquidator          = "liquidator"
	AttributeKeyLiquidationId       = "liquidation_id"
	AttributeKeyLiquidationRecordId = "liquidation_record_id"
	AttributeKeyDebtAmount          = "debt_amount"
	AttributeKeyCollateralAmount    = "collateral_amount"
	AttributeKeyBonusAmount         = "bonus_amount"

	AttributeKeyTxHash = "tx_hash"
)

const (
	AttributeValueSeparator = ","
)
