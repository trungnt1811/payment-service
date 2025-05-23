package workers

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/genefriendway/onchain-handler/conf"
	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/contracts/abigen/erc20token"
	cachetypes "github.com/genefriendway/onchain-handler/internal/adapters/cache/types"
	"github.com/genefriendway/onchain-handler/internal/delivery/dto"
	ucasetypes "github.com/genefriendway/onchain-handler/internal/domain/ucases/types"
	workertypes "github.com/genefriendway/onchain-handler/internal/workers/types"
	"github.com/genefriendway/onchain-handler/pkg/blockchain"
	clienttypes "github.com/genefriendway/onchain-handler/pkg/blockchain/client/types"
	"github.com/genefriendway/onchain-handler/pkg/crypto"
	"github.com/genefriendway/onchain-handler/pkg/logger"
	"github.com/genefriendway/onchain-handler/pkg/payment"
	"github.com/genefriendway/onchain-handler/pkg/utils"
)

type walletInfo struct {
	ID          uint64
	TokenAmount *big.Int
}

type paymentWalletWithdrawWorker struct {
	ctx                    context.Context
	ethClient              clienttypes.Client
	network                constants.NetworkType
	chainID                uint64
	cacheRepo              cachetypes.CacheRepository
	tokenTransferUCase     ucasetypes.TokenTransferUCase
	paymentWalletUCase     ucasetypes.PaymentWalletUCase
	tokenContractAddresses []string
	masterWalletAddress    string
	mnemonic               string
	passphrase             string
	salt                   string
	gasBufferMultiplier    float64
	withdrawInterval       string
	isRunning              bool
	mu                     sync.Mutex
}

func NewPaymentWalletWithdrawWorker(
	ctx context.Context,
	ethClient clienttypes.Client,
	network constants.NetworkType,
	chainID uint64,
	cacheRepo cachetypes.CacheRepository,
	tokenTransferUCase ucasetypes.TokenTransferUCase,
	paymentWalletUCase ucasetypes.PaymentWalletUCase,
	tokenContractAddresses []string,
	masterWalletAddress string,
	mnemonic, passphrase, salt string,
	gasBufferMultiplier float64,
	withdrawInterval string,
) workertypes.Worker {
	return &paymentWalletWithdrawWorker{
		ctx:                    ctx,
		ethClient:              ethClient,
		network:                network,
		chainID:                chainID,
		cacheRepo:              cacheRepo,
		tokenTransferUCase:     tokenTransferUCase,
		paymentWalletUCase:     paymentWalletUCase,
		tokenContractAddresses: tokenContractAddresses,
		masterWalletAddress:    masterWalletAddress,
		mnemonic:               mnemonic,
		passphrase:             passphrase,
		salt:                   salt,
		gasBufferMultiplier:    gasBufferMultiplier,
		withdrawInterval:       withdrawInterval,
	}
}

func (w *paymentWalletWithdrawWorker) Start(ctx context.Context) {
	for {
		// Calculate the duration until the next scheduled time (e.g., midnight)
		var sleepDuration time.Duration

		switch w.withdrawInterval {
		case constants.WithdrawIntervalHourly:
			sleepDuration = time.Hour
		case constants.WithdrawIntervalDaily:
			now := time.Now()
			nextRun := time.Date(
				now.Year(), now.Month(), now.Day()+1, // Next day
				0, 0, 0, 0, // 00:00:00
				now.Location(),
			)
			sleepDuration = time.Until(nextRun)
		default:
			logger.GetLogger().Errorf("Invalid RunInterval: %s. Defaulting to hourly.", w.withdrawInterval)
			sleepDuration = time.Hour
		}

		logger.GetLogger().Infof("Next payment wallet withdrawal on network %s scheduled in: %s", w.network, sleepDuration)

		// Sleep until the next scheduled time or exit early if the context is canceled
		select {
		case <-time.After(sleepDuration):
			go w.run(ctx)
		case <-ctx.Done():
			logger.GetLogger().Infof("Shutting down paymentWalletWithdrawWorker on network %s", w.network)
			return
		}
	}
}

