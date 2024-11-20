package blockchain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/pkg/dto"
	"github.com/genefriendway/onchain-handler/pkg/logger"
	"github.com/genefriendway/onchain-handler/pkg/utils"
)

const (
	batchSize  = 250                    // Max addresses per batch
	batchDelay = 250 * time.Millisecond // Delay between each batch
)

// validateURL ensures the URL is valid
func validateURL(rawURL string) error {
	_, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	return nil
}

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
func GetTokenBalances(rpcURL, tokenAddress string, addresses []string) (map[string]string, error) {
	// Validate the URL
	if err := validateURL(rpcURL); err != nil {
		return nil, err
	}

	// Store the result of balances
	balances := make(map[string]string)

	// Split addresses into batches based on batchSize
	for start := 0; start < len(addresses); start += batchSize {
		end := start + batchSize
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
						"to":   tokenAddress,
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

		// Send batch request with retry logic
		resp, err := sendBatchRequestWithRetry(rpcURL, reqBody)
		if err != nil {
			return nil, err
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
				return nil, fmt.Errorf("error converting balance to token amount: %v", err)
			}
			balances[addressBatch[response.ID]] = balanceInEth
		}

		// Delay between batches to prevent overwhelming the RPC provider
		time.Sleep(batchDelay)
	}

	return balances, nil
}

// GetNativeTokenBalances retrieves the native token balances (e.g., AVAX, BNB) for a list of addresses.
func GetNativeTokenBalances(rpcURL string, addresses []string) (map[string]string, error) {
	// Validate the URL
	if err := validateURL(rpcURL); err != nil {
		return nil, err
	}

	// Store the result of balances
	balances := make(map[string]string)

	// Split addresses into batches based on batchSize
	for start := 0; start < len(addresses); start += batchSize {
		end := start + batchSize
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

		// Send batch request with retry logic
		resp, err := sendBatchRequestWithRetry(rpcURL, reqBody)
		if err != nil {
			return nil, err
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
		time.Sleep(batchDelay)
	}

	return balances, nil
}
