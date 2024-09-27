package dto

type LockEventDTOResponse struct {
	NextPage int            `json:"next_page"`
	Page     int            `json:"page"`
	Size     int            `json:"size"`
	Total    int64          `json:"total,omitempty"`
	Data     []LockEventDTO `json:"data"`
}
