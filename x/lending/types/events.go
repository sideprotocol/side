package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Lending module event types
const (
	EventTypeRepay     = "repay"
	EventTypeDefault   = "default"
	EventTypeLiquidate = "liquidate"

	AttributeKeyLoanId = "loan_id"

	AttributeKeyBorrower     = "borrower"
	AttributeKeyAdaptorPoint = "adaptor_point"

	AttributeKeyAgencyPubKey = "agency_pub_key"

	AttributeKeySigHashes = "sig_hashes"
)

// GetSigHashesAttributes gets the attribute list for the given sig hashes
func GetSigHashesAttributes(sigHashes []string) []sdk.Attribute {
	attributes := []sdk.Attribute{}

	for _, sigHash := range sigHashes {
		attributes = append(attributes, sdk.NewAttribute(AttributeKeySigHashes, sigHash))
	}

	return attributes
}
