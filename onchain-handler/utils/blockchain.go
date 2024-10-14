package utils

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/genefriendway/onchain-handler/blockchain/interfaces"
	"github.com/genefriendway/onchain-handler/conf"
	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/contracts/abigen/bulksender"
	"github.com/genefriendway/onchain-handler/contracts/abigen/lifepointtoken"
	"github.com/genefriendway/onchain-handler/contracts/abigen/usdtmock"
)

// ConnectToNetwork connects to the blockchain network via an RPC URL
func ConnectToNetwork(rpcUrl string) (*ethclient.Client, error) {
	client, err := ethclient.Dial(rpcUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the Ethereum client: %w", err)
	}
	return client, nil
}

// GetFromAddress gets the address associated with a given private key
func GetFromAddress(privateKey *ecdsa.PrivateKey) (string, error) {
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return "", errors.New("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	return fromAddress.Hex(), nil
}

// GetAuth creates a new keyed transactor for signing transactions with the given private key and network chain ID
func GetAuth(
	ctx context.Context,
	client *ethclient.Client,
	privateKey *ecdsa.PrivateKey,
	chainID *big.Int,
) (*bind.TransactOpts, error) {
	fromAddress := crypto.PubkeyToAddress(privateKey.PublicKey)
	nonce, err := client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %w", err)
	}

	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %w", err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to create transactor: %w", err)
	}

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)     // 0 wei, since we're not sending Ether
	auth.GasLimit = uint64(300000) // Set the gas limit (adjust as needed)
	auth.GasPrice = gasPrice       // Set the gas price

	return auth, nil
}

// PrivateKeyFromHex converts a private key string in hex format to an ECDSA private key
func PrivateKeyFromHex(privateKeyHex string) (*ecdsa.PrivateKey, error) {
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to convert private key from hex: %w", err)
	}
	return privateKey, nil
}

// LoadABI loads and parses the ABI from a JSON file
func LoadABI(path string) (abi.ABI, error) {
	file, err := os.Open(path)
	if err != nil {
		return abi.ABI{}, fmt.Errorf("failed to open ABI file at %s: %w", path, err)
	}
	defer file.Close()

	// Read and parse the ABI JSON file
	var parsedABI abi.ABI
	if err := json.NewDecoder(file).Decode(&parsedABI); err != nil {
		return abi.ABI{}, fmt.Errorf("failed to decode ABI JSON from file %s: %w", path, err)
	}

	return parsedABI, nil
}

// PollForLogsFromBlock polls logs from a specified block number onwards for the given contract.
func PollForLogsFromBlock(
	ctx context.Context,
	client *ethclient.Client, // Ethereum client
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
	logs, err := client.FilterLogs(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve logs from block %d: %w", fromBlock, err)
	}

	return logs, nil
}

// GetLatestBlockNumber retrieves the latest block number from the Ethereum client
func GetLatestBlockNumber(ctx context.Context, client *ethclient.Client) (*big.Int, error) {
	header, err := client.HeaderByNumber(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch the latest block header: %w", err)
	}
	return header.Number, nil
}

// ParseHexToUint64 parses a hex string to uint64.
func ParseHexToUint64(hexStr string) (uint64, error) {
	// Ensure the string is lowercase and remove the "0x" prefix if present.
	hexStr = strings.TrimPrefix(strings.ToLower(hexStr), "0x")

	// If the string is empty or all zeros, return 0.
	if len(hexStr) == 0 || strings.TrimLeft(hexStr, "0") == "" {
		return 0, nil
	}

	// Decode the hex string into bytes.
	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return 0, fmt.Errorf("invalid hex string: %w", err)
	}

	// Convert the bytes to uint64. Only the last 8 bytes are relevant for uint64.
	var value uint64
	for _, b := range bytes {
		value = value<<8 + uint64(b)
	}

	return value, nil
}

