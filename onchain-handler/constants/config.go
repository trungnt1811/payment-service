package constants

import (
	"time"
)

// Global cache key
const (
	LatestBlockCacheKey  = "latest_block_"
	LatestBlockCacheTime = 1 * time.Minute

	LastProcessedBlockCacheKey  = "last_processed_block_"
	LastProcessedBlockCacheTime = 1 * time.Minute

	TokenDecimals = "token_decimals_"
)

// Event listener config
const (
	DefaultEventChannelBufferSize = 100  // Buffer size for event channel
	DefaultBlockOffset            = 10   // Default block offset if last processed block is missing
	ApiMaxBlocksPerRequest        = 2048 // Maximum number of blocks to query at once
)

// Retry config
const (
	MaxRetries = 3               // Maximum number of retries when polling fails
	RetryDelay = 3 * time.Second // Delay between retries
)

// Queue config
const (
	DequeueInterval = 5 * time.Second
	MaxQueueLimit   = 10000 // Upper bound for queue size
	MinQueueLimit   = 100   // Minimum size to avoid shrinking too much
	ShrinkThreshold = 0.5   // Shrink when less than 50% of the queue is used
	ScaleFactor     = 1.5   // Factor to scale up the queue when needed
	ShrinkFactor    = 0.75  // Factor to scale down the queue when shrinking
)

// Cache config
const (
	DefaultExpiration = 30 * time.Second
	CleanupInterval   = 1 * time.Minute
)

// Worker config
const (
	LatestBlockFetchInterval          = 3 * time.Second
	ExpiredOrderCatchupInterval       = 1 * time.Minute
	OrderCleanInterval                = 1 * time.Minute
	PaymentWalletBalanceFetchInterval = 1 * time.Minute
)

// Batch constants
const (
	BatchSize  = 250
	BatchDelay = 250 * time.Millisecond
)

// Webhook constants
const (
	MaxWebhookWorkers = 10
	WebhookTimeout    = 5 * time.Second
)

// Withdraw interval
const (
	WithdrawIntervalDaily  = "daily"
	WithdrawIntervalHourly = "hourly"
)
