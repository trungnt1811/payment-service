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

	"github.com/genefriendway/onchain-handler/conf"
	"github.com/genefriendway/onchain-handler/contracts/abigen/lifepointtoken"
	"github.com/genefriendway/onchain-handler/internal/utils/log"
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

// DistributeTokens distributes tokens from the token distribution address to user wallets using bulk transfer
func DistributeTokens(client *ethclient.Client, config *conf.Configuration, recipients map[string]*big.Int) (*string, error) {
	// Load Blockchain configuration
	chainID := config.Blockchain.ChainID
	privateKey := config.Blockchain.PrivateKeyDistributionAddress
	tokenAddress := config.Blockchain.LifePointAddress

	// Get authentication for signing transactions
	privateKeyECDSA, err := PrivateKeyFromHex(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get private key: %w", err)
	}

	auth, err := GetAuth(client, privateKeyECDSA, new(big.Int).SetUint64(uint64(chainID)))
	if err != nil {
		return nil, fmt.Errorf("failed to get auth: %w", err)
	}

	// Set up the reward token contract instance
	LPToken, err := lifepointtoken.NewLifepointtoken(common.HexToAddress(tokenAddress), client)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate ERC20 contract: %w", err)
	}

	// Prepare recipient addresses and values for bulk transfer
	var recipientAddresses []common.Address
	var tokenAmounts []*big.Int
	for recipientAddress, amount := range recipients {
		recipientAddresses = append(recipientAddresses, common.HexToAddress(recipientAddress))
		tokenAmounts = append(tokenAmounts, amount)
	}

	// Call the bulkTransfer function in the Solidity contract
	tx, err := LPToken.BulkTransfer(auth, recipientAddresses, tokenAmounts)
	if err != nil {
		log.LG.Errorf("Failed to execute bulk transfer: %v", err)
		return nil, err
	}

	// Get the transaction hash after a successful transfer
	txHash := tx.Hash().Hex()

	// Log the transaction hash for tracking
	log.LG.Infof("Bulk transfer executed. Tx hash: %s\n", txHash)

	// Return success with the transaction hash
	return &txHash, nil
}
