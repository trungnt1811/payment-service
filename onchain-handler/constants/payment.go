package constants

// Payment orders status
const (
	Pending    = "PENDING"
	Processing = "PROCESSING"
	Success    = "SUCCESS"
	Partial    = "PARTIAL"
	Expired    = "EXPIRED"
	Failed     = "FAILED"
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

// Gas price multiplier
const (
	GasPriceMultiplier = 2.0
)

// Token transfer type
const (
	InternalTransfer = "INTERNAL_TRANSFER"
	Transfer         = "TRANSFER"
	Withdraw         = "WITHDRAW"
	Deposit          = "DEPOSIT"
)
