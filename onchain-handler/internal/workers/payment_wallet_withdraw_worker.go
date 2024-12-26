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
	ID                uint64
	TokenAmount       *big.Int
	NativeTokenAmount *big.Int
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

		logger.GetLogger().Infof("Next payment wallet withdrawal scheduled in: %s", sleepDuration)

		// Sleep until the next scheduled time or exit early if the context is canceled
		select {
		case <-time.After(sleepDuration):
			go w.run(ctx)
		case <-ctx.Done():
			logger.GetLogger().Info("Shutting down paymentWalletWithdrawWorker")
			return
		}
	}
}

func (w *paymentWalletWithdrawWorker) run(ctx context.Context) {
	w.mu.Lock()
	if w.isRunning {
		logger.GetLogger().Warn("Previous paymentWalletWithdrawWorker run still in progress, skipping this cycle")
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
			logger.GetLogger().Infof("Withdrawal process succeeded on attempt %d", attempt)
			return
		}

		logger.GetLogger().Errorf("Withdrawal process failed on attempt %d: %v", attempt, lastErr)

		// Check if the context has been canceled to avoid infinite retries when shutting down
		select {
		case <-ctx.Done():
			logger.GetLogger().Info("Withdrawal process stopped due to context cancellation")
			return
		case <-time.After(constants.RetryDelay): // Wait before retrying
		}
	}

	logger.GetLogger().Errorf("Withdrawal process failed after %d attempts: %v", constants.MaxRetries, lastErr)
}

func (w *paymentWalletWithdrawWorker) withdraw(ctx context.Context) error {
	// Step 1: Get native token symbol
	nativeTokenSymbol, err := blockchain.GetNativeTokenSymbol(w.network)
	if err != nil {
		return fmt.Errorf("failed to get native token symbol: %w", err)
	}

	// Step 2: Get token decimals from cache
	decimals, err := blockchain.GetTokenDecimalsFromCache(w.tokenContractAddress, string(w.network), w.cacheRepo)
	if err != nil {
		return fmt.Errorf("failed to get token decimals from cache: %w", err)
	}

	// Step 3: Fetch payment wallets with balances
	wallets, err := w.paymentWalletUCase.GetPaymentWalletsWithBalances(ctx, true, &w.network)
	if err != nil {
		return fmt.Errorf("failed to get payment wallets with balances: %w", err)
	}

	// Step 4: Get receiving wallet (address and private key)
	account, privateKey, err := payment.GetReceivingWallet(w.mnemonic, w.passphrase, w.salt)
	if err != nil {
		return fmt.Errorf("failed to get receiving wallet: %w", err)
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
			logger.GetLogger().Warnf("Skipping withdrawal for wallet %s due to nil token amount", address)
			continue
		}
		if err := w.processWallet(
			ctx, address, nativeTokenSymbol, receivingWalletAddress, receivingWalletPrivateKey, walletInfo, decimals,
		); err != nil {
			logger.GetLogger().Errorf("Failed to process wallet %s: %v", address, err)
		}
	}

	// Step 7: Transfer all USDT tokens from receiving wallet to master wallet
	if err := w.transferFromReceivingToMasterWallet(ctx, receivingWalletAddress, receivingWalletPrivateKey, decimals); err != nil {
		return fmt.Errorf("failed to transfer from receiving wallet to master wallet: %w", err)
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
		return fmt.Errorf("failed to get USDT balance for receiving wallet: %w", err)
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
		return fmt.Errorf("failed to transfer USDT to master wallet: %w", err)
	}

	// Step 4: Calculate fee and create the payload
	fee := utils.CalculateFee(gasUsed, gasPrice)
	tokenAmount, err := utils.ConvertSmallestUnitToFloatToken(usdtBalance.String(), decimals)
	if err != nil {
		logger.GetLogger().Errorf("Failed to convert token amount for transfer: %v", err)
		return err
	}

	payload := dto.TokenTransferHistoryDTO{
		Network:         string(w.network),
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
		logger.GetLogger().Errorf("Failed to create token transfer history for receiving wallet transfer: %v", err)
		return err
	}

	// Step 6: Log successful transfer
	logger.GetLogger().Infof(
		"Transferred %s USDT from receiving wallet to master wallet. Transaction hash: %s. Fee: %s",
		usdtBalance.String(), txHash.Hex(), fee,
	)

	return nil
}

