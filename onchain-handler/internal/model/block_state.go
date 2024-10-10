package model

// BlockState represents the state of the last processed block.
type BlockState struct {
	ID        uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	LastBlock uint64 `json:"last_block"` // The last processed block
}

func (m *BlockState) TableName() string {
	return "block_state"
}
