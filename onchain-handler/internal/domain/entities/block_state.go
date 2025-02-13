package entities

// BlockState represents the state of the last processed block.
type BlockState struct {
	ID                 uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	LatestBlock        uint64 `json:"latest_block"`         // The lastest block from chain
	LastProcessedBlock uint64 `json:"last_processed_block"` // The last processed block
	Network            string `json:"network"`
}

func (m *BlockState) TableName() string {
	return "block_state"
}
