package eth

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/contracts/abigen/bulksender"
	"github.com/genefriendway/onchain-handler/contracts/abigen/erc20token"
	"github.com/genefriendway/onchain-handler/pkg/blockchain/utils"
	pkgcrypto "github.com/genefriendway/onchain-handler/pkg/crypto"
	"github.com/genefriendway/onchain-handler/pkg/interfaces"
	"github.com/genefriendway/onchain-handler/pkg/logger"
)

type client struct {
	client *ethclient.Client
}

// NewClient new eth client instance
func NewClient(rpcUrl string) interfaces.Client {
	ethClient, err := ethclient.Dial(rpcUrl)
	if err != nil {
		logger.GetLogger().Fatalf("Failed to connect to ETH client: %v", err)
		return nil
	}
	return &client{client: ethClient}
}

func (c *client) Close() {
	c.client.Close()
}

// PollForLogsFromBlock polls logs from a specified block number onwards for the given contract.
func (c *client) PollForLogsFromBlock(
	ctx context.Context,
	contractAddresses []common.Address, // Contract addresses to filter logs
	fromBlock uint64, // Block number to start querying from
	endBlock uint64,
) ([]types.Log, error) {
	// Prepare filter query for the logs
	query := ethereum.FilterQuery{
		Addresses: contractAddresses,            // Contract addresses to filter logs from
		FromBlock: big.NewInt(int64(fromBlock)), // Start block for querying logs
		ToBlock:   big.NewInt(int64(endBlock)),  // End block for querying logs
	}

	// Poll for logs
	logs, err := c.client.FilterLogs(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve logs from block %d: %w", fromBlock, err)
	}

	return logs, nil
}

// GetLatestBlockNumber retrieves the latest block number from the blockchain
func (c *client) GetLatestBlockNumber(ctx context.Context) (*big.Int, error) {
	header, err := c.client.HeaderByNumber(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch the latest block header: %w", err)
	}
	return header.Number, nil
}

