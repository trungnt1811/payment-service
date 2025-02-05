package workers

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/genefriendway/onchain-handler/constants"
	infrainterfaces "github.com/genefriendway/onchain-handler/infra/interfaces"
	"github.com/genefriendway/onchain-handler/internal/dto"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
	"github.com/genefriendway/onchain-handler/pkg/blockchain"
	"github.com/genefriendway/onchain-handler/pkg/crypto"
	pkginterfaces "github.com/genefriendway/onchain-handler/pkg/interfaces"
	"github.com/genefriendway/onchain-handler/pkg/logger"
	"github.com/genefriendway/onchain-handler/pkg/payment"
	"github.com/genefriendway/onchain-handler/pkg/utils"
)

type walletInfo struct {
	ID          uint64
	TokenAmount *big.Int
}

type paymentWalletWithdrawWorker struct {
	ctx                  context.Context
	ethClient            pkginterfaces.Client
	network              constants.NetworkType
	chainID              uint64
	cacheRepo            infrainterfaces.CacheRepository
	tokenTransferUCase   interfaces.TokenTransferUCase
	paymentWalletUCase   interfaces.PaymentWalletUCase
	tokenContractAddress string
	masterWalletAddress  string
	mnemonic             string
	passphrase           string
	salt                 string
	gasBufferMultiplier  float64
	withdrawInterval     string
	isRunning            bool
	mu                   sync.Mutex
}

