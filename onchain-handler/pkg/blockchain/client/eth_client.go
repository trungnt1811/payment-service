package client

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/contracts/abigen/erc20token"
	clienttypes "github.com/genefriendway/onchain-handler/pkg/blockchain/client/types"
	"github.com/genefriendway/onchain-handler/pkg/logger"
	"github.com/genefriendway/onchain-handler/pkg/utils"
)

// roundRobinClient manages a pool of RPC clients for round-robin usage
type roundRobinClient struct {
	clients        []*ethclient.Client
	endpoints      []string
	counter        int
	mu             sync.Mutex
	failureTracker map[int]time.Time // Tracks failed clients and their cooldown periods
	cooldown       time.Duration     // Cooldown period for retrying a failed client
}

// NewRoundRobinClient creates a new RoundRobinClient
func NewRoundRobinClient(rpcEndpoints []string) (clienttypes.Client, error) {
	if len(rpcEndpoints) == 0 {
		return nil, fmt.Errorf("no RPC endpoints provided")
	}

	clients := make([]*ethclient.Client, len(rpcEndpoints))
	for i, endpoint := range rpcEndpoints {
		client, err := ethclient.Dial(endpoint)
		if err != nil {
			return nil, err
		}
		clients[i] = client
	}

	return &roundRobinClient{
		clients:        clients,
		endpoints:      rpcEndpoints,
		counter:        0,
		failureTracker: make(map[int]time.Time),
		cooldown:       constants.EthClientCooldown,
	}, nil
}

// getClient retrieves the next client in the round-robin rotation
func (c *roundRobinClient) getClient() *ethclient.Client {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Track the index of the last client
	lastClientIndex := len(c.clients) - 1

	for {
		clientIndex := c.counter
		c.counter = (c.counter + 1) % len(c.clients)

		// Check if the client is in cooldown
		if cooldownEnd, failed := c.failureTracker[clientIndex]; failed {
			if time.Now().Before(cooldownEnd) && clientIndex != lastClientIndex {
				logger.GetLogger().Warnf("Skipping client %s due to recent failure (cooldown active)", c.endpoints[clientIndex])
				continue
			}
			// Remove from failure tracker if cooldown has expired
			delete(c.failureTracker, clientIndex)
		}

		// Return the client if not in cooldown or it's the last client
		return c.clients[clientIndex]
	}
}

// getAuth creates a new keyed transactor for signing transactions with the given private key and network chain ID
func (c *roundRobinClient) getAuth(
	ctx context.Context,
	privateKey *ecdsa.PrivateKey,
	chainID *big.Int,
	client *ethclient.Client,
) (*bind.TransactOpts, error) {
	if privateKey == nil || chainID == nil || client == nil {
		return nil, fmt.Errorf("invalid parameters: privateKey, chainID, and client must not be nil")
	}

	fromAddress := crypto.PubkeyToAddress(privateKey.PublicKey)

	// Fetch pending nonce
	pendingNonce, err := client.PendingNonceAt(ctx, fromAddress) // Highest pending nonce
	if err != nil {
		return nil, fmt.Errorf("failed to get pending nonce: %w", err)
	}

	// Get gas price
	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %w", err)
	}

	bufferedGasPrice, err := utils.CalculateBufferedGasPrice(gasPrice, constants.GasPriceMultiplier)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate buffered gas price: %w", err)
	}

	logger.GetLogger().Debugf("Using buffered gas price: %s wei", bufferedGasPrice.String())

	// Create transactor
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to create transactor: %w", err)
	}

	// Set transaction options
	auth.Nonce = big.NewInt(int64(pendingNonce))
	auth.Value = big.NewInt(0) // 0 wei, since we're not sending Ether
	auth.GasPrice = bufferedGasPrice

	logger.GetLogger().Infof("Using nonce %d for address %s", pendingNonce, fromAddress.Hex())
	return auth, nil
}

// detectPendingTxs checks for pending transactions and returns the latest and pending nonces.
func (c *roundRobinClient) detectPendingTxs(
	ctx context.Context,
	client *ethclient.Client,
	fromAddress common.Address,
) (latestNonce, pendingNonce uint64, hasPending bool, err error) {
	latestNonce, err = client.NonceAt(ctx, fromAddress, nil)
	if err != nil {
		return 0, 0, false, fmt.Errorf("failed to get latest nonce: %w", err)
	}

	pendingNonce, err = client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		return 0, 0, false, fmt.Errorf("failed to get pending nonce: %w", err)
	}

	hasPending = pendingNonce > latestNonce
	return latestNonce, pendingNonce, hasPending, nil
}