// BulkTransfer transfers tokens from the pool address to recipients using bulk transfer
func (c *client) BulkTransfer(
	ctx context.Context,
	chainID uint64,
	bulkSenderContractAddress, poolAddress, poolPrivateKey, tokenContractAddress string,
	recipients []string,
	amounts []*big.Int,
) (*string, *string, *big.Float, error) {
	var txHash, tokenSymbol string
	var txFeeInEth *big.Float

	// Get token address and symbol based on the symbol
	var erc20Token interface{}
	var err error
	erc20Token, err = erc20token.NewErc20token(common.HexToAddress(tokenContractAddress), c.client)
	if err != nil {
		return &txHash, &tokenSymbol, txFeeInEth, fmt.Errorf("failed to instantiate ERC20 contract for token contract address %s: %w", tokenContractAddress, err)
	}

	privateKeyECDSA, err := pkgcrypto.PrivateKeyFromHex(poolPrivateKey)
	if err != nil {
		return &txHash, &tokenSymbol, txFeeInEth, fmt.Errorf("failed to retrieve pool private key: %w", err)
	}

	// Function to handle nonce retrieval and retry logic
	var nonce uint64
	nonce, err = c.getNonceWithRetry(ctx, poolAddress)
	if err != nil {
		return &txHash, &tokenSymbol, txFeeInEth, fmt.Errorf("failed to retrieve nonce after retry: %w", err)
	}

	auth, err := c.getAuth(ctx, privateKeyECDSA, new(big.Int).SetUint64(chainID))
	if err != nil {
		return &txHash, &tokenSymbol, txFeeInEth, fmt.Errorf("failed to create auth object for pool %s: %w", poolAddress, err)
	}
	auth.Nonce = new(big.Int).SetUint64(nonce) // Set the correct nonce

	// Set up the bulk transfer contract instance
	bulkSender, err := bulksender.NewBulksender(common.HexToAddress(bulkSenderContractAddress), c.client)
	if err != nil {
		return &txHash, &tokenSymbol, txFeeInEth, fmt.Errorf("failed to instantiate bulk sender contract: %w", err)
	}

	// Type assertion to ERC20Token interface
	token, ok := erc20Token.(interfaces.ERC20Token)
	if !ok {
		return &txHash, &tokenSymbol, txFeeInEth, fmt.Errorf("erc20Token does not implement ERC20Token interface for token contract address %s", tokenContractAddress)
	}

	// Get the token symbol from the contract
	tokenSymbol, err = token.Symbol(nil)
	if err != nil {
		return &txHash, &tokenSymbol, txFeeInEth, fmt.Errorf("failed to retrieve token symbol for token contract address %s: %w", tokenContractAddress, err)
	}

	// Calculate total amount to transfer for approval
	totalAmount := big.NewInt(0)
	for _, amount := range amounts {
		totalAmount.Add(totalAmount, amount)
	}

	// Check pool address balance
	poolBalance, err := token.BalanceOf(nil, common.HexToAddress(poolAddress))
	if err != nil {
		return &txHash, &tokenSymbol, txFeeInEth, fmt.Errorf("failed to get pool balance: %w", err)
	}

	// Ensure the pool has enough balance
	if poolBalance.Cmp(totalAmount) < 0 {
		return &txHash, &tokenSymbol, txFeeInEth, fmt.Errorf("insufficient pool balance: required %s %s, available %s %s", totalAmount.String(), tokenSymbol, poolBalance.String(), tokenSymbol)
	}

	// Approve the bulk transfer contract to spend tokens on behalf of the pool wallet
	tx, err := token.Approve(auth, common.HexToAddress(bulkSenderContractAddress), totalAmount)
	if err != nil {
		return &txHash, &tokenSymbol, txFeeInEth, fmt.Errorf("failed to approve bulk sender contract: %w", err)
	}
	txHash = tx.Hash().Hex() // Get the transaction hash

	// Wait for approval to be mined
	receipt, err := bind.WaitMined(ctx, c.client, tx)
	if err != nil {
		return &txHash, &tokenSymbol, txFeeInEth, fmt.Errorf("failed to wait for approval transaction to be mined: %w", err)
	}
	if receipt.Status != 1 {
		return &txHash, &tokenSymbol, txFeeInEth, fmt.Errorf("approval transaction failed for %s", txHash)
	}

	// Increment nonce for the next transaction
	nonce++
	auth.Nonce = new(big.Int).SetUint64(nonce)

	// Call the bulk transfer function on the bulk sender contract
	tx, err = bulkSender.BulkTransfer(auth, utils.ConvertToCommonAddresses(recipients), amounts, common.HexToAddress(tokenContractAddress))
	if err != nil {
		return &txHash, &tokenSymbol, txFeeInEth, fmt.Errorf("failed to execute bulk transfer: %w", err)
	}
	txHash = tx.Hash().Hex() // Update transaction hash for bulk transfer

	// Wait for the bulk transfer transaction to be mined
	receipt, err = bind.WaitMined(ctx, c.client, tx)
	if err != nil {
		return &txHash, &tokenSymbol, txFeeInEth, fmt.Errorf("failed to wait for bulk transfer transaction to be mined: %w", err)
	}

	// Calculate transaction fee (gasUsed * gasPrice)
	gasUsed := receipt.GasUsed
	gasPrice := auth.GasPrice
	txFee := new(big.Int).Mul(new(big.Int).SetUint64(gasUsed), gasPrice)
	// Convert txFee from wei to Eth (1 Eth = 10^18 wei)
	weiInEth := big.NewFloat(constants.TokenDecimalsMultiplier)
	txFeeInEth = new(big.Float).Quo(new(big.Float).SetInt(txFee), weiInEth)

	// Check the transaction status
	if receipt.Status != 1 {
		return &txHash, &tokenSymbol, txFeeInEth, fmt.Errorf("bulk transfer transaction failed: %s", txHash)
	}

	return &txHash, &tokenSymbol, txFeeInEth, nil
}

// getAuth creates a new keyed transactor for signing transactions with the given private key and network chain ID
func (c *client) getAuth(
	ctx context.Context,
	privateKey *ecdsa.PrivateKey,
	chainID *big.Int,
) (*bind.TransactOpts, error) {
	fromAddress := crypto.PubkeyToAddress(privateKey.PublicKey)
	nonce, err := c.client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %w", err)
	}

	gasPrice, err := c.client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %w", err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to create transactor: %w", err)
	}

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0) // 0 wei, since we're not sending Ether
	// auth.GasLimit = uint64(300000) // Set the gas limit (adjust as needed)
	auth.GasPrice = gasPrice // Set the gas price

	return auth, nil
}

// Helper function to retry nonce retrieval
func (c *client) getNonceWithRetry(ctx context.Context, poolAddress string) (uint64, error) {
	var nonce uint64
	var err error
	for retryCount := 0; retryCount < constants.MaxRetries; retryCount++ {
		nonce, err = c.client.PendingNonceAt(ctx, common.HexToAddress(poolAddress))
		if err == nil {
			return nonce, nil
		}
		time.Sleep(constants.RetryDelay) // Backoff before retrying
	}
	return 0, fmt.Errorf("failed to retrieve nonce after %d retries: %w", constants.MaxRetries, err)
}
