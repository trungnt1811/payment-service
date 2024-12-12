package blockchain

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/infra/caching"
	infrainterfaces "github.com/genefriendway/onchain-handler/infra/interfaces"
	"github.com/genefriendway/onchain-handler/pkg/dto"
	pkginterfaces "github.com/genefriendway/onchain-handler/pkg/interfaces"
	"github.com/genefriendway/onchain-handler/pkg/logger"
	"github.com/genefriendway/onchain-handler/pkg/utils"
)

// sendBatchRequestWithRetry sends a batch request with retry logic
func sendBatchRequestWithRetry(rpcURL string, reqBody []byte) (*http.Response, error) {
	backoff := constants.RetryDelay
	var resp *http.Response
	var err error

	for i := 0; i < constants.MaxRetries; i++ {
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err = client.Post(rpcURL, "application/json", bytes.NewBuffer(reqBody))
		if err == nil && resp.StatusCode == http.StatusOK {
			return resp, nil
		}

		logger.GetLogger().Infof("Attempt %d failed, retrying in %v", i+1, backoff)
		time.Sleep(backoff)
		backoff *= 2
	}

	return nil, fmt.Errorf("failed to get balances after %d retries: %v", constants.MaxRetries, err)
}

// GetTokenBalances sends a batch request for ERC-20 balances using `eth_call`, splitting into batches if necessary.
// It uses a list of RPC URLs for round-robin retry logic.
func GetTokenBalances(rpcURLs []string, tokenContractAddress string, tokenDecimals uint8, addresses []string) (map[string]string, error) {
	// Store the result of balances
	balances := make(map[string]string)

	// Initialize a counter for round-robin
	var rpcCounter int

	// Helper to get the next RPC URL in round-robin
	getNextRPCURL := func() string {
		url := rpcURLs[rpcCounter%len(rpcURLs)]
		rpcCounter++
		return url
	}

	// Split addresses into batches based on batchSize
	for start := 0; start < len(addresses); start += constants.BatchSize {
		end := start + constants.BatchSize
		if end > len(addresses) {
			end = len(addresses)
		}
		addressBatch := addresses[start:end]

		// Create batch requests for ERC-20 `balanceOf` calls
		var requests []dto.JSONRPCRequest
		for i, address := range addressBatch {
			data := "0x" + constants.Erc20BalanceOfMethodID + "000000000000000000000000" + common.HexToAddress(address).Hex()[2:]
			request := dto.JSONRPCRequest{
				Jsonrpc: "2.0",
				Method:  "eth_call",
				Params: []interface{}{
					map[string]interface{}{
						"to":   tokenContractAddress,
						"data": data,
					},
					"latest",
				},
				ID: i,
			}
			requests = append(requests, request)
		}

		// Marshal requests into JSON
		reqBody, err := json.Marshal(requests)
		if err != nil {
			return nil, fmt.Errorf("error marshaling batch request: %v", err)
		}

		// Retry logic for sending batch requests
		var resp *http.Response
		for attempt := 0; attempt < constants.MaxRetries; attempt++ {
			rpcURL := getNextRPCURL()
			logger.GetLogger().Debugf("Sending batch request to RPC endpoint: %s (attempt %d/%d)", rpcURL, attempt+1, constants.MaxRetries)

			resp, err = sendBatchRequestWithRetry(rpcURL, reqBody)
			if err == nil && resp.StatusCode == http.StatusOK {
				break // Success
			}

			logger.GetLogger().Warnf("Failed to send batch request to RPC endpoint: %s (attempt %d/%d). Error: %v", rpcURL, attempt+1, constants.MaxRetries, err)
			if resp != nil {
				resp.Body.Close()
			}
			time.Sleep(constants.RetryDelay)
		}

		if err != nil {
			return nil, fmt.Errorf("failed to send batch request after retries: %v", err)
		}
		defer resp.Body.Close()

		// Decode the JSON-RPC response
		var responses []dto.JSONRPCResponse
		if err := json.NewDecoder(resp.Body).Decode(&responses); err != nil {
			return nil, fmt.Errorf("error decoding response: %v", err)
		}

		// Map balances by address in the batch
		for _, response := range responses {
			if response.Error != nil {
				logger.GetLogger().Errorf("Error in response ID %d: %v", response.ID, response.Error.Message)
				continue
			}
			balance := new(big.Int)
			balance.SetString(response.Result[2:], 16)
			balanceInEth, err := utils.ConvertSmallestUnitToFloatToken(balance.String(), tokenDecimals)
			if err != nil {
				return nil, fmt.Errorf("error converting balance to token amount: %v", err)
			}
			balances[addressBatch[response.ID]] = balanceInEth
		}

		// Delay between batches to prevent overwhelming the RPC provider
		time.Sleep(constants.BatchDelay)
	}

	return balances, nil
}