func (w *paymentWalletWithdrawWorker) run(ctx context.Context) {
	w.mu.Lock()
	if w.isRunning {
		logger.GetLogger().Warnf("Previous paymentWalletWithdrawWorker on network %s run still in progress, skipping this cycle", w.network)
		w.mu.Unlock()
		return
	}

	w.isRunning = true
	w.mu.Unlock()

	defer func() {
		w.mu.Lock()
		w.isRunning = false
		w.mu.Unlock()
	}()

	var lastErr error
	for attempt := 1; attempt <= constants.MaxRetries; attempt++ {
		lastErr = w.withdraw(ctx)
		if lastErr == nil {
			logger.GetLogger().Infof("Withdrawal process on network %s succeeded on attempt %d", w.network, attempt)
			return
		}

		logger.GetLogger().Errorf("Withdrawal process on network %s failed on attempt %d: %v", w.network, attempt, lastErr)

		// Check if the context has been canceled to avoid infinite retries when shutting down
		select {
		case <-ctx.Done():
			logger.GetLogger().Infof("Withdrawal process on network %s stopped due to context cancellation", w.network)
			return
		case <-time.After(constants.RetryDelay): // Wait before retrying
		}
	}

	logger.GetLogger().Errorf("Withdrawal process on network %s failed after %d attempts: %v", w.network, constants.MaxRetries, lastErr)
}

func (w *paymentWalletWithdrawWorker) withdraw(ctx context.Context) error {
	// Step 1: Get native token symbol
	nativeTokenSymbol, err := blockchain.GetNativeTokenSymbol(w.network)
	if err != nil {
		return fmt.Errorf("failed to get native token symbol on network %s: %w", w.network, err)
	}

	// Step 2: Fetch payment wallets with balances
	wallets, err := w.paymentWalletUCase.GetPaymentWalletsWithBalances(
		ctx, &w.network, []string{constants.USDC, constants.USDT},
	)
	if err != nil {
		return fmt.Errorf("failed to get payment wallets with balances on network %s: %w", w.network, err)
	}

	// Step 3: Get receiving wallet (address and private key)
	account, privKey, err := payment.GetReceivingWallet(w.mnemonic, w.passphrase, w.salt)
	if err != nil {
		return fmt.Errorf("failed to get receiving wallet on network %s: %w", w.network, err)
	}

	receivingAddr := account.Address.Hex()
	receivingPrivKey, err := crypto.PrivateKeyToHex(privKey)
	if err != nil {
		return fmt.Errorf("failed to convert private key to hex: %w", err)
	}

	// Loop through each token contract
	for _, tokenAddr := range w.tokenContractAddresses {
		decimals, err := blockchain.GetTokenDecimalsFromCache(tokenAddr, w.network.String(), w.cacheRepo)
		if err != nil {
			logger.GetLogger().Errorf("Skipping token %s: failed to get decimals: %v", tokenAddr, err)
			continue
		}

		// Get token symbol
		tokenSymbol, err := conf.GetTokenSymbol(tokenAddr)
		if err != nil {
			return fmt.Errorf("failed to get token symbol from token contract address %s: %w", tokenAddr, err)
		}

		addressWalletMap := w.mapWallets(wallets, w.network.String(), tokenSymbol, decimals)

		for address, walletInfo := range addressWalletMap {
			if walletInfo.TokenAmount == nil {
				continue
			}
			err := w.processWallet(
				ctx, address, nativeTokenSymbol, receivingAddr, receivingPrivKey, walletInfo, decimals, tokenAddr, tokenSymbol,
			)
			if err != nil {
				logger.GetLogger().Errorf(
					"Failed to process wallet %s for token %s on network %s: %v", address, tokenAddr, w.network, err,
				)
			}
			time.Sleep(constants.DefaultNetworkDelay)
		}

		if err := w.transferFromReceivingToMasterWallet(ctx, receivingAddr, receivingPrivKey, decimals, tokenAddr, tokenSymbol); err != nil {
			logger.GetLogger().Errorf(
				"Failed to transfer from receiving to master for token %s on network %s: %v", tokenSymbol, w.network, err,
			)
		}
		time.Sleep(constants.DefaultNetworkDelay)
	}

	return nil
}