func NewPaymentWalletWithdrawWorker(
	ctx context.Context,
	ethClient pkginterfaces.Client,
	network constants.NetworkType,
	chainID uint64,
	cacheRepo infrainterfaces.CacheRepository,
	tokenTransferUCase interfaces.TokenTransferUCase,
	paymentWalletUCase interfaces.PaymentWalletUCase,
	tokenContractAddress string,
	masterWalletAddress string,
	mnemonic, passphrase, salt string,
	gasBufferMultiplier float64,
	withdrawInterval string,
) interfaces.Worker {
	return &paymentWalletWithdrawWorker{
		ctx:                  ctx,
		ethClient:            ethClient,
		network:              network,
		chainID:              chainID,
		cacheRepo:            cacheRepo,
		tokenTransferUCase:   tokenTransferUCase,
		paymentWalletUCase:   paymentWalletUCase,
		tokenContractAddress: tokenContractAddress,
		masterWalletAddress:  masterWalletAddress,
		mnemonic:             mnemonic,
		passphrase:           passphrase,
		salt:                 salt,
		gasBufferMultiplier:  gasBufferMultiplier,
		withdrawInterval:     withdrawInterval,
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

	// Step 2: Get token decimals from cache
	decimals, err := blockchain.GetTokenDecimalsFromCache(w.tokenContractAddress, w.network.String(), w.cacheRepo)
	if err != nil {
		return fmt.Errorf("failed to get token decimals from cache on network %s: %w", w.network, err)
	}

	// Step 3: Fetch payment wallets with balances
	wallets, err := w.paymentWalletUCase.GetPaymentWalletsWithBalances(ctx, true, &w.network)
	if err != nil {
		return fmt.Errorf("failed to get payment wallets with balances on network %s: %w", w.network, err)
	}

	// Step 4: Get receiving wallet (address and private key)
	account, privateKey, err := payment.GetReceivingWallet(w.mnemonic, w.passphrase, w.salt)
	if err != nil {
		return fmt.Errorf("failed to get receiving wallet on network %s: %w", w.network, err)
	}

	receivingWalletAddress := account.Address.Hex()
	receivingWalletPrivateKey, err := crypto.PrivateKeyToHex(privateKey)
	if err != nil {
		return fmt.Errorf("failed to convert private key to hex: %w", err)
	}

	// Step 5: Map wallets to address information
	addressWalletMap := w.mapWallets(wallets, nativeTokenSymbol, decimals)

	// Step 6: Process each wallet
	for address, walletInfo := range addressWalletMap {
		if walletInfo.TokenAmount == nil {
			logger.GetLogger().Warnf("Skipping withdrawal for wallet %s due to nil token amount on network %s", address, w.network)
			continue
		}
		if err := w.processWallet(
			ctx, address, nativeTokenSymbol, receivingWalletAddress, receivingWalletPrivateKey, walletInfo, decimals,
		); err != nil {
			logger.GetLogger().Errorf("Failed to process wallet %s on network %s: %v", address, w.network, err)
		}
		// Delay between wallet processing to avoid spamming the network
		time.Sleep(constants.DefaultNetworkDelay)
	}

	// Step 7: Transfer all USDT tokens from receiving wallet to master wallet
	if err := w.transferFromReceivingToMasterWallet(ctx, receivingWalletAddress, receivingWalletPrivateKey, decimals); err != nil {
		return fmt.Errorf("failed to transfer from receiving wallet to master wallet on network %s: %w", w.network, err)
	}

	return nil
}

func (w *paymentWalletWithdrawWorker) transferFromReceivingToMasterWallet(
	ctx context.Context,
	receivingWalletAddress, receivingWalletPrivateKey string,
	decimals uint8,
) error {
	// Step 1: Get the balance of the USDT token in the receiving wallet
	usdtBalance, err := w.ethClient.GetTokenBalance(ctx, w.tokenContractAddress, receivingWalletAddress)
	if err != nil {
		return fmt.Errorf("failed to get USDT balance for receiving wallet on network %s: %w", w.network, err)
	}

	// Step 2: Skip transfer if no balance
	if usdtBalance.Cmp(big.NewInt(0)) == 0 {
		logger.GetLogger().Infof("No USDT balance in receiving wallet in network %s. Skipping transfer to master wallet", w.network)
		return nil
	}

	// Step 3: Perform the transfer to the master wallet
	txHash, gasUsed, gasPrice, err := w.ethClient.TransferToken(
		ctx,
		w.chainID,
		w.tokenContractAddress,
		receivingWalletPrivateKey,
		w.masterWalletAddress,
		usdtBalance,
	)
	if err != nil {
		return fmt.Errorf("failed to transfer USDT to master wallet on network %s: %w", w.network, err)
	}

	// Step 4: Calculate fee and create the payload
	fee := utils.CalculateFee(gasUsed, gasPrice)
	tokenAmount, err := utils.ConvertSmallestUnitToFloatToken(usdtBalance.String(), decimals)
	if err != nil {
		logger.GetLogger().Errorf("Failed to convert token amount for transfer USDT on network %s: %v", w.network, err)
		return err
	}

	payload := dto.TokenTransferHistoryDTO{
		Network:         w.network.String(),
		TransactionHash: txHash.Hex(),
		FromAddress:     receivingWalletAddress,
		ToAddress:       w.masterWalletAddress,
		TokenAmount:     tokenAmount,
		Status:          true,
		Symbol:          constants.USDT,
		ErrorMessage:    "",
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
	logger.GetLogger().Infof(
		"Transferred %s USDT from receiving wallet to master wallet  on network %s. Transaction hash: %s. Fee: %s",
		usdtBalance.String(), w.network, txHash.Hex(), fee,
	)

	return nil
}

func (w *paymentWalletWithdrawWorker) mapWallets(wallets []dto.PaymentWalletBalanceDTO, nativeTokenSymbol string, decimals uint8) map[string]walletInfo {
	addressWalletMap := make(map[string]walletInfo)
	for _, wallet := range wallets {
		for _, networkBalance := range wallet.NetworkBalances {
			if networkBalance.Network == w.network.String() {
				for _, tokenBalance := range networkBalance.TokenBalances {
					if tokenBalance.Symbol == nativeTokenSymbol {
						continue
					}
					amount, err := utils.ConvertFloatTokenToSmallestUnit(tokenBalance.Amount, decimals)
					if err != nil {
						logger.GetLogger().Errorf("Failed to convert amount for wallet %s on network %s: %v", wallet.Address, w.network, err)
						continue
					}
					walletDetails := addressWalletMap[wallet.Address]
					walletDetails.ID = wallet.ID
					walletDetails.TokenAmount = amount
					addressWalletMap[wallet.Address] = walletDetails
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
	tokenAmount, err := w.ethClient.GetTokenBalance(ctx, w.tokenContractAddress, address)
	if err != nil {
		return fmt.Errorf("failed to get USDT balance for payment wallet %s on network %s: %w", address, w.network, err)
	}
	if tokenAmount.Cmp(big.NewInt(0)) == 0 {
		logger.GetLogger().Infof("No USDT balance in payment wallet %s in network %s. Skipping transfer to receiving wallet", address, w.network)
		return nil
	}

	// Step 3: Calculate required gas
	requiredGas, err := w.calculateRequiredGas(ctx, address, receivingWalletAddress, tokenAmount)
	if err != nil {
		return fmt.Errorf("failed to calculate required gas for wallet %s on network %s: %w", address, w.network, err)
	}

	var payloads []dto.TokenTransferHistoryDTO

	// Step 4: Transfer native token for gas if required
	if requiredGas.Cmp(big.NewInt(0)) > 0 {
		txHash, gasUsed, gasPrice, err := w.ethClient.TransferNativeToken(ctx, w.chainID, receivingWalletPrivateKey, address, requiredGas)
		if err != nil {
			return fmt.Errorf("failed to transfer native token to %s on network %s: %w", address, w.network, err)
		}

		fee := utils.CalculateFee(gasUsed, gasPrice)
		nativeAmount, _ := utils.ConvertSmallestUnitToFloatToken(requiredGas.String(), constants.NativeTokenDecimalPlaces)

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

	// Step 5: Transfer USDT to the receiving wallet
	txHash, gasUsed, gasPrice, err := w.ethClient.TransferToken(
		ctx, w.chainID, w.tokenContractAddress, privateKeyHex, receivingWalletAddress, tokenAmount,
	)
	if err != nil {
		return fmt.Errorf("failed to transfer USDT from %s to receiving wallet on network %s: %w", address, w.network, err)
	}

	fee := utils.CalculateFee(gasUsed, gasPrice)
	tokenAmountStr, err := utils.ConvertSmallestUnitToFloatToken(tokenAmount.String(), decimals)
	if err != nil {
		logger.GetLogger().Errorf("Failed to convert token amount for transfer USDT on network %s: %v", w.network, tokenAmount)
	}

	// Update payment wallet balance
	if err = w.paymentWalletUCase.SubtractPaymentWalletBalance(ctx, walletInfo.ID, tokenAmountStr, w.network, constants.USDT); err != nil {
		logger.GetLogger().Errorf("Failed to subtract payment wallet balance on network %s: %v", w.network, err)
		return err
	}

	payloads = append(payloads, dto.TokenTransferHistoryDTO{
		Network:         w.network.String(),
		TransactionHash: txHash.Hex(),
		FromAddress:     address,
		ToAddress:       receivingWalletAddress,
		TokenAmount:     tokenAmountStr,
		Status:          true,
		Symbol:          constants.USDT,
		ErrorMessage:    "",
		Fee:             fee,
		Type:            constants.InternalTransfer,
	})
	logger.GetLogger().Infof("USDT transferred from %s to receiving wallet on network %s. Transaction hash: %s", address, w.network, txHash.Hex())

	// Step 6: Persist transfer histories
	if err := w.tokenTransferUCase.CreateTokenTransferHistories(w.ctx, payloads); err != nil {
		logger.GetLogger().Errorf("Failed to create token transfer histories on network %s: %v", w.network, err)
		return err
	}

	return nil
}

// calculateRequiredGas calculates the required gas for a wallet withdrawal.
func (w *paymentWalletWithdrawWorker) calculateRequiredGas(ctx context.Context, address, receivingWalletAddress string, tokenAmount *big.Int) (*big.Int, error) {
	// Step 1: Estimate the gas required for ERC20 token transfer
	estimatedGas, err := w.ethClient.EstimateGasERC20(
		common.HexToAddress(w.tokenContractAddress),
		common.HexToAddress(address),
		common.HexToAddress(receivingWalletAddress),
		tokenAmount,
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