// executeWithRetry executes a function and retries with the next client on failure
func (c *roundRobinClient) executeWithRetry(
	fn func(client *ethclient.Client) (any, error),
) (any, error) {
	var lastErr error
	baseDelay := constants.RetryDelay

	for range c.clients {
		client := c.getClient()
		clientIndex := (c.counter - 1 + len(c.clients)) % len(c.clients) // Get the last used client index

		logger.GetLogger().Debugf("Using RPC endpoint: %s", c.endpoints[clientIndex])

		// Retry the current client up to 3 times before switching
		for attempt := 1; attempt <= 3; attempt++ {
			result, err := fn(client)
			if err == nil {
				return result, nil // Success
			}

			lastErr = err
			logger.GetLogger().Warnf("RPC call failed on endpoint %s (attempt %d/3): %v", c.endpoints[clientIndex], attempt, err)
			time.Sleep(baseDelay * (1 << (attempt - 1))) // Exponential backoff

			// If this was the last attempt for this client, mark it as failed
			if attempt == 3 {
				c.mu.Lock()
				c.failureTracker[clientIndex] = time.Now().Add(c.cooldown)
				c.mu.Unlock()
				logger.GetLogger().Warnf("Marking RPC endpoint %s as temporarily failed (cooldown active)", c.endpoints[clientIndex])
			}
		}
	}

	// Try the last client as a fallback
	lastClient := c.clients[len(c.clients)-1]
	lastClientIndex := len(c.clients) - 1
	logger.GetLogger().Warnf("Falling back to last RPC endpoint: %s", c.endpoints[lastClientIndex])

	for attempt := 1; attempt <= 3; attempt++ {
		result, err := fn(lastClient)
		if err == nil {
			return result, nil // Success
		}
		lastErr = err
		logger.GetLogger().Warnf("Fallback client failed on attempt %d/3: %v", attempt, err)
		time.Sleep(baseDelay * (1 << (attempt - 1))) // Exponential backoff
	}

	// Return the last encountered error if all retries fail
	return nil, fmt.Errorf("all RPC clients failed after retries, including fallback: %w", lastErr)
}

// PollForLogsFromBlock polls logs from the given block number onwards
func (c *roundRobinClient) PollForLogsFromBlock(
	ctx context.Context,
	contractAddresses []common.Address,
	fromBlock uint64,
	endBlock uint64,
) ([]types.Log, error) {
	query := ethereum.FilterQuery{
		Addresses: contractAddresses,
		FromBlock: big.NewInt(int64(fromBlock)),
		ToBlock:   big.NewInt(int64(endBlock)),
	}

	result, err := c.executeWithRetry(func(client *ethclient.Client) (any, error) {
		logs, err := client.FilterLogs(ctx, query)
		if err != nil {
			return nil, fmt.Errorf("failed to poll logs from block %d to %d: %w", fromBlock, endBlock, err)
		}
		return logs, nil
	})
	if err != nil {
		return nil, err
	}
	return result.([]types.Log), nil
}

// GetLatestBlockNumber retrieves the latest block number using round-robin
func (c *roundRobinClient) GetLatestBlockNumber(ctx context.Context) (*big.Int, error) {
	// Use executeWithRetry to perform the operation
	result, err := c.executeWithRetry(func(client *ethclient.Client) (any, error) {
		header, err := client.HeaderByNumber(ctx, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch the latest block header: %w", err)
		}
		return header.Number, nil
	})
	if err != nil {
		return nil, err
	}

	// Cast the result to the appropriate type
	return result.(*big.Int), nil
}

// GetTokenDecimals retrieves the decimal precision of an ERC20 token by its contract address using round-robin
func (c *roundRobinClient) GetTokenDecimals(ctx context.Context, tokenContractAddress string) (uint8, error) {
	// Use executeWithRetry to perform the operation
	result, err := c.executeWithRetry(func(client *ethclient.Client) (any, error) {
		// Instantiate the ERC20 token contract
		token, err := erc20token.NewErc20token(common.HexToAddress(tokenContractAddress), client)
		if err != nil {
			return nil, fmt.Errorf("failed to instantiate ERC20 token contract at address %s: %w", tokenContractAddress, err)
		}

		// Call the Decimals method
		decimals, err := token.Decimals(&bind.CallOpts{Context: ctx})
		if err != nil {
			return nil, fmt.Errorf("failed to get token decimals for contract %s: %w", tokenContractAddress, err)
		}

		return decimals, nil
	})
	if err != nil {
		return 0, err
	}

	// Cast the result to uint8 and return
	return result.(uint8), nil
}

