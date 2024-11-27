package eth

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"
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
	weiInEth := big.NewFloat(constants.NativeTokenDecimalPlaces)
	txFeeInEth = new(big.Float).Quo(new(big.Float).SetInt(txFee), weiInEth)

	// Check the transaction status
	if receipt.Status != 1 {
		return &txHash, &tokenSymbol, txFeeInEth, fmt.Errorf("bulk transfer transaction failed: %s", txHash)
	}

	return &txHash, &tokenSymbol, txFeeInEth, nil
}

// GetTokenDecimals retrieves the decimal precision of an ERC20 token by its contract address
func (c *client) GetTokenDecimals(ctx context.Context, tokenContractAddress string) (uint8, error) {
	// Instantiate the ERC20 token contract
	token, err := erc20token.NewErc20token(common.HexToAddress(tokenContractAddress), c.client)
	if err != nil {
		return 0, fmt.Errorf("failed to instantiate ERC20 token contract at address %s: %w", tokenContractAddress, err)
	}

	// Call the Decimals method to get the token's decimal precision
	decimals, err := token.Decimals(&bind.CallOpts{Context: ctx})
	if err != nil {
		return 0, fmt.Errorf("failed to get token decimals for contract %s: %w", tokenContractAddress, err)
	}

	return decimals, nil
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
	auth.GasPrice = gasPrice   // Set the gas price

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

func (c *client) EstimateGasERC20(
	tokenAddress common.Address,
	fromAddress common.Address,
	toAddress common.Address,
	amount *big.Int,
) (uint64, error) {
	parsedABI, err := abi.JSON(strings.NewReader(erc20token.Erc20tokenMetaData.ABI))
	if err != nil {
		return 0, fmt.Errorf("failed to parse ERC-20 ABI: %w", err)
	}

	// Encode the 'transfer' function call with parameters
	data, err := parsedABI.Pack("transfer", toAddress, amount)
	if err != nil {
		return 0, fmt.Errorf("failed to pack transfer data: %w", err)
	}

	// Estimate gas by simulating the transaction
	gasLimit, err := c.client.EstimateGas(context.Background(), ethereum.CallMsg{
		From:  fromAddress,
		To:    &tokenAddress,
		Value: big.NewInt(0), // ERC-20 transfers don't send native tokens
		Data:  data,          // Encoded transfer data
	})
	if err != nil {
		return 0, fmt.Errorf("failed to estimate gas: %w", err)
	}

	return gasLimit, nil
}

// TransferToken transfers ERC-20 tokens from one address to another
func (c *client) TransferToken(
	ctx context.Context,
	chainID uint64,
	tokenContractAddress, fromPrivateKeyHex, toAddressHex string,
	amount *big.Int,
) (common.Hash, uint64, *big.Int, error) {
	// Parse private key
	fromPrivateKey, err := crypto.HexToECDSA(fromPrivateKeyHex)
	if err != nil {
		return common.Hash{}, 0, nil, fmt.Errorf("invalid private key: %w", err)
	}

	// Parse addresses
	toAddress := common.HexToAddress(toAddressHex)
	tokenAddress := common.HexToAddress(tokenContractAddress)

	// Get an authorized transactor
	auth, err := c.getAuth(ctx, fromPrivateKey, new(big.Int).SetUint64(chainID))
	if err != nil {
		return common.Hash{}, 0, nil, fmt.Errorf("failed to get authorized transactor: %w", err)
	}

	// Load the ERC-20 token contract
	token, err := erc20token.NewErc20token(tokenAddress, c.client)
	if err != nil {
		return common.Hash{}, 0, nil, fmt.Errorf("failed to load token contract: %w", err)
	}

	// Send the transfer transaction
	tx, err := token.Transfer(auth, toAddress, amount)
	if err != nil {
		return common.Hash{}, 0, nil, fmt.Errorf("failed to send transaction: %w", err)
	}

	// Wait for the transaction to be mined
	receipt, err := bind.WaitMined(ctx, c.client, tx)
	if err != nil {
		return common.Hash{}, 0, nil, fmt.Errorf("failed to wait for transaction to be mined: %w", err)
	}

	// Calculate transaction fee (gas used * gas price)
	gasPrice := auth.GasPrice
	gasUsed := receipt.GasUsed

	// Return transaction hash, gas used, and gas price
	return tx.Hash(), gasUsed, gasPrice, nil
}

// TransferNativeToken transfers native tokens from one address to another
func (c *client) TransferNativeToken(
	ctx context.Context,
	chainID uint64,
	fromPrivateKeyHex, toAddressHex string,
	amount *big.Int,
) (common.Hash, uint64, *big.Int, error) {
	fromPrivateKey, err := crypto.HexToECDSA(fromPrivateKeyHex)
	if err != nil {
		return common.Hash{}, 0, nil, fmt.Errorf("invalid private key: %w", err)
	}
	toAddress := common.HexToAddress(toAddressHex)

	// Get an authorized transactor
	auth, err := c.getAuth(ctx, fromPrivateKey, new(big.Int).SetUint64(chainID))
	if err != nil {
		return common.Hash{}, 0, nil, fmt.Errorf("failed to get authorized transactor: %w", err)
	}

	// Estimate gas for the transaction
	estimatedGas, err := c.client.EstimateGas(ctx, ethereum.CallMsg{
		To:    &toAddress,
		From:  auth.From,
		Value: amount,
	})
	if err != nil {
		return common.Hash{}, 0, nil, fmt.Errorf("failed to estimate gas: %w", err)
	}

	// Get the gas price
	gasPrice, err := c.client.SuggestGasPrice(ctx)
	if err != nil {
		return common.Hash{}, 0, nil, fmt.Errorf("failed to get gas price: %w", err)
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
		return common.Hash{}, 0, nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Send the transaction
	err = c.client.SendTransaction(ctx, signedTx)
	if err != nil {
		return common.Hash{}, 0, nil, fmt.Errorf("failed to send transaction: %w", err)
	}

	return signedTx.Hash(), estimatedGas, gasPrice, nil
}

func (c *client) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	return c.client.SuggestGasPrice(ctx)
}
