package dto

type TokenTransferPayloadDTO struct {
	// PoolName should be: LP_Community, LP_Staking, LP_Revenue, LP_Treasury, USDT_Treasury
	PoolName    string `json:"pool_name"`
	ToAddress   string `json:"to_address"`
	TokenAmount uint64 `json:"token_amount"`
	RequestID   string `json:"request_id"`
}
