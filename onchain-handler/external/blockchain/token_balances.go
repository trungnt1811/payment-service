package blockchain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/external/dto"
	"github.com/genefriendway/onchain-handler/log"
)

const (
	batchSize  = 250                    // Max addresses per batch
	batchDelay = 250 * time.Millisecond // Delay between each batch
)

// GetTokenBalances sends a batch request for ERC-20 balances using `eth_call`, splitting into batches if necessary.
func GetTokenBalances(rpcURL, tokenAddress string, addresses []string) (map[string]*big.Int, error) {
	// Store the result of balances
	balances := make(map[string]*big.Int)

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
			// Prepare data for the `balanceOf` call
			data := "0x" + constants.ERC20BalanceOfMethodID + "000000000000000000000000" + common.HexToAddress(address).Hex()[2:]
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

		// Retry logic for sending the batch request
		var resp *http.Response
		backoff := constants.RetryDelay
		for i := 0; i < constants.MaxRetries; i++ {
			client := &http.Client{Timeout: 10 * time.Second}
			resp, err = client.Post(rpcURL, "application/json", bytes.NewBuffer(reqBody))
			if err == nil && resp.StatusCode == http.StatusOK {
				defer resp.Body.Close()
				break
			}

			log.LG.Infof("Attempt %d failed, retrying in %v\n", i+1, backoff)
			time.Sleep(backoff)
			backoff *= 2
		}
		if err != nil {
			return nil, fmt.Errorf("failed to get balances after %d retries: %v", constants.MaxRetries, err)
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
				log.LG.Infof("Error in response ID %d: %v", response.ID, response.Error.Message)
				continue
			}

			balance := new(big.Int)
			// Ensure the Result is valid before trying to parse
			if response.Result != "" && len(response.Result) > 2 {
				balance.SetString(response.Result[2:], 16)
				balances[addressBatch[response.ID]] = balance
			} else {
				log.LG.Warnf("Empty or invalid result for address %s in batch request", addressBatch[response.ID])
			}
		}

		// Delay between batches to prevent overwhelming the RPC provider
		time.Sleep(batchDelay)
	}

	return balances, nil
}
