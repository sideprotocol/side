package types

// Lending module event types
const (
	EventTypeRepay     = "repay"
	EventTypeDefault   = "default"
	EventTypeLiquidate = "liquidate"

	AttributeKeyLoanId = "loan_id"

	AttributeKeyBorrower     = "borrower"
	AttributeKeyAdaptorPoint = "adaptor_point"

	AttributeKeyAgencyPubKey = "agency_pub_key"
	AttributeKeySigHashes    = "sig_hashes"
)

const (
	AttributeValueSeparator = ","
)
