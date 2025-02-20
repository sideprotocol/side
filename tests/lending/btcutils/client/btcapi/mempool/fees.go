package mempool

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Fees defines the fees struct
type Fees struct {
	FastestFee  int64 `json:"fastestFee"`
	HalfHourFee int64 `json:"halfHourFee"`
	HourFee     int64 `json:"hourFee"`
	EconomyFee  int64 `json:"economyFee"`
	MinimumFee  int64 `json:"minimumFee"`
}

// GetFees gets the recommended fees
func (c *Client) GetFees() (*Fees, error) {
	statusCode, resp, err := c.BaseClient.Request(http.MethodGet, fmt.Sprintf("%s/v1/fees/recommended", c.MempoolAPI), c.BaseClient.GetBaseOptions())
	if err != nil {
		return nil, fmt.Errorf("failed to query fees, err: %v", err)
	}

	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to query fees, status code: %d, response: %s", statusCode, string(resp))
	}

	var fees Fees
	err = json.Unmarshal(resp, &fees)
	if err != nil {
		return nil, fmt.Errorf("failed to query fees: invalid response, err: %v", err)
	}

	return &fees, nil
}
