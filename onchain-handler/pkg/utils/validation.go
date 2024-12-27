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
