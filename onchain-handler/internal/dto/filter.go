package dto

type TokenTransferFilterDTO struct {
	FromPoolName    *string `json:"from_pool_name,omitempty"`
	TransactionHash *string `json:"transaction_hash,omitempty"`
	FromAddress     *string `json:"from_address,omitempty"`
	ToAddress       *string `json:"to_address,omitempty"`
	Symbol          *string `json:"symbol,omitempty"`
}

// Convert filters DTO to a map[string]interface{}
func (d *TokenTransferFilterDTO) ToMap() map[string]interface{} {
	filterMap := make(map[string]interface{})

	if d.TransactionHash != nil && *d.TransactionHash != "" {
		filterMap["transaction_hash"] = *d.TransactionHash
	}
	if d.FromPoolName != nil && *d.FromPoolName != "" {
		filterMap["from_pool_name"] = *d.FromPoolName
	}
	if d.FromAddress != nil && *d.FromAddress != "" {
		filterMap["from_address"] = *d.FromAddress
	}
	if d.ToAddress != nil && *d.ToAddress != "" {
		filterMap["to_address"] = *d.ToAddress
	}
	if d.Symbol != nil && *d.Symbol != "" {
		filterMap["symbol"] = *d.Symbol
	}

	return filterMap
}