func (w *paymentWalletWithdrawWorker) mapWallets(wallets []dto.PaymentWalletBalanceDTO, nativeTokenSymbol string, decimals uint8) map[string]walletInfo {
	addressWalletMap := make(map[string]walletInfo)
	for _, wallet := range wallets {
		for _, networkBalance := range wallet.NetworkBalances {
			if networkBalance.Network == string(w.network) {
				for _, tokenBalance := range networkBalance.TokenBalances {
					if tokenBalance.Symbol == nativeTokenSymbol {
						continue
					}
					amount, err := utils.ConvertFloatTokenToSmallestUnit(tokenBalance.Amount, decimals)
					if err != nil {
						logger.GetLogger().Errorf("Failed to convert amount for wallet %s: %v", wallet.Address, err)
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

	// Step 2: Calculate required gas
	requiredGas, err := w.calculateRequiredGas(ctx, address, receivingWalletAddress, walletInfo)
	if err != nil {
		return fmt.Errorf("failed to calculate required gas for wallet %s: %w", address, err)
	}

	var payloads []dto.TokenTransferHistoryDTO

	// Step 3: Transfer native token for gas if required
	if requiredGas.Cmp(big.NewInt(0)) > 0 {
		txHash, gasUsed, gasPrice, err := w.ethClient.TransferNativeToken(ctx, w.chainID, receivingWalletPrivateKey, address, requiredGas)
		if err != nil {
			return fmt.Errorf("failed to transfer native token to %s: %w", address, err)
		}

		fee := utils.CalculateFee(gasUsed, gasPrice)
		nativeAmount, _ := utils.ConvertSmallestUnitToFloatToken(requiredGas.String(), constants.NativeTokenDecimalPlaces)

		payloads = append(payloads, dto.TokenTransferHistoryDTO{
			Network:         string(w.network),
			TransactionHash: txHash.Hex(),
			FromAddress:     receivingWalletAddress,
			ToAddress:       address,
			TokenAmount:     nativeAmount,
			Status:          true,
			Symbol:          nativeTokenSymbol,
			ErrorMessage:    "",
			Fee:             fee,
			Type:            constants.Withdraw,
		})
		logger.GetLogger().Infof("Native token sent to %s for gas. Transaction hash: %s", address, txHash.Hex())
	}

	// Step 4: Transfer USDT to the receiving wallet
	txHash, gasUsed, gasPrice, err := w.ethClient.TransferToken(
		ctx, w.chainID, w.tokenContractAddress, privateKeyHex, receivingWalletAddress, walletInfo.TokenAmount,
	)
	if err != nil {
		return fmt.Errorf("failed to transfer USDT from %s to receiving wallet: %w", address, err)
	}

	fee := utils.CalculateFee(gasUsed, gasPrice)
	tokenAmount, _ := utils.ConvertSmallestUnitToFloatToken(walletInfo.TokenAmount.String(), decimals)

	// Update payment wallet balance
	if err = w.paymentWalletUCase.SubtractPaymentWalletBalance(ctx, walletInfo.ID, tokenAmount, w.network, constants.USDT); err != nil {
		logger.GetLogger().Errorf("Failed to subtract payment wallet balance: %v", err)
		return err
	}

	payloads = append(payloads, dto.TokenTransferHistoryDTO{
		Network:         string(w.network),
		TransactionHash: txHash.Hex(),
		FromAddress:     address,
		ToAddress:       receivingWalletAddress,
		TokenAmount:     tokenAmount,
		Status:          true,
		Symbol:          constants.USDT,
		ErrorMessage:    "",
		Fee:             fee,
		Type:            constants.Withdraw,
	})
	logger.GetLogger().Infof("USDT transferred from %s to receiving wallet. Transaction hash: %s", address, txHash.Hex())

	// Step 5: Persist transfer histories
	if err := w.tokenTransferUCase.CreateTokenTransferHistories(w.ctx, payloads); err != nil {
		logger.GetLogger().Errorf("Failed to create token transfer histories: %v", err)
		return err
	}

	return nil
}

// calculateRequiredGas calculates the required gas for a wallet withdrawal.
func (w *paymentWalletWithdrawWorker) calculateRequiredGas(ctx context.Context, address, receivingWalletAddress string, walletInfo walletInfo) (*big.Int, error) {
	// Step 1: Estimate the gas required for ERC20 token transfer
	estimatedGas, err := w.ethClient.EstimateGasERC20(
		common.HexToAddress(w.tokenContractAddress),
		common.HexToAddress(address),
		common.HexToAddress(receivingWalletAddress),
		walletInfo.TokenAmount,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to estimate gas: %w", err)
	}

	// Step 2: Fetch gas tip cap and base fee for dynamic fee transactions
	gasTipCap, err := w.ethClient.SuggestGasTipCap(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to suggest gas tip cap: %w", err)
	}

	baseFee, err := w.ethClient.GetBaseFee(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get base fee: %w", err)
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
		return nil, fmt.Errorf("failed to get native token balance: %w", err)
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