// Helper function to get pools' private key by pools' addresses
func getPoolPrivateKey(config *conf.Configuration, poolAddress string) (string, error) {
	switch poolAddress {
	// LP Treasury pool
	case config.Blockchain.LPTreasuryPool.LPTreasuryAddress:
		return config.Blockchain.LPTreasuryPool.PrivateKeyLPTreasury, nil
	// LP Community pool
	case config.Blockchain.LPCommunityPool.LPCommunityAddress:
		return config.Blockchain.LPCommunityPool.LPCommunityAddress, nil
	// LP Revenue pool
	case config.Blockchain.LPRevenuePool.LPRevenueAddress:
		return config.Blockchain.LPRevenuePool.PrivateKeyLPRevenue, nil
	// LP Staking pool
	case config.Blockchain.LPStakingPool.LPStakingAddress:
		return config.Blockchain.LPStakingPool.PrivateKeyLPStaking, nil
	// USDT Treasury pool
	case config.Blockchain.USDTTreasuryPool.USDTTreasuryAddress:
		return config.Blockchain.USDTTreasuryPool.PrivateKeyUSDTTreasury, nil
	default:
		return "", fmt.Errorf("failed to get private key for pool address: %s", poolAddress)
	}
}

// BulkTransfer transfers tokens from the pool address to recipients using bulk transfer
func BulkTransfer(
	ctx context.Context,
	client *ethclient.Client,
	config *conf.Configuration,
	poolAddress, symbol string,
	recipients []string,
	amounts []*big.Int,
) (*string, *string, *big.Float, error) {
	chainID := config.Blockchain.ChainID
	bulkSenderContractAddress := config.Blockchain.SmartContract.BulkSenderContractAddress

	var txHash, tokenSymbol string
	var txFeeInAVAX *big.Float

	// Get token address, pool private key, and symbol based on the symbol
	var erc20Token interface{}
	var err error
	var tokenAddress, poolPrivateKey string
	// Get pool private key
	poolPrivateKey, err = getPoolPrivateKey(config, poolAddress)
	if err != nil {
		return &txHash, &tokenSymbol, txFeeInAVAX, fmt.Errorf("failed to get private key for pool address: %s", poolAddress)
	}
	// Get erc20 token
	if symbol == constants.USDT {
		tokenAddress = config.Blockchain.SmartContract.USDTContractAddress
		erc20Token, err = usdtmock.NewUsdtmock(common.HexToAddress(tokenAddress), client)
	} else {
		tokenAddress = config.Blockchain.SmartContract.LifePointContractAddress
		erc20Token, err = lifepointtoken.NewLifepointtoken(common.HexToAddress(tokenAddress), client)
	}
	if err != nil {
		return &txHash, &tokenSymbol, txFeeInAVAX, fmt.Errorf("failed to instantiate ERC20 contract for %s: %w", symbol, err)
	}

	privateKeyECDSA, err := PrivateKeyFromHex(poolPrivateKey)
	if err != nil {
		return &txHash, &tokenSymbol, txFeeInAVAX, fmt.Errorf("failed to retrieve pool private key: %w", err)
	}

	// Function to handle nonce retrieval and retry logic
	var nonce uint64
	nonce, err = getNonceWithRetry(ctx, client, poolAddress)
	if err != nil {
		return &txHash, &tokenSymbol, txFeeInAVAX, fmt.Errorf("failed to retrieve nonce after retry: %w", err)
	}

	auth, err := GetAuth(ctx, client, privateKeyECDSA, new(big.Int).SetUint64(uint64(chainID)))
	if err != nil {
		return &txHash, &tokenSymbol, txFeeInAVAX, fmt.Errorf("failed to create auth object for pool %s: %w", poolAddress, err)
	}
	auth.Nonce = new(big.Int).SetUint64(nonce) // Set the correct nonce

	// Set up the bulk transfer contract instance
	bulkSender, err := bulksender.NewBulksender(common.HexToAddress(bulkSenderContractAddress), client)
	if err != nil {
		return &txHash, &tokenSymbol, txFeeInAVAX, fmt.Errorf("failed to instantiate bulk sender contract: %w", err)
	}

	// Type assertion to ERC20Token interface
	token, ok := erc20Token.(interfaces.ERC20Token)
	if !ok {
		return &txHash, &tokenSymbol, txFeeInAVAX, fmt.Errorf("erc20Token does not implement ERC20Token interface for %s", symbol)
	}

	// Get the token symbol from the contract
	tokenSymbol, err = token.Symbol(nil)
	if err != nil {
		return &txHash, &tokenSymbol, txFeeInAVAX, fmt.Errorf("failed to retrieve token symbol for %s: %w", symbol, err)
	}

	// Calculate total amount to transfer for approval
	totalAmount := big.NewInt(0)
	for _, amount := range amounts {
		totalAmount.Add(totalAmount, amount)
	}

	// Check pool address balance
	poolBalance, err := token.BalanceOf(nil, common.HexToAddress(poolAddress))
	if err != nil {
		return &txHash, &tokenSymbol, txFeeInAVAX, fmt.Errorf("failed to get pool balance: %w", err)
	}

	// Ensure the pool has enough balance
	if poolBalance.Cmp(totalAmount) < 0 {
		return &txHash, &tokenSymbol, txFeeInAVAX, fmt.Errorf("insufficient pool balance: required %s %s, available %s %s", totalAmount.String(), tokenSymbol, poolBalance.String(), tokenSymbol)
	}

	// Approve the bulk transfer contract to spend tokens on behalf of the pool wallet
	tx, err := token.Approve(auth, common.HexToAddress(bulkSenderContractAddress), totalAmount)
	if err != nil {
		return &txHash, &tokenSymbol, txFeeInAVAX, fmt.Errorf("failed to approve bulk sender contract: %w", err)
	}
	txHash = tx.Hash().Hex() // Get the transaction hash

	// Wait for approval to be mined
	receipt, err := bind.WaitMined(ctx, client, tx)
	if err != nil {
		return &txHash, &tokenSymbol, txFeeInAVAX, fmt.Errorf("failed to wait for approval transaction to be mined: %w", err)
	}
	if receipt.Status != 1 {
		return &txHash, &tokenSymbol, txFeeInAVAX, fmt.Errorf("approval transaction failed for %s", txHash)
	}

	// Increment nonce for the next transaction
	nonce++
	auth.Nonce = new(big.Int).SetUint64(nonce)

	// Call the bulk transfer function on the bulk sender contract
	tx, err = bulkSender.BulkTransfer(auth, convertToCommonAddresses(recipients), amounts, common.HexToAddress(tokenAddress))
	if err != nil {
		return &txHash, &tokenSymbol, txFeeInAVAX, fmt.Errorf("failed to execute bulk transfer: %w", err)
	}
	txHash = tx.Hash().Hex() // Update transaction hash for bulk transfer

	// Wait for the bulk transfer transaction to be mined
	receipt, err = bind.WaitMined(ctx, client, tx)
	if err != nil {
		return &txHash, &tokenSymbol, txFeeInAVAX, fmt.Errorf("failed to wait for bulk transfer transaction to be mined: %w", err)
	}

	// Calculate final transaction fee (gasUsed * gasPrice)
	gasUsed := receipt.GasUsed
	txFee := new(big.Int).Mul(new(big.Int).SetUint64(gasUsed), auth.GasPrice)
	weiInAVAX := big.NewFloat(constants.LifePointDecimals)
	txFeeInAVAX = new(big.Float).Quo(new(big.Float).SetInt(txFee), weiInAVAX)

	// Check the transaction status
	if receipt.Status != 1 {
		return &txHash, &tokenSymbol, txFeeInAVAX, fmt.Errorf("bulk transfer transaction failed: %s", txHash)
	}

	return &txHash, &tokenSymbol, txFeeInAVAX, nil
}

// Helper function to retry nonce retrieval
func getNonceWithRetry(ctx context.Context, client *ethclient.Client, poolAddress string) (uint64, error) {
	var nonce uint64
	var err error
	for retryCount := 0; retryCount < 3; retryCount++ {
		nonce, err = client.PendingNonceAt(ctx, common.HexToAddress(poolAddress))
		if err == nil {
			return nonce, nil
		}
		time.Sleep(2 * time.Second) // Backoff before retrying
	}
	return 0, fmt.Errorf("failed to retrieve nonce after 3 retries: %w", err)
}

// Helper function to convert string addresses to common.Address type
func convertToCommonAddresses(recipients []string) []common.Address {
	var addresses []common.Address
	for _, recipient := range recipients {
		addresses = append(addresses, common.HexToAddress(recipient))
	}
	return addresses
}
