package constants

import "time"

// Token decimals
const (
	TokenDecimalsMultiplier = 1e18
	DecimalPlaces           = 18
)

// Pool names
const (
	LPCommunity  = "LP_Community"
	LPStaking    = "LP_Staking"
	LPRevenue    = "LP_Revenue"
	LPTreasury   = "LP_Treasury"
	USDTTreasury = "USDT_Treasury"
)

// Token symbols
const (
	USDT = "USDT"
	LP   = "LP"
)

// Global cache key
const (
	LatestBlockCacheKey  = "latest_block_"
	LatestBlockCacheTime = 5 * time.Minute
)

// Payment orders status
const (
	Pending = "PENDING"
	Success = "SUCCESS"
	Partial = "PARTIAL"
	Expired = "EXPIRED"
	Failed  = "FAILED"
)

// ERC-20 transfer event ABI
const (
	Erc20TransferEventABI = `[{"anonymous":false,"inputs":[{"indexed":true,"name":"from","type":"address"},{"indexed":true,"name":"to","type":"address"},{"indexed":false,"name":"value","type":"uint256"}],"name":"Transfer","type":"event"}]`
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
	DequeueInterval = 15 * time.Second
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
	LatestBlockFetchingInterval = 5 * time.Second
	ExpiredOrderCatchupInterval = 5 * time.Minute
	OrderCleanInterval          = 15 * time.Minute
)

// Event
const (
	TransferEventName = "Transfer"
)

// Payment order config
const (
	OrderCutoffTime = 24 * time.Hour
)

// Block confirmations
const (
	ConfirmationDepth = 30
)

// Wallet type
type WalletType string

const (
	PaymentWallet WalletType = "PaymentWallet"
	UserWallet    WalletType = "UserWallet"
)

// Method ID
const (
	Erc20BalanceOfMethodID = "70a08231" // ERC20BalanceOfMethodID is the first 4 bytes of the keccak256 hash of "balanceOf(address)"
)

// Network type
type NetworkType string

const (
	Bsc        NetworkType = "BSC"
	AvaxCChain NetworkType = "AVAX C-Chain"
)
