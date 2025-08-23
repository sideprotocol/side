package types

// Compact converts the signing request to the compact version
func (req *SigningRequest) Compact() *CompactSigningRequest {
	return &CompactSigningRequest{
		Address:      req.Address,
		Sequence:     req.Sequence,
		Type:         req.Type,
		Txid:         req.Txid,
		Signers:      GetSigners(req.Psbt),
		SigHashes:    GetSigHashes(req.Psbt),
		CreationTime: req.CreationTime,
		Status:       req.Status,
	}
}

// GlobalRateLimitEnabled returns true if the global rate limit is enabled, false otherwise
func GlobalRateLimitEnabled(rateLimit *RateLimit) bool {
	return rateLimit.GlobalRateLimit.Quota > 0
}

// AddressRateLimitEnabled returns true if the per address rate limit is enabled, false otherwise
func AddressRateLimitEnabled(rateLimit *RateLimit) bool {
	return rateLimit.AddressRateLimit.Quota > 0
}