func (w *paymentWalletWithdrawWorker) transferFromReceivingToMasterWallet(
	ctx context.Context,
	receivingWalletAddress, receivingWalletPrivateKey string,
	decimals uint8,
	tokenAddress, tokenSymbol string,
) error {
	// Step 1: Get the balance of the token in the receiving wallet
	tokenBalance, err := w.ethClient.GetTokenBalance(ctx, tokenAddress, receivingWalletAddress)
	if err != nil {
		return fmt.Errorf("failed to get %s balance for receiving wallet on network %s: %w", tokenSymbol, w.network, err)
	}

	// Step 2: Skip transfer if no balance
	if tokenBalance.Cmp(big.NewInt(0)) == 0 {
		logger.GetLogger().Infof(
			"No %s balance in receiving wallet on network %s. Skipping transfer to master wallet", tokenSymbol, w.network,
		)
		return nil
	}

	// Enforce minimum withdrawal threshold
	minThreshold := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil) // 10^decimals
	minThreshold.Mul(minThreshold, big.NewInt(constants.MinimumWithdrawThreshold))
	if tokenBalance.Cmp(minThreshold) < 0 {
		logger.GetLogger().Infof(
			"Withdrawal amount from receiving wallet to master wallet on network %s is below %d %s. Skipping withdrawal.",
			w.network,
			constants.MinimumWithdrawThreshold,
			tokenSymbol,
		)
		return nil
	}

	// Step 3: Perform the transfer to the master wallet
	txHash, gasUsed, gasPrice, receiptStatus, err := w.ethClient.TransferToken(
		ctx,
		w.chainID,
		tokenAddress,
		receivingWalletPrivateKey,
		w.masterWalletAddress,
		tokenBalance,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to transfer %s from receiving wallet to master wallet on network %s: %w",
			tokenSymbol,
			w.network,
			err,
		)
	}

	// Step 4: Calculate fee and create the payload
	fee := utils.CalculateFee(gasUsed, gasPrice)
	tokenAmount, err := utils.ConvertSmallestUnitToFloatToken(tokenBalance.String(), decimals)
	if err != nil {
		logger.GetLogger().Errorf(
			"Failed to convert token amount for transfer %s on network %s: %v", tokenSymbol, w.network, err,
		)
		return err
	}

	status := true
	errorMessage := ""

	if receiptStatus != 1 {
		status = false
		errorMessage = "execution reverted"
	}

	payload := dto.TokenTransferHistoryDTO{
		Network:         w.network.String(),
		TransactionHash: txHash.Hex(),
		FromAddress:     receivingWalletAddress,
		ToAddress:       w.masterWalletAddress,
		TokenAmount:     tokenAmount,
		Status:          status,
		Symbol:          tokenSymbol,
		ErrorMessage:    errorMessage,
		Fee:             fee,
		Type:            constants.Withdraw,
	}

	// Step 5: Persist transfer history
	err = w.tokenTransferUCase.CreateTokenTransferHistories(ctx, []dto.TokenTransferHistoryDTO{payload})
	if err != nil {
		logger.GetLogger().Errorf("Failed to create token transfer history for receiving wallet transfer on network %s: %v", w.network, err)
		return err
	}

	// Step 6: Log successful transfer
	if receiptStatus == 1 {
		logger.GetLogger().Infof(
			"Transferred %s %s from receiving wallet to master wallet  on network %s. Transaction hash: %s. Fee: %s",
			tokenBalance.String(), tokenSymbol, w.network, txHash.Hex(), fee,
		)
	} else {
		logger.GetLogger().Errorf(
			"%s transfer failed from receiving wallet to master wallet on network %s. Transaction hash: %s",
			tokenSymbol, w.network, txHash.Hex(),
		)
	}

	return nil
}

func (w *paymentWalletWithdrawWorker) mapWallets(
	wallets []dto.PaymentWalletBalanceDTO,
	network string,
	tokenSymbol string,
	decimals uint8,
) map[string]walletInfo {
	addressWalletMap := make(map[string]walletInfo)
	for _, wallet := range wallets {
		for _, networkBalance := range wallet.NetworkBalances {
			if networkBalance.Network != network {
				continue
			}
			for _, tokenBalance := range networkBalance.TokenBalances {
				if tokenBalance.Symbol != tokenSymbol {
					continue
				}
				amount, err := utils.ConvertFloatTokenToSmallestUnit(tokenBalance.Amount, decimals)
				if err != nil {
					logger.GetLogger().Errorf("Failed to convert amount for wallet %s on network %s: %v", wallet.Address, network, err)
					continue
				}
				addressWalletMap[wallet.Address] = walletInfo{
					ID:          wallet.ID,
					TokenAmount: amount,
				}
			}
		}
	}
	return addressWalletMap
}

