package types

// Auctioin module event types
const (
	EventTypeBid                    = "bid"
	EventTypeSignPaymentTransaction = "sign_payment_transaction"

	AttributeKeyBidId     = "bid_id"
	AttributeKeyBidder    = "bidder"
	AttributeKeyAuctionId = "auction_id"
	AttributeKeyBidPrice  = "bid_price"
	AttributeKeyBidAmount = "bid_amount"

	AttributeKeyAgencyPubKey = "agency_pub_key"
	AttributeKeySigHashes    = "sig_hashes"
)

const (
	AttributeValueSeparator = ","
)
