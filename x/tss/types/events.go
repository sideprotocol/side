package types

import (
	fmt "fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TSS module event types
const (
	EventTypeInitiateDKG     = "initiate_dkg"
	EventTypeCompleteDKG     = "complete_dkg"
	EventTypeInitiateSigning = "initiate_signing"
	EventTypeCompleteSigning = "complete_signing"

	AttributeKeySender = "sender"

	AttributeKeyId       = "id"
	AttributeKeyModule   = "module"
	AttributeKeyScopedId = "scoped_id"
	AttributeKeyType     = "type"
	AttributeKeyIntent   = "intent"

	AttributeKeyParticipants   = "participants"
	AttributeKeyThreshold      = "threshold"
	AttributeKeyBatchSize      = "batch_size"
	AttributeKeyExpirationTime = "expiration_time"

	AttributeKeyParticipant = "participant"

	AttributeKeyPubKey       = "pub_key"
	AttributeKeySigHashes    = "sig_hashes"
	AttributeKeyNonce        = "nonce"
	AttributeKeyAdaptorPoint = "adaptor_point"
)

const (
	AttributeValueSeparator = ","
)

// GetSigningRequestEventAttributes gets the event attributes for the signing request
func GetSigningRequestEventAttributes(id uint64, module string, scopedId string, ty SigningType, intent int32, pubKey string, sigHashes []string, options *SigningOptions) []sdk.Attribute {
	attributes := []sdk.Attribute{}

	attributes = append(attributes, sdk.NewAttribute(AttributeKeyId, fmt.Sprintf("%d", id)))
	attributes = append(attributes, sdk.NewAttribute(AttributeKeyModule, module))
	attributes = append(attributes, sdk.NewAttribute(AttributeKeyScopedId, scopedId))
	attributes = append(attributes, sdk.NewAttribute(AttributeKeyType, fmt.Sprintf("%d", ty)))
	attributes = append(attributes, sdk.NewAttribute(AttributeKeyIntent, fmt.Sprintf("%d", intent)))
	attributes = append(attributes, sdk.NewAttribute(AttributeKeyPubKey, pubKey))
	attributes = append(attributes, sdk.NewAttribute(AttributeKeySigHashes, strings.Join(sigHashes, AttributeValueSeparator)))

	if options != nil {
		attributes = append(attributes, GetSigningOptionAttribute(ty, options))
	}

	return attributes
}

// GetSigningOptionAttribute gets the event attribute for the signing option
// Assume that the options match the signing type
func GetSigningOptionAttribute(signingType SigningType, options *SigningOptions) sdk.Attribute {
	switch signingType {
	case SigningType_SIGNING_TYPE_SCHNORR_WITH_COMMITMENT:
		return sdk.Attribute{
			Key:   AttributeKeyNonce,
			Value: options.Nonce,
		}

	case SigningType_SIGNING_TYPE_SCHNORR_ADAPTOR:
		return sdk.Attribute{
			Key:   AttributeKeyAdaptorPoint,
			Value: options.AdaptorPoint,
		}

	default:
		return sdk.Attribute{}
	}
}
