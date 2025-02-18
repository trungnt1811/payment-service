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

// Order set config
const (
	CleanSetInterval    = 5 * time.Second // Interval to clean up the set
	DefaultFillSetLimit = 100             // Default limit to fill the set
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
