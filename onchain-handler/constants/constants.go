package constants

import (
	"time"
)

// Pagination
const (
	DEFAULT_PAGE_TEXT    = "page"
	DEFAULT_SIZE_TEXT    = "size"
	DEFAULT_PAGE         = "1"
	DEFAULT_PAGE_SIZE    = "10"
	DEFAULT_MIN_PAGESIZE = 5
	DEFAULT_MAX_PAGESIZE = 100
)

// Token decimals
const (
	NativeTokenDecimalsMultiplier = 1e18
	NativeTokenDecimalPlaces      = 18 // for native token like ETH, BNB,... and AVAX
)

// Pool names
const (
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
	LatestBlockCacheTime = 1 * time.Minute

	LastProcessedBlockCacheKey  = "last_processed_block_"
	LastProcessedBlockCacheTime = 1 * time.Minute

	TokenDecimals = "token_decimals_"
)

// Payment orders status
const (
	Pending    = "PENDING"
	Processing = "PROCESSING"
	Success    = "SUCCESS"
	Partial    = "PARTIAL"
	Expired    = "EXPIRED"
	Failed     = "FAILED"
)

// Token transfer type
const (
	InternalTransfer = "INTERNAL_TRANSFER"
	Transfer         = "TRANSFER"
	Withdraw         = "WITHDRAW"
	Deposit          = "DEPOSIT"
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

// Event
const (
	TransferEventName = "Transfer"
)

// Block confirmations
const (
	ConfirmationDepth = 20
)

// Wallet type
type WalletType string

func (t WalletType) String() string {
	return string(t)
}

const (
	PaymentWallet   WalletType = "PaymentWallet"
	UserWallet      WalletType = "UserWallet"
	ReceivingWallet WalletType = "ReceivingWallet"
)

// Method ID
const (
	Erc20BalanceOfMethodID = "70a08231" // ERC20BalanceOfMethodID is the first 4 bytes of the keccak256 hash of "balanceOf(address)"
)

// Network type
type NetworkType string

func (t NetworkType) String() string {
	return string(t)
}

const (
	Bsc        NetworkType = "BSC"
	AvaxCChain NetworkType = "AVAX C-Chain"
)

// SQL constants
const (
	SqlCase = "CASE"
	SqlEnd  = " END"
)

// Order direction
type OrderDirection string

func (t OrderDirection) String() string {
	return string(t)
}

const (
	Asc  OrderDirection = "ASC"
	Desc OrderDirection = "DESC"
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

// Granularity constants
const (
	Daily   = "DAILY"
	Weekly  = "WEEKLY"
	Monthly = "MONTHLY"
	Yearly  = "YEARLY"
)

// Gas price multiplier
const (
	GasPriceMultiplier = 2.0
)

// Withdraw interval
const (
	WithdrawIntervalDaily  = "daily"
	WithdrawIntervalHourly = "hourly"
)

// Eth client cooldown
const (
	EthClientCooldown = 5 * time.Minute
)

// Network delay
const (
	DefaultNetworkDelay = 10 * time.Second
)
