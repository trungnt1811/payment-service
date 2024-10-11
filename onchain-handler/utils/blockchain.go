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

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/genefriendway/onchain-handler/blockchain/interfaces"
	"github.com/genefriendway/onchain-handler/conf"
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
func GetAuth(client *ethclient.Client, privateKey *ecdsa.PrivateKey, chainID *big.Int) (*bind.TransactOpts, error) {
	fromAddress := crypto.PubkeyToAddress(privateKey.PublicKey)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %w", err)
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
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

// BulkTransfer transfers tokens from the pool address to user wallets using bulk transfer
func BulkTransfer(client *ethclient.Client, config *conf.Configuration, poolAddress string, recipients []string, amounts []*big.Int) (*string, *string, *big.Float, error) {
	chainID := config.Blockchain.ChainID
	bulkSenderContractAddress := config.Blockchain.SmartContract.BulkSenderContractAddress

	// Get the token address, pool private key, and symbol based on the pool address
	tokenAddress, poolPrivateKey, symbol, err := getPoolDetails(poolAddress, config)
	if err != nil {
		return nil, nil, nil, err
	}

	fmt.Println("CHECK")
	fmt.Println(poolAddress)
	fmt.Println(tokenAddress)
	fmt.Println(symbol)

	// Get authentication for signing transactions
	privateKeyECDSA, err := PrivateKeyFromHex(poolPrivateKey)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get pool private key: %w", err)
	}

	// Get the initial nonce
	nonce, err := client.PendingNonceAt(context.Background(), common.HexToAddress(poolAddress))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get nonce: %w", err)
	}

	auth, err := GetAuth(client, privateKeyECDSA, new(big.Int).SetUint64(uint64(chainID)))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get auth: %w", err)
	}
	auth.Nonce = new(big.Int).SetUint64(nonce)

	// Set up the ERC20 token contract instance (LifePoint or USDT, depending on pool)
	erc20Token, err := getERC20TokenInstance(tokenAddress, symbol, client)
	if err != nil {
		return nil, nil, nil, err
	}

	// Set up the bulk transfer contract instance
	bulkSender, err := bulksender.NewBulksender(common.HexToAddress(bulkSenderContractAddress), client)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to instantiate bulk sender contract: %w", err)
	}

	// Calculate total amount to transfer for approval
	totalAmount := big.NewInt(0)
	for _, amount := range amounts {
		totalAmount = new(big.Int).Add(totalAmount, amount)
	}

	// Type assertion to ERC20Token interface
	token, ok := erc20Token.(interfaces.ERC20Token)
	if !ok {
		return nil, nil, nil, fmt.Errorf("erc20Token does not implement ERC20Token interface")
	}

	// Get the token symbol from the contract
	tokenSymbol, err := token.Symbol(nil)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get token symbol from the contract: %w", err)
	}

	// Approve the bulk transfer contract to spend tokens on behalf of the pool wallet
	tx, err := token.Approve(auth, common.HexToAddress(bulkSenderContractAddress), totalAmount)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to approve bulk sender contract: %w", err)
	}

	// Wait for approval to be mined
	receipt, err := bind.WaitMined(context.Background(), client, tx)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to wait for approval transaction to be mined: %w", err)
	}
	if receipt.Status != 1 {
		return nil, nil, nil, fmt.Errorf("approval transaction failed")
	}

	nonce++ // Increment nonce for the next transaction
	auth.Nonce = new(big.Int).SetUint64(nonce)

	// Call the bulk transfer function on the bulk sender contract
	tx, err = bulkSender.BulkTransfer(auth, convertToCommonAddresses(recipients), amounts, common.HexToAddress(tokenAddress))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to execute bulk transfer: %w", err)
	}

	// Wait for the bulk transfer transaction to be mined and get the receipt
	receipt, err = bind.WaitMined(context.Background(), client, tx)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to wait for bulk transfer transaction to be mined: %w", err)
	}

	txHash := tx.Hash().Hex()

	// Check the transaction status
	if receipt.Status != 1 {
		return nil, nil, nil, fmt.Errorf("bulk transfer transaction failed: %s", txHash)
	}

	// Calculate transaction fee (gasUsed * gasPrice)
	gasUsed := receipt.GasUsed
	gasPrice := auth.GasPrice
	txFee := new(big.Int).Mul(new(big.Int).SetUint64(gasUsed), gasPrice)

	// Convert txFee from wei to AVAX (1 AVAX = 10^18 wei)
	weiInAVAX := big.NewFloat(1e18)
	txFeeInAVAX := new(big.Float).Quo(new(big.Float).SetInt(txFee), weiInAVAX)

	// Return transaction hash, token symbol, and transaction fee
	return &txHash, &tokenSymbol, txFeeInAVAX, nil
}

// Helper function to convert string addresses to common.Address type
func convertToCommonAddresses(recipients []string) []common.Address {
	var addresses []common.Address
	for _, recipient := range recipients {
		addresses = append(addresses, common.HexToAddress(recipient))
	}
	return addresses
}

// Helper function to get the ERC20 token instance
func getERC20TokenInstance(tokenAddress, symbol string, client *ethclient.Client) (interface{}, error) {
	var erc20Token interface{}
	var err error

	if symbol == "USDT" {
		erc20Token, err = usdtmock.NewUsdtmock(common.HexToAddress(tokenAddress), client)
	} else {
		erc20Token, err = lifepointtoken.NewLifepointtoken(common.HexToAddress(tokenAddress), client)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to instantiate ERC20 contract: %w", err)
	}

	return erc20Token, nil
}

// Helper function to get pool details based on the pool address
func getPoolDetails(poolAddress string, config *conf.Configuration) (string, string, string, error) {
	switch poolAddress {
	case config.Blockchain.USDTTreasuryPool.USDTTreasuryAddress:
		return config.Blockchain.SmartContract.USDTContractAddress,
			config.Blockchain.USDTTreasuryPool.PrivateKeyUSDTTreasury, "USDT", nil
	case config.Blockchain.LPTreasuryPool.LPTreasuryAddress:
		return config.Blockchain.SmartContract.LifePointContractAddress,
			config.Blockchain.LPTreasuryPool.PrivateKeyLPTreasury, "LP", nil
	case config.Blockchain.LPCommunityPool.LPCommunityAddress:
		return config.Blockchain.SmartContract.LifePointContractAddress,
			config.Blockchain.LPCommunityPool.PrivateKeyLPCommunity, "LP", nil
	case config.Blockchain.LPRevenuePool.LPRevenueAddress:
		return config.Blockchain.SmartContract.LifePointContractAddress,
			config.Blockchain.LPRevenuePool.PrivateKeyLPRevenue, "LP", nil
	case config.Blockchain.LPStakingPool.LPStakingAddress:
		return config.Blockchain.SmartContract.LifePointContractAddress,
			config.Blockchain.LPStakingPool.PrivateKeyLPStaking, "LP", nil
	default:
		return "", "", "", fmt.Errorf("unrecognized pool address: %s", poolAddress)
	}
}