func (w *paymentWalletWithdrawWorker) processWallet(
	ctx context.Context,
	address, nativeTokenSymbol, receivingWalletAddress, receivingWalletPrivateKey string,
	walletInfo walletInfo, decimals uint8,
	tokenAddress, tokenSymbol string,
) error {
	// Step 1: Generate account and validate
	account, privateKey, err := crypto.GenerateAccount(w.mnemonic, w.passphrase, w.salt, constants.PaymentWallet, walletInfo.ID)
	if err != nil || account.Address.Hex() != address {
		return fmt.Errorf("account generation or address mismatch: %v", err)
	}

	privateKeyHex, err := crypto.PrivateKeyToHex(privateKey)
	if err != nil {
		return fmt.Errorf("failed to convert private key: %w", err)
	}

	// Step 2: Check if the payment wallet has a balance (onchain check)
	tokenAmount, err := w.ethClient.GetTokenBalance(ctx, tokenAddress, address)
	if err != nil {
		return fmt.Errorf(
			"failed to get %s balance for payment wallet %s on network %s: %w", tokenSymbol, address, w.network, err,
		)
	}
	if tokenAmount.Cmp(big.NewInt(0)) == 0 {
		logger.GetLogger().Infof(
			"No %s balance in payment wallet %s on network %s. Skipping transfer to receiving wallet", tokenSymbol, address, w.network,
		)
		return nil
	}

	// Withdraw the minimum of the wallet balance and the token amount
	withdrawAmount := tokenAmount
	if walletInfo.TokenAmount.Cmp(tokenAmount) < 0 {
		withdrawAmount = walletInfo.TokenAmount
	}

	// Enforce minimum withdrawal threshold
	minThreshold := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil) // 10^decimals
	minThreshold.Mul(minThreshold, big.NewInt(constants.MinimumWithdrawThreshold))
	if withdrawAmount.Cmp(minThreshold) < 0 {
		logger.GetLogger().Infof(
			"Withdrawal amount for wallet %s on network %s is below %d %s. Skipping withdrawal.",
			address,
			w.network,
			constants.MinimumWithdrawThreshold,
			tokenSymbol,
		)
		return nil
	}

	// Step 3: Calculate required gas
	requiredGas, err := w.calculateRequiredGas(ctx, address, receivingWalletAddress, withdrawAmount, tokenAddress)
	if err != nil {
		return fmt.Errorf("failed to calculate required gas for wallet %s on network %s: %w", address, w.network, err)
	}

	var payloads []dto.TokenTransferHistoryDTO

	// Step 4: Transfer native token for gas if required
	if requiredGas.Cmp(big.NewInt(0)) > 0 {
		txHash, gasUsed, gasPrice, err := w.ethClient.TransferNativeToken(
			ctx, w.chainID, receivingWalletPrivateKey, address, requiredGas,
		)
		if err != nil {
			return fmt.Errorf("failed to transfer native token to %s on network %s: %w", address, w.network, err)
		}

		fee := utils.CalculateFee(gasUsed, gasPrice)
		nativeAmount, err := utils.ConvertSmallestUnitToFloatToken(requiredGas.String(), constants.NativeTokenDecimalPlaces)
		if err != nil {
			return fmt.Errorf(
				"failed to convert token amount for %s transfer on network %s: %v", tokenSymbol, w.network, err,
			)
		}

		payloads = append(payloads, dto.TokenTransferHistoryDTO{
			Network:         w.network.String(),
			TransactionHash: txHash.Hex(),
			FromAddress:     receivingWalletAddress,
			ToAddress:       address,
			TokenAmount:     nativeAmount,
			Status:          true,
			Symbol:          nativeTokenSymbol,
			ErrorMessage:    "",
			Fee:             fee,
			Type:            constants.InternalTransfer,
		})
		logger.GetLogger().Infof("Native token sent to %s for gas on network %s. Transaction hash: %s", address, w.network, txHash.Hex())
	}

	// Step 5: Transfer token to the receiving wallet
	txHash, gasUsed, gasPrice, receiptStatus, err := w.ethClient.TransferToken(
		ctx, w.chainID,
		tokenAddress,
		privateKeyHex,
		receivingWalletAddress,
		withdrawAmount,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to transfer %s from payment wallet %s to receiving wallet on network %s: %w",
			tokenSymbol,
			address,
			w.network,
			err,
		)
	}

	fee := utils.CalculateFee(gasUsed, gasPrice)
	withdrawAmountStr, err := utils.ConvertSmallestUnitToFloatToken(withdrawAmount.String(), decimals)
	if err != nil {
		logger.GetLogger().Errorf(
			"Failed to convert token amount for transfer %s on network %s: %v", tokenSymbol, w.network, withdrawAmount,
		)
	}

	status := true
	errorMessage := ""
	if receiptStatus == 1 {
		// Update payment wallet balance
		if err = w.paymentWalletUCase.SubtractPaymentWalletBalance(ctx, walletInfo.ID, withdrawAmountStr, w.network, tokenSymbol); err != nil {
			logger.GetLogger().Errorf("Failed to subtract payment wallet balance on network %s: %v", w.network, err)
			return err
		}
	} else {
		status = false
		errorMessage = "execution reverted"
	}

	payloads = append(payloads, dto.TokenTransferHistoryDTO{
		Network:         w.network.String(),
		TransactionHash: txHash.Hex(),
		FromAddress:     address,
		ToAddress:       receivingWalletAddress,
		TokenAmount:     withdrawAmountStr,
		Status:          status,
		Symbol:          tokenSymbol,
		ErrorMessage:    errorMessage,
		Fee:             fee,
		Type:            constants.InternalTransfer,
	})
	logger.GetLogger().Infof(
		"%s transferred from %s to receiving wallet on network %s. Transaction hash: %s",
		tokenSymbol,
		address,
		w.network,
		txHash.Hex(),
	)

	// Step 6: Persist transfer histories
	if err := w.tokenTransferUCase.CreateTokenTransferHistories(w.ctx, payloads); err != nil {
		logger.GetLogger().Errorf("Failed to create token transfer histories on network %s: %v", w.network, err)
		return err
	}

	return nil
}

