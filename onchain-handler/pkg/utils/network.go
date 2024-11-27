package utils

import (
	"fmt"

	"github.com/genefriendway/onchain-handler/constants"
)

func GetNativeTokenSymbol(network constants.NetworkType) (string, error) {
	switch network {
	case constants.AvaxCChain:
		return "AVAX", nil
	case constants.Bsc:
		return "BNB", nil
	default:
		return "", fmt.Errorf("unsupported network type: %s", network)
	}
}
