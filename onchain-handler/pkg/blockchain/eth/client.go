package eth

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
	"github.com/genefriendway/onchain-handler/contracts/abigen/bulksender"
	"github.com/genefriendway/onchain-handler/contracts/abigen/erc20token"
	pkgcrypto "github.com/genefriendway/onchain-handler/pkg/crypto"
	"github.com/genefriendway/onchain-handler/pkg/interfaces"
	"github.com/genefriendway/onchain-handler/pkg/logger"
	"github.com/genefriendway/onchain-handler/pkg/utils"
)

// roundRobinClient manages a pool of RPC clients for round-robin usage
type roundRobinClient struct {
	clients   []*ethclient.Client
	endpoints []string
	counter   int
	mu        sync.Mutex
}

// NewRoundRobinClient creates a new RoundRobinClient
func NewRoundRobinClient(rpcEndpoints []string) (interfaces.Client, error) {
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
		clients:   clients,
		endpoints: rpcEndpoints,
		counter:   0,
	}, nil
}

// getClient retrieves the next client in the round-robin rotation
func (c *roundRobinClient) getClient() *ethclient.Client {
	c.mu.Lock()
	defer c.mu.Unlock()
	client := c.clients[c.counter]
	c.counter = (c.counter + 1) % len(c.clients)
	return client
}