// calculateRequiredGas calculates the required gas for a wallet withdrawal.
func (w *paymentWalletWithdrawWorker) calculateRequiredGas(
	ctx context.Context, address, receivingWalletAddress string, tokenAmount *big.Int, tokenAddress string,
) (*big.Int, error) {
	// Step 1: Estimate the gas required for ERC20 token transfer
	estimatedGas, err := w.ethClient.EstimateGasGeneric(
		common.HexToAddress(tokenAddress),           // Contract address
		common.HexToAddress(address),                // From address
		erc20token.Erc20tokenMetaData.ABI,           // ERC-20 ABI (adjust if using a different standard)
		"transfer",                                  // Method name (e.g., "transfer" for ERC-20)
		common.HexToAddress(receivingWalletAddress), // To address
		tokenAmount,                                 // Amount
	)
	if err != nil {
		return nil, fmt.Errorf("failed to estimate gas on network %s: %w", w.network, err)
	}

	// Step 2: Fetch gas tip cap and base fee for dynamic fee transactions
	gasTipCap, err := w.ethClient.SuggestGasTipCap(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to suggest gas tip cap on network %s: %w", w.network, err)
	}

	baseFee, err := w.ethClient.GetBaseFee(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get base fee on network %s: %w", w.network, err)
	}

	// Step 3: Calculate max fee per gas (baseFee + 2 * gasTipCap)
	gasFeeCap := new(big.Int).Add(baseFee, new(big.Int).Mul(gasTipCap, big.NewInt(2)))

	// Step 4: Calculate the required gas cost
	requiredGas := new(big.Int).Mul(big.NewInt(int64(estimatedGas)), gasFeeCap)

	// Step 5: Apply the gas buffer multiplier
	multiplier := big.NewFloat(w.gasBufferMultiplier)
	finalGas := new(big.Float).Mul(new(big.Float).SetInt(requiredGas), multiplier)
	finalGas.Int(requiredGas) // `requiredGas` now contains the final gas value with the buffer applied.

	// Step 6: Get the native token balance of the wallet
	nativeTokenAmount, err := w.ethClient.GetNativeTokenBalance(ctx, address)
	if err != nil {
		return nil, fmt.Errorf("failed to get native token balance on network %s: %w", w.network, err)
	}

	// Step 7: Adjust the final gas cost (stored in `requiredGas`) based on the existing native token balance
	if nativeTokenAmount != nil && nativeTokenAmount.Cmp(requiredGas) < 0 {
		requiredGas.Sub(requiredGas, nativeTokenAmount)
	}

	// Step 8: Ensure the required gas is not negative
	if requiredGas.Cmp(big.NewInt(0)) < 0 {
		requiredGas.Set(big.NewInt(0))
	}

	return requiredGas, nil
}
