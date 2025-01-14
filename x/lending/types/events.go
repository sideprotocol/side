package types

// Lending module event types
const (
	EventTypeRepay     = "repay"
	EventTypeDefault   = "default"
	EventTypeLiquidate = "liquidate"

	AttributeKeyLoanId = "loan_id"

	AttributeKeyBorrower     = "borrower"
	AttributeKeyAdaptorPoint = "adaptor_point"

	AttributeKeyEventPubKey = "event_pub_key"
	AttributeKeyEventNonce  = "event_nonce"
	AttributeKeyEventPrice  = "event_price"
)
