package constants

import (
	"time"
)

// Global cache key
const (
	TokenDecimals = "token_decimals_"
)

// Event listener config
const (
	DefaultEventChannelBufferSize = 1000 // Buffer size for event channel
	DefaultBlockOffset            = 10   // Default block offset if last processed block is missing
	APIMaxBlocksPerRequest        = 2048 // Maximum number of blocks to query at once
)

// Retry config
const (
	MaxRetries = 3               // Maximum number of retries
	RetryDelay = 3 * time.Second // Delay between retries
)

// Order set config
const (
	CleanSetInterval = 5 * time.Second // Interval to clean up the set
)

// Cache config
const (
	DefaultExpiration = 30 * time.Second
	CleanupInterval   = 1 * time.Minute
)

// Worker config
const (
	LatestBlockFetchInterval    = 5 * time.Second
	ExpiredOrderCatchupInterval = 1 * time.Minute
	OrderCleanInterval          = 5 * time.Second
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

// Eth client cooldown
const (
	EthClientCooldown = 15 * time.Second
)

// Network delay
const (
	DefaultNetworkDelay = 10 * time.Second
)