// EstimateGasGeneric estimates the gas required for a contract call with a custom ABI and method
func (c *roundRobinClient) EstimateGasGeneric(
	contractAddress common.Address,
	fromAddress common.Address,
	abiDef string, // Raw ABI string
	method string, // Method name (e.g., "transfer", "transferFrom")
	args ...any, // Variable arguments for the method
) (uint64, error) {
	// Parse the provided ABI
	parsedABI, err := abi.JSON(strings.NewReader(abiDef))
	if err != nil {
		return 0, fmt.Errorf("failed to parse ABI: %w", err)
	}

	// Encode the method call with the provided arguments
	data, err := parsedABI.Pack(method, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to pack method data for %s: %w", method, err)
	}

	// Use executeWithRetry for gas estimation
	result, err := c.executeWithRetry(func(client *ethclient.Client) (any, error) {
		// Simulate the transaction to estimate gas
		gasLimit, err := client.EstimateGas(context.Background(), ethereum.CallMsg{
			From:  fromAddress,
			To:    &contractAddress,
			Value: big.NewInt(0), // No native token transfer by default
			Data:  data,          // Encoded method data
		})
		if err != nil {
			return nil, fmt.Errorf("failed to estimate gas: %w", err)
		}
		return gasLimit, nil
	})
	if err != nil {
		return 0, fmt.Errorf("failed to estimate gas after retries: %w", err)
	}

	return result.(uint64), nil
}

// TransferToken transfers ERC-20 tokens from one address to another with round-robin retry logic.
func (c *roundRobinClient) TransferToken(
	ctx context.Context,
	chainID uint64,
	tokenContractAddress, fromPrivateKeyHex, toAddressHex string,
	amount *big.Int,
) (common.Hash, uint64, *big.Int, uint64, error) {
	// Validate input
	if amount == nil || amount.Cmp(big.NewInt(0)) <= 0 {
		return common.Hash{}, 0, nil, 0, fmt.Errorf("invalid amount: must be greater than 0")
	}

	fromPrivateKey, err := crypto.HexToECDSA(fromPrivateKeyHex)
	if err != nil {
		return common.Hash{}, 0, nil, 0, fmt.Errorf("invalid private key: %w", err)
	}

	fromAddress := crypto.PubkeyToAddress(fromPrivateKey.PublicKey)
	toAddress := common.HexToAddress(toAddressHex)
	tokenAddress := common.HexToAddress(tokenContractAddress)

	// Set a timeout context for the operation
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	result, err := c.executeWithRetry(func(client *ethclient.Client) (any, error) {
		// Get an authorized transactor
		auth, err := c.getAuth(ctx, fromPrivateKey, new(big.Int).SetUint64(chainID), client)
		if err != nil {
			return nil, fmt.Errorf("failed to get authorized transactor: %w", err)
		}

		// Load the ERC-20 token contract
		token, err := erc20token.NewErc20token(tokenAddress, client)
		if err != nil {
			return nil, fmt.Errorf("failed to load token contract: %w", err)
		}

		// Detect pending transactions
		latestNonce, pendingNonce, hasPending, err := c.detectPendingTxs(ctx, client, fromAddress)
		if err != nil {
			return nil, err
		}

		if hasPending {
			logger.GetLogger().Warnf("Pending transactions detected for address %s: Latest=%d, Pending=%d", fromAddress.Hex(), latestNonce, pendingNonce)
			auth.Nonce = big.NewInt(int64(latestNonce))
			gasPrice := auth.GasPrice
			auth.GasPrice = new(big.Int).Mul(gasPrice, big.NewInt(2)) // Increase gas price for replacement
			logger.GetLogger().Infof("Replacing pending transaction with higher gas price: %s", gasPrice.String())
		}

		// Send the transfer transaction
		tx, err := token.Transfer(auth, toAddress, amount)
		if err != nil {
			return nil, fmt.Errorf("failed to send transaction: %w", err)
		}

		// Wait for the transaction to be mined
		receipt, receiptErr := bind.WaitMined(ctx, client, tx)
		if receiptErr != nil {
			logger.GetLogger().Errorf("Failed to wait for transaction %s to be mined: %v", tx.Hash().Hex(), receiptErr)
			// Return the transaction hash even if receipt retrieval fails
			return struct {
				Hash          common.Hash
				GasUsed       uint64
				GasPrice      *big.Int
				ReceiptStatus uint64
			}{
				Hash:          tx.Hash(),
				GasUsed:       0,
				GasPrice:      auth.GasPrice,
				ReceiptStatus: 0, // Unknown status
			}, nil
		}

		// Return the transaction details
		return struct {
			Hash          common.Hash
			GasUsed       uint64
			GasPrice      *big.Int
			ReceiptStatus uint64
		}{
			Hash:          tx.Hash(),
			GasUsed:       receipt.GasUsed,
			GasPrice:      auth.GasPrice,
			ReceiptStatus: receipt.Status,
		}, nil
	})
	// Extract the result safely
	if err != nil {
		logger.GetLogger().Errorf("Token transfer failed: %v", err)
		return common.Hash{}, 0, nil, 0, fmt.Errorf("failed to transfer token after retries: %w", err)
	}

	res := result.(struct {
		Hash          common.Hash
		GasUsed       uint64
		GasPrice      *big.Int
		ReceiptStatus uint64
	})

	// Log the transaction outcome
	logger.GetLogger().Infof(
		"Token transfer executed: txHash=%s, gasUsed=%d, gasPrice=%s, receiptStatus=%d",
		res.Hash.Hex(), res.GasUsed, res.GasPrice.String(), res.ReceiptStatus,
	)

	return res.Hash, res.GasUsed, res.GasPrice, res.ReceiptStatus, nil
}

