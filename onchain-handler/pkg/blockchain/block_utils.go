package blockchain

import (
	"fmt"

	"github.com/genefriendway/onchain-handler/constants"
)

// GetConfirmationDepth returns the confirmation depth for a given network.
func GetConfirmationDepth(network constants.NetworkType) (uint64, error) {
	switch network.String() {
	case constants.Bsc.String():
		return constants.ConfirmationDepthBSC, nil
	case constants.AvaxCChain.String():
		return constants.ConfirmationDepthAVAX, nil
	default:
		return 0, fmt.Errorf("unsupported network: %s", network)
	}
}
