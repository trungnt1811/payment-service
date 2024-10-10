package dto

import "github.com/genefriendway/onchain-handler/conf"

type TokenTransferFilterDTO struct {
	PoolName        *string `json:"pool_name,omitempty"`
	TransactionHash *string `json:"transaction_hash,omitempty"`
	ToAddress       *string `json:"to_address,omitempty"`
	Symbol          *string `json:"symbol,omitempty"`
}

// Convert filters DTO to a map[string]interface{}
func (d *TokenTransferFilterDTO) ToMap() map[string]interface{} {
	filterMap := make(map[string]interface{})

	if d.TransactionHash != nil && *d.TransactionHash != "" {
		filterMap["transaction_hash"] = *d.TransactionHash
	}
	if d.PoolName != nil && *d.PoolName != "" {
		fromAddress, err := conf.GetPoolAddress(*d.PoolName)
		if err == nil {
			filterMap["from_address"] = fromAddress
		}
	}
	if d.ToAddress != nil && *d.ToAddress != "" {
		filterMap["to_address"] = *d.ToAddress
	}
	if d.Symbol != nil && *d.Symbol != "" {
		filterMap["symbol"] = *d.Symbol
	}

	return filterMap
}
