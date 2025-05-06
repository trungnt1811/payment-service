package utils

import (
	"fmt"

	"github.com/genefriendway/onchain-handler/constants"
)

func ValidateNetworkType(network string) error {
	// Validate network type is either BSC or AVAX C-Chain
	validNetworks := map[constants.NetworkType]bool{
		constants.Bsc:        true,
		constants.AvaxCChain: true,
	}
	networkType := constants.NetworkType(network)
	if !validNetworks[networkType] {
		return fmt.Errorf("invalid network type: %s, must be BSC or AVAX C-Chain", network)
	}

	return nil
}

func ValidateSymbol(symbol string) error {
	// Validate symbol is either USDT or USDC
	validSymbols := map[string]bool{
		constants.USDT: true,
		constants.USDC: true,
	}
	if !validSymbols[symbol] {
		return fmt.Errorf("invalid symbol: %s, must be USDT or USDC", symbol)
	}

	return nil
}