// GetNativeTokenBalances retrieves the native token balances (e.g., AVAX, BNB) for a list of addresses using round-robin retry logic.
func GetNativeTokenBalances(rpcURLs, addresses []string) (map[string]string, error) {
	// Store the result of balances
	balances := make(map[string]string)

	// Initialize a counter for round-robin
	var rpcCounter int

	// Helper to get the next RPC URL in round-robin
	getNextRPCURL := func() string {
		url := rpcURLs[rpcCounter%len(rpcURLs)]
		rpcCounter++
		return url
	}

	// Split addresses into batches based on batchSize
	for start := 0; start < len(addresses); start += constants.BatchSize {
		end := start + constants.BatchSize
		if end > len(addresses) {
			end = len(addresses)
		}
		addressBatch := addresses[start:end]

		// Create batch requests for native token balances
		var requests []dto.JSONRPCRequest
		for i, address := range addressBatch {
			request := dto.JSONRPCRequest{
				Jsonrpc: "2.0",
				Method:  "eth_getBalance",
				Params: []interface{}{
					common.HexToAddress(address).Hex(),
					"latest",
				},
				ID: i,
			}
			requests = append(requests, request)
		}

		// Marshal requests into JSON
		reqBody, err := json.Marshal(requests)
		if err != nil {
			return nil, fmt.Errorf("error marshaling batch request: %v", err)
		}

		// Retry logic for sending batch requests
		var resp *http.Response
		for attempt := 0; attempt < constants.MaxRetries; attempt++ {
			rpcURL := getNextRPCURL()
			logger.GetLogger().Debugf("Sending batch request to RPC endpoint: %s (attempt %d/%d)", rpcURL, attempt+1, constants.MaxRetries)

			resp, err = sendBatchRequestWithRetry(rpcURL, reqBody)
			if err == nil && resp.StatusCode == http.StatusOK {
				break // Success
			}

			logger.GetLogger().Warnf("Failed to send batch request to RPC endpoint: %s (attempt %d/%d). Error: %v", rpcURL, attempt+1, constants.MaxRetries, err)
			if resp != nil {
				resp.Body.Close()
			}
			time.Sleep(constants.RetryDelay)
		}

		if err != nil {
			return nil, fmt.Errorf("failed to send batch request after retries: %v", err)
		}
		defer resp.Body.Close()

		// Decode the JSON-RPC response
		var responses []dto.JSONRPCResponse
		if err := json.NewDecoder(resp.Body).Decode(&responses); err != nil {
			return nil, fmt.Errorf("error decoding response: %v", err)
		}

		// Map balances by address in the batch
		for _, response := range responses {
			if response.Error != nil {
				logger.GetLogger().Errorf("Error in response ID %d: %v", response.ID, response.Error.Message)
				continue
			}
			balance := new(big.Int)
			balance.SetString(response.Result[2:], 16)
			balanceInEth, err := utils.ConvertSmallestUnitToFloatToken(balance.String(), constants.NativeTokenDecimalPlaces)
			if err != nil {
				return nil, fmt.Errorf("error converting balance to native token amount: %v", err)
			}
			balances[addressBatch[response.ID]] = balanceInEth
		}

		// Delay between batches to prevent overwhelming the RPC provider
		time.Sleep(constants.BatchDelay)
	}

	return balances, nil
}

// FetchTokenDecimals retrieves the token decimals for the given token contract address either from the cache or directly from the blockchain.
// If the cache is empty, it saves the token decimals to the cache after retrieving them from the blockchain.
func FetchTokenDecimals(
	ctx context.Context,
	ethClient pkginterfaces.Client,
	tokenContractAddress, network string,
	cacheRepo infrainterfaces.CacheRepository,
) (uint8, error) {
	cacheKey := &caching.Keyer{Raw: constants.TokenDecimals + network + tokenContractAddress}

	var cachedDecimals uint8
	err := cacheRepo.RetrieveItem(cacheKey, &cachedDecimals)
	if err == nil {
		logger.GetLogger().Debugf(
			"Token decimals retrieved from cache. Token: %s, Network: %s, Decimals: %d",
			tokenContractAddress,
			network,
			cachedDecimals,
		)
		return cachedDecimals, nil
	}

	// If not found in cache, get the token decimals from the blockchain
	decimals, err := ethClient.GetTokenDecimals(ctx, tokenContractAddress)
	if err != nil {
		return 0, fmt.Errorf(
			"failed to get token decimals from blockchain. Token: %s, Network: %s, Error: %w",
			tokenContractAddress,
			network,
			err,
		)
	}

	logger.GetLogger().Debugf(
		"Token decimals retrieved from blockchain. Token: %s, Network: %s, Decimals: %d",
		tokenContractAddress,
		network,
		decimals,
	)

	// Save the retrieved decimals to the cache
	// Duration -1 means the item will never expire
	err = cacheRepo.SaveItem(cacheKey, decimals, -1)
	if err != nil {
		logger.GetLogger().Errorf(
			"Failed to persist token decimals to cache. Token: %s, Network: %s, Error: %v",
			tokenContractAddress,
			network,
			err,
		)
	}

	return decimals, nil
}

func GetTokenDecimalsFromCache(
	tokenContractAddress, network string,
	cacheRepo infrainterfaces.CacheRepository,
) (uint8, error) {
	cacheKey := &caching.Keyer{Raw: constants.TokenDecimals + network + tokenContractAddress}

	var cachedDecimals uint8
	err := cacheRepo.RetrieveItem(cacheKey, &cachedDecimals)
	if err == nil {
		logger.GetLogger().Debugf(
			"Token decimals retrieved from cache. Token: %s, Network: %s, Decimals: %d",
			tokenContractAddress,
			network,
			cachedDecimals,
		)
		return cachedDecimals, nil
	}

	return 0, fmt.Errorf("token decimals not found in cache. Token: %s, Network: %s", tokenContractAddress, network)
}
