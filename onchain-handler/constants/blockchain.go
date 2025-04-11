package constants

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ERC-20 transfer event ABI
const (
	Erc20TransferEventABI = `[{"anonymous":false,"inputs":[{"indexed":true,"name":"from","type":"address"},{"indexed":true,"name":"to","type":"address"},{"indexed":false,"name":"value","type":"uint256"}],"name":"Transfer","type":"event"}]`
)

// Event
const (
	TransferEventName = "Transfer"
)

// Block confirmations
const (
	DefaultConfirmationDepth = 15
	ConfirmationDepthBSC     = 20
	ConfirmationDepthAVAX    = 15
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

// ValidNetworks contains the allowed values for NetworkType
var ValidNetworks = map[NetworkType]bool{
	Bsc:        true,
	AvaxCChain: true,
}

// UnmarshalJSON ensures only valid network types are parsed from JSON.
func (t *NetworkType) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	// Normalize input
	normalized := NetworkType(strings.TrimSpace(str))

	// Check if it's a valid network
	if !ValidNetworks[normalized] {
		supportedNetworks := make([]string, 0, len(ValidNetworks))
		for key := range ValidNetworks {
			supportedNetworks = append(supportedNetworks, string(key))
		}
		return fmt.Errorf("unsupported network type: %s. Supported networks: %v", str, supportedNetworks)
	}

	*t = normalized
	return nil
}

// Token decimals
const (
	NativeTokenDecimalsMultiplier = 1e18
	NativeTokenDecimalPlaces      = 18 // for native token like ETH, BNB,... and AVAX
)

// Token symbols
const (
	USDT = "USDT"
)
