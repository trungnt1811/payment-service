package dto

type TokenTransferHistoryDTOResponse struct {
	NextPage int                       `json:"next_page"`
	Page     int                       `json:"page"`
	Size     int                       `json:"size"`
	Total    int64                     `json:"total,omitempty"`
	Data     []TokenTransferHistoryDTO `json:"data"`
}

type TokenTransferResultDTOResponse struct {
	RequestID    string `json:"request_id"`
	Status       bool   `json:"status"`
	ErrorMessage string `json:"error_message"`
}