// executeWithRetry executes a function and retries with the next client on failure
func (r *roundRobinClient) executeWithRetry(
	fn func(client *ethclient.Client) (interface{}, error),
) (interface{}, error) {
	var lastErr error
	baseDelay := constants.RetryDelay

	for i := 0; i < len(r.clients); i++ {
		client := r.getClient()
		logger.GetLogger().Debugf("Using RPC endpoint: %s", r.endpoints[r.counter])

		// Execute the function with the current client
		result, err := fn(client)
		if err == nil {
			// Success
			return result, nil
		}

		// Log failure
		logger.GetLogger().Warnf("RPC call failed on endpoint %s: %v", r.endpoints[r.counter], err)
		lastErr = err

		// Add exponential backoff
		delay := baseDelay * (1 << i) // Exponential backoff: 200ms, 400ms, 800ms, etc.
		logger.GetLogger().Debugf("Delaying next retry by %v", delay)
		time.Sleep(delay)
	}

	// Return the last encountered error if all retries fail
	return nil, fmt.Errorf("all RPC clients failed: %w", lastErr)
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

	result, err := c.executeWithRetry(func(client *ethclient.Client) (interface{}, error) {
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
	result, err := c.executeWithRetry(func(client *ethclient.Client) (interface{}, error) {
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

// BulkTransfer transfers tokens from the pool address to recipients using bulk transfer with round-robin retry logic.
func (c *roundRobinClient) BulkTransfer(
	ctx context.Context,
	chainID uint64,
	bulkSenderContractAddress, poolAddress, poolPrivateKey, tokenContractAddress string,
	recipients []string,
	amounts []*big.Int,
) (*string, *string, *big.Float, error) {
	var txHash, tokenSymbol string
	var txFeeInEth *big.Float

	// Use executeWithRetry for the bulk transfer operation
	_, err := c.executeWithRetry(func(client *ethclient.Client) (interface{}, error) {
		// Step 1: Instantiate ERC20 Token Contract
		erc20Token, err := erc20token.NewErc20token(common.HexToAddress(tokenContractAddress), client)
		if err != nil {
			return nil, fmt.Errorf("failed to instantiate ERC20 contract for token contract address %s: %w", tokenContractAddress, err)
		}

		// Step 2: Parse Private Key
		privateKeyECDSA, err := pkgcrypto.PrivateKeyFromHex(poolPrivateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve pool private key: %w", err)
		}

		// Step 3: Get Nonce
		nonce, err := c.getNonceWithRetry(ctx, poolAddress)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve nonce: %w", err)
		}

		// Step 4: Create Auth Object
		auth, err := c.getAuth(ctx, privateKeyECDSA, new(big.Int).SetUint64(chainID))
		if err != nil {
			return nil, fmt.Errorf("failed to create auth object for pool %s: %w", poolAddress, err)
		}
		auth.Nonce = new(big.Int).SetUint64(nonce) // Set the correct nonce

		// Step 5: Instantiate BulkSender Contract
		bulkSender, err := bulksender.NewBulksender(common.HexToAddress(bulkSenderContractAddress), client)
		if err != nil {
			return nil, fmt.Errorf("failed to instantiate bulk sender contract: %w", err)
		}

		// Step 6: Get Token Symbol
		tokenSymbol, err = erc20Token.Symbol(nil)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve token symbol for token contract address %s: %w", tokenContractAddress, err)
		}

		// Step 7: Calculate Total Amount
		totalAmount := big.NewInt(0)
		for _, amount := range amounts {
			totalAmount.Add(totalAmount, amount)
		}

		// Step 8: Check Pool Balance
		poolBalance, err := erc20Token.BalanceOf(nil, common.HexToAddress(poolAddress))
		if err != nil {
			return nil, fmt.Errorf("failed to get pool balance: %w", err)
		}

		if poolBalance.Cmp(totalAmount) < 0 {
			return nil, fmt.Errorf("insufficient pool balance: required %s %s, available %s %s", totalAmount.String(), tokenSymbol, poolBalance.String(), tokenSymbol)
		}

		// Step 9: Approve Bulk Transfer Contract
		tx, err := erc20Token.Approve(auth, common.HexToAddress(bulkSenderContractAddress), totalAmount)
		if err != nil {
			return nil, fmt.Errorf("failed to approve bulk sender contract: %w", err)
		}
		txHash = tx.Hash().Hex() // Get the transaction hash

		// Wait for Approval Transaction to Be Mined
		receipt, err := bind.WaitMined(ctx, client, tx)
		if err != nil || receipt.Status != 1 {
			return nil, fmt.Errorf("approval transaction failed: %s", txHash)
		}

		// Increment Nonce for the Bulk Transfer
		nonce++
		auth.Nonce = new(big.Int).SetUint64(nonce)

		// Step 10: Execute Bulk Transfer
		tx, err = bulkSender.BulkTransfer(auth, utils.ConvertToCommonAddresses(recipients), amounts, common.HexToAddress(tokenContractAddress))
		if err != nil {
			return nil, fmt.Errorf("failed to execute bulk transfer: %w", err)
		}
		txHash = tx.Hash().Hex() // Update transaction hash for bulk transfer

		// Wait for the Bulk Transfer Transaction to Be Mined
		receipt, err = bind.WaitMined(ctx, client, tx)
		if err != nil || receipt.Status != 1 {
			return nil, fmt.Errorf("bulk transfer transaction failed: %s", txHash)
		}

		// Calculate Transaction Fee
		gasUsed := receipt.GasUsed
		gasPrice := auth.GasPrice
		txFee := new(big.Int).Mul(new(big.Int).SetUint64(gasUsed), gasPrice)
		weiInEth := big.NewFloat(constants.NativeTokenDecimalPlaces)
		txFeeInEth = new(big.Float).Quo(new(big.Float).SetInt(txFee), weiInEth)

		logger.GetLogger().Infof("Bulk transfer successful on endpoint: %s", c.endpoints[c.counter])
		return nil, nil
	})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to execute bulk transfer: %w", err)
	}

	return &txHash, &tokenSymbol, txFeeInEth, nil
}

// GetTokenDecimals retrieves the decimal precision of an ERC20 token by its contract address using round-robin
func (c *roundRobinClient) GetTokenDecimals(ctx context.Context, tokenContractAddress string) (uint8, error) {
	// Use executeWithRetry to perform the operation
	result, err := c.executeWithRetry(func(client *ethclient.Client) (interface{}, error) {
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

// getAuth creates a new keyed transactor for signing transactions with the given private key and network chain ID using round-robin
func (c *roundRobinClient) getAuth(
	ctx context.Context,
	privateKey *ecdsa.PrivateKey,
	chainID *big.Int,
) (*bind.TransactOpts, error) {
	if privateKey == nil || chainID == nil {
		return nil, fmt.Errorf("invalid parameters: privateKey and chainID must not be nil")
	}

	result, err := c.executeWithRetry(func(client *ethclient.Client) (interface{}, error) {
		fromAddress := crypto.PubkeyToAddress(privateKey.PublicKey)

		// Get nonce
		nonce, err := client.PendingNonceAt(ctx, fromAddress)
		if err != nil {
			return nil, fmt.Errorf("failed to get nonce: %w", err)
		}

		// Get gas price
		gasPrice, err := client.SuggestGasPrice(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get gas price: %w", err)
		}

		// Create transactor
		auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
		if err != nil {
			return nil, fmt.Errorf("failed to create transactor: %w", err)
		}

		// Set transaction options
		auth.Nonce = big.NewInt(int64(nonce))
		auth.Value = big.NewInt(0) // 0 wei, since we're not sending Ether
		auth.GasPrice = gasPrice

		return auth, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get auth after retries: %w", err)
	}

	return result.(*bind.TransactOpts), nil
}

// getNonceWithRetry fetches the nonce for the given pool address using round-robin with retries
func (c *roundRobinClient) getNonceWithRetry(ctx context.Context, poolAddress string) (uint64, error) {
	if poolAddress == "" {
		return 0, fmt.Errorf("invalid poolAddress: address cannot be empty")
	}

	result, err := c.executeWithRetry(func(client *ethclient.Client) (interface{}, error) {
		nonce, err := client.PendingNonceAt(ctx, common.HexToAddress(poolAddress))
		if err != nil {
			return nil, fmt.Errorf("failed to fetch nonce: %w", err)
		}
		return nonce, nil
	})
	if err != nil {
		return 0, fmt.Errorf("failed to fetch nonce after retries: %w", err)
	}

	return result.(uint64), nil
}

// EstimateGasERC20 estimates the gas required for an ERC-20 token transfer using round-robin
func (c *roundRobinClient) EstimateGasERC20(
	tokenAddress common.Address,
	fromAddress common.Address,
	toAddress common.Address,
	amount *big.Int,
) (uint64, error) {
	// Parse the ERC-20 ABI
	parsedABI, err := abi.JSON(strings.NewReader(erc20token.Erc20tokenMetaData.ABI))
	if err != nil {
		return 0, fmt.Errorf("failed to parse ERC-20 ABI: %w", err)
	}

	// Encode the 'transfer' function call with parameters
	data, err := parsedABI.Pack("transfer", toAddress, amount)
	if err != nil {
		return 0, fmt.Errorf("failed to pack transfer data: %w", err)
	}

	// Use executeWithRetry for gas estimation
	result, err := c.executeWithRetry(func(client *ethclient.Client) (interface{}, error) {
		// Simulate the transaction to estimate gas
		gasLimit, err := client.EstimateGas(context.Background(), ethereum.CallMsg{
			From:  fromAddress,
			To:    &tokenAddress,
			Value: big.NewInt(0), // ERC-20 transfers don't send native tokens
			Data:  data,          // Encoded transfer data
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
) (common.Hash, uint64, *big.Int, error) {
	fromPrivateKey, err := crypto.HexToECDSA(fromPrivateKeyHex)
	if err != nil {
		return common.Hash{}, 0, nil, fmt.Errorf("invalid private key: %w", err)
	}

	toAddress := common.HexToAddress(toAddressHex)
	tokenAddress := common.HexToAddress(tokenContractAddress)

	result, err := c.executeWithRetry(func(client *ethclient.Client) (interface{}, error) {
		// Get an authorized transactor
		auth, err := c.getAuth(ctx, fromPrivateKey, new(big.Int).SetUint64(chainID))
		if err != nil {
			return nil, fmt.Errorf("failed to get authorized transactor: %w", err)
		}

		// Load the ERC-20 token contract
		token, err := erc20token.NewErc20token(tokenAddress, client)
		if err != nil {
			return nil, fmt.Errorf("failed to load token contract: %w", err)
		}

		// Send the transfer transaction
		tx, err := token.Transfer(auth, toAddress, amount)
		if err != nil {
			return nil, fmt.Errorf("failed to send transaction: %w", err)
		}

		// Wait for the transaction to be mined
		receipt, err := bind.WaitMined(ctx, client, tx)
		if err != nil {
			return nil, fmt.Errorf("failed to wait for transaction to be mined: %w", err)
		}

		// Calculate transaction fee
		gasPrice := auth.GasPrice
		gasUsed := receipt.GasUsed

		return struct {
			Hash     common.Hash
			GasUsed  uint64
			GasPrice *big.Int
		}{
			Hash:     tx.Hash(),
			GasUsed:  gasUsed,
			GasPrice: gasPrice,
		}, nil
	})
	if err != nil {
		return common.Hash{}, 0, nil, fmt.Errorf("failed to transfer token after retries: %w", err)
	}

	// Extract and return the result
	res := result.(struct {
		Hash     common.Hash
		GasUsed  uint64
		GasPrice *big.Int
	})
	return res.Hash, res.GasUsed, res.GasPrice, nil
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

	toAddress := common.HexToAddress(toAddressHex)

	// Use `executeWithRetry` to simplify retry logic
	result, err := c.executeWithRetry(func(client *ethclient.Client) (interface{}, error) {
		// Get an authorized transactor
		auth, err := c.getAuth(ctx, fromPrivateKey, new(big.Int).SetUint64(chainID))
		if err != nil {
			return nil, fmt.Errorf("failed to get authorized transactor: %w", err)
		}

		// Estimate gas for the transaction
		estimatedGas, err := client.EstimateGas(ctx, ethereum.CallMsg{
			To:    &toAddress,
			From:  auth.From,
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
	result, err := c.executeWithRetry(func(client *ethclient.Client) (interface{}, error) {
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

// GetTokenBalance retrieves the balance of a specific ERC20 token for a given wallet address using round-robin retry logic.
func (c *roundRobinClient) GetTokenBalance(
	ctx context.Context,
	tokenContractAddress string,
	walletAddress string,
) (*big.Int, error) {
	// Use executeWithRetry to perform the operation
	result, err := c.executeWithRetry(func(client *ethclient.Client) (interface{}, error) {
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

// Close closes all underlying clients
func (c *roundRobinClient) Close() {
	for _, client := range c.clients {
		client.Close()
	}
}