// TransferNativeToken transfers native tokens from one address to another with round-robin retry logic.
func (c *roundRobinClient) TransferNativeToken(
	ctx context.Context,
	chainID uint64,
	fromPrivateKeyHex, toAddressHex string,
	amount *big.Int,
) (common.Hash, uint64, *big.Int, error) {
	// Parse private key
	fromPrivateKey, err := crypto.HexToECDSA(fromPrivateKeyHex)
	if err != nil {
		return common.Hash{}, 0, nil, fmt.Errorf("invalid private key: %w", err)
	}

	fromAddress := crypto.PubkeyToAddress(fromPrivateKey.PublicKey)
	toAddress := common.HexToAddress(toAddressHex)

	// Use `executeWithRetry` to simplify retry logic
	result, err := c.executeWithRetry(func(client *ethclient.Client) (any, error) {
		// Get the sender's balance
		balance, err := client.BalanceAt(ctx, fromAddress, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch balance for address %s: %w", fromAddress.Hex(), err)
		}

		// Estimate gas for the transaction
		estimatedGas, err := client.EstimateGas(ctx, ethereum.CallMsg{
			To:    &toAddress,
			From:  fromAddress,
			Value: amount,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to estimate gas: %w", err)
		}

		// Get the gas price
		gasPrice, err := client.SuggestGasPrice(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get gas price: %w", err)
		}

		// Calculate the total gas cost
		gasCost := new(big.Int).Mul(new(big.Int).SetUint64(estimatedGas), gasPrice)

		// Check if the balance is sufficient
		totalCost := new(big.Int).Add(amount, gasCost)
		if balance.Cmp(totalCost) < 0 {
			return nil, fmt.Errorf(
				"insufficient balance for address %s: required %s (amount=%s, gasCost=%s), available %s",
				fromAddress.Hex(), totalCost.String(), amount.String(), gasCost.String(), balance.String(),
			)
		}

		// Get an authorized transactor
		auth, err := c.getAuth(ctx, fromPrivateKey, new(big.Int).SetUint64(chainID), client)
		if err != nil {
			return nil, fmt.Errorf("failed to get authorized transactor: %w", err)
		}

		// Create and sign the transaction
		tx := types.NewTransaction(
			auth.Nonce.Uint64(),
			toAddress,
			amount,
			estimatedGas,
			gasPrice,
			nil,
		)
		signedTx, err := types.SignTx(tx, types.LatestSignerForChainID(new(big.Int).SetUint64(chainID)), fromPrivateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to sign transaction: %w", err)
		}

		// Send the transaction
		err = client.SendTransaction(ctx, signedTx)
		if err != nil {
			return nil, fmt.Errorf("failed to send transaction: %w", err)
		}

		// Return the result
		return struct {
			Hash         common.Hash
			EstimatedGas uint64
			GasPrice     *big.Int
		}{
			Hash:         signedTx.Hash(),
			EstimatedGas: estimatedGas,
			GasPrice:     gasPrice,
		}, nil
	})
	if err != nil {
		return common.Hash{}, 0, nil, fmt.Errorf("failed to transfer native token after retries: %w", err)
	}

	// Extract and return the result
	res := result.(struct {
		Hash         common.Hash
		EstimatedGas uint64
		GasPrice     *big.Int
	})
	return res.Hash, res.EstimatedGas, res.GasPrice, nil
}

// SuggestGasPrice retrieves the suggested gas price using round-robin retry logic.
func (c *roundRobinClient) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	// Use `executeWithRetry` to simplify retry logic
	result, err := c.executeWithRetry(func(client *ethclient.Client) (any, error) {
		gasPrice, err := client.SuggestGasPrice(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to suggest gas price: %w", err)
		}
		return gasPrice, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve suggested gas price after retries: %w", err)
	}

	// Cast and return the result
	return result.(*big.Int), nil
}

// GetBaseFee fetches the current base fee for dynamic fee transactions.
func (c *roundRobinClient) GetBaseFee(ctx context.Context) (*big.Int, error) {
	result, err := c.executeWithRetry(func(client *ethclient.Client) (any, error) {
		header, err := client.HeaderByNumber(ctx, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch header for base fee: %w", err)
		}
		return header.BaseFee, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch base fee: %w", err)
	}

	return result.(*big.Int), nil
}

// SuggestGasTipCap retrieves the suggested gas tip cap for dynamic fee transactions using round-robin retry logic.
func (c *roundRobinClient) SuggestGasTipCap(ctx context.Context) (*big.Int, error) {
	// Use executeWithRetry to simplify retry logic
	result, err := c.executeWithRetry(func(client *ethclient.Client) (any, error) {
		// Use the client's SuggestGasTipCap method to fetch the tip cap
		gasTipCap, err := client.SuggestGasTipCap(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to suggest gas tip cap: %w", err)
		}
		return gasTipCap, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve suggested gas tip cap after retries: %w", err)
	}

	// Cast the result to *big.Int and return
	return result.(*big.Int), nil
}

// GetTokenBalance retrieves the balance of a specific ERC20 token for a given wallet address using round-robin retry logic.
func (c *roundRobinClient) GetTokenBalance(
	ctx context.Context,
	tokenContractAddress string,
	walletAddress string,
) (*big.Int, error) {
	// Use executeWithRetry to perform the operation
	result, err := c.executeWithRetry(func(client *ethclient.Client) (any, error) {
		// Convert the token contract address and wallet address to common.Address
		tokenAddress := common.HexToAddress(tokenContractAddress)
		accountAddress := common.HexToAddress(walletAddress)

		// Load the ERC20 token contract
		token, err := erc20token.NewErc20token(tokenAddress, client)
		if err != nil {
			return nil, fmt.Errorf("failed to load ERC20 token contract: %w", err)
		}

		// Retrieve the token balance
		balance, err := token.BalanceOf(&bind.CallOpts{Context: ctx}, accountAddress)
		if err != nil {
			return nil, fmt.Errorf("failed to get token balance for wallet %s: %w", walletAddress, err)
		}

		return balance, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve token balance after retries: %w", err)
	}

	// Cast the result to *big.Int and return
	return result.(*big.Int), nil
}

// GetNativeTokenBalance retrieves the native token balance (e.g., ETH, BNB) of a wallet address using round-robin retry logic.
func (c *roundRobinClient) GetNativeTokenBalance(
	ctx context.Context,
	walletAddress string,
) (*big.Int, error) {
	// Use executeWithRetry to perform the operation
	result, err := c.executeWithRetry(func(client *ethclient.Client) (any, error) {
		accountAddress := common.HexToAddress(walletAddress)

		// Retrieve the native token balance
		balance, err := client.BalanceAt(ctx, accountAddress, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to get native token balance for wallet %s: %w", walletAddress, err)
		}

		return balance, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve native token balance after retries: %w", err)
	}

	// Cast the result to *big.Int and return
	return result.(*big.Int), nil
}

// Close closes all underlying clients
func (c *roundRobinClient) Close() {
	for _, client := range c.clients {
		client.Close()
	}
}
