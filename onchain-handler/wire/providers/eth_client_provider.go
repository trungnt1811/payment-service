package providers

import (
	"fmt"
	"sync"

	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/pkg/blockchain/client"
	clienttypes "github.com/genefriendway/onchain-handler/pkg/blockchain/client/types"
	"github.com/genefriendway/onchain-handler/pkg/logger"
)

var (
	clientMap = make(map[constants.NetworkType]clienttypes.Client)
	clientMux sync.Mutex
)

// ProvideEthClient provides a singleton instance of an Ethereum client for a specific network.
// It takes the network name and RPC URLs as arguments.
func ProvideEthClient(network constants.NetworkType, rpcUrls []string) (clienttypes.Client, error) {
	clientMux.Lock()
	defer clientMux.Unlock()

	// Check if the client for the network already exists
	if client, exists := clientMap[network]; exists {
		return client, nil
	}

	// Create a new Ethereum client if not already initialized
	client, err := client.NewRoundRobinClient(rpcUrls)
	if err != nil {
		logger.GetLogger().Errorf("Failed to initialize Ethereum client for network %s: %v", network, err)
		return nil, fmt.Errorf("failed to initialize Ethereum client for network %s: %w", network, err)
	}

	// Store the client in the map for future use
	clientMap[network] = client
	return client, nil
}
