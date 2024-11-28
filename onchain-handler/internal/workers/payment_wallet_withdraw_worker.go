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
	"github.com/genefriendway/onchain-handler/pkg/utils"
)

type walletInfo struct {
	ID                uint64
	TokenAmount       *big.Int
	NativeTokenAmount *big.Int
}

type paymentWalletWithdrawWorker struct {
	ctx                    context.Context
	ethClient              pkginterfaces.Client
	network                constants.NetworkType
	chainID                uint64
	cacheRepo              infrainterfaces.CacheRepository
	tokenTransferUCase     interfaces.TokenTransferUCase
	paymentWalletUCase     interfaces.PaymentWalletUCase
	tokenContractAddress   string
	masterWalletAddress    string
	masterWalletPrivateKey string
	mnemonic               string
	passphrase             string
	salt                   string
	isRunning              bool
	mu                     sync.Mutex
}

func NewPaymentWalletWithdrawWorker(
	ctx context.Context,
	ethClient pkginterfaces.Client,
	network constants.NetworkType,
	chainID uint64,
	cacheRepo infrainterfaces.CacheRepository,
	tokenTransferUCase interfaces.TokenTransferUCase,
	paymentWalletUCase interfaces.PaymentWalletUCase,
	tokenContractAddress, masterWalletAddress, masterWalletPrivateKey,
	mnemonic, passphrase, salt string,
) interfaces.Worker {
	return &paymentWalletWithdrawWorker{
		ctx:                    ctx,
		ethClient:              ethClient,
		network:                network,
		chainID:                chainID,
		cacheRepo:              cacheRepo,
		tokenTransferUCase:     tokenTransferUCase,
		paymentWalletUCase:     paymentWalletUCase,
		tokenContractAddress:   tokenContractAddress,
		masterWalletAddress:    masterWalletAddress,
		masterWalletPrivateKey: masterWalletPrivateKey,
		mnemonic:               mnemonic,
		passphrase:             passphrase,
		salt:                   salt,
	}
}

func (w *paymentWalletWithdrawWorker) Start(ctx context.Context) {
	for {
		// Calculate the duration until the next scheduled time (e.g., midnight)
		now := time.Now()
		nextRun := time.Date(
			now.Year(), now.Month(), now.Day()+1, // Next day
			0, 0, 0, 0, // 00:00:00
			now.Location(),
		)
		sleepDuration := time.Until(nextRun)

		logger.GetLogger().Infof("Next payment wallet withdrawal scheduled at: %s", nextRun)

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

	if err := w.withdraw(ctx); err != nil {
		logger.GetLogger().Errorf("Withdrawal process failed: %v", err)
	}
}

func (w *paymentWalletWithdrawWorker) withdraw(ctx context.Context) error {
	nativeTokenSymbol, err := utils.GetNativeTokenSymbol(w.network)
	if err != nil {
		return err
	}

	decimals, err := blockchain.GetTokenDecimalsFromCache(w.tokenContractAddress, string(w.network), w.cacheRepo)
	if err != nil {
		return err
	}

	wallets, err := w.paymentWalletUCase.GetPaymentWalletsWithBalances(ctx, true, &w.network)
	if err != nil {
		return err
	}

	addressWalletMap := w.mapWallets(wallets, nativeTokenSymbol, decimals)
	for address, walletInfo := range addressWalletMap {
		if err := w.processWallet(ctx, address, nativeTokenSymbol, walletInfo, decimals); err != nil {
			logger.GetLogger().Errorf("Failed to process wallet %s: %v", address, err)
		}
	}
	return nil
}

func (w *paymentWalletWithdrawWorker) mapWallets(wallets []dto.PaymentWalletBalanceDTO, nativeTokenSymbol string, decimals uint8) map[string]walletInfo {
	addressWalletMap := make(map[string]walletInfo)
	for _, wallet := range wallets {
		for _, networkBalance := range wallet.NetworkBalances {
			if networkBalance.Network == string(w.network) {
				for _, tokenBalance := range networkBalance.TokenBalances {
					amount, err := utils.ConvertFloatTokenToSmallestUnit(tokenBalance.Amount, decimals)
					if err != nil {
						logger.GetLogger().Errorf("Failed to convert amount for wallet %s: %v", wallet.Address, err)
						continue
					}

					walletDetails := addressWalletMap[wallet.Address]
					walletDetails.ID = wallet.ID
					if tokenBalance.Symbol == constants.USDT {
						walletDetails.TokenAmount = amount
					} else if tokenBalance.Symbol == nativeTokenSymbol {
						walletDetails.NativeTokenAmount = amount
					}
					addressWalletMap[wallet.Address] = walletDetails
				}
			}
		}
	}
	return addressWalletMap
}

func (w *paymentWalletWithdrawWorker) processWallet(
	ctx context.Context, address, nativeTokenSymbol string, walletInfo walletInfo, decimals uint8,
) error {
	if walletInfo.TokenAmount == nil {
		logger.GetLogger().Warnf("Skipping withdrawal for wallet %s due to nil token amount", address)
		return nil
	}

	account, privateKey, err := crypto.GenerateAccount(w.mnemonic, w.passphrase, w.salt, constants.PaymentWallet, walletInfo.ID)
	if err != nil || account.Address.Hex() != address {
		return fmt.Errorf("account generation or address mismatch: %v", err)
	}

	privateKeyHex, err := crypto.PrivateKeyToHex(privateKey)
	if err != nil {
		return err
	}

	requiredGas, err := w.calculateRequiredGas(ctx, address, walletInfo)
	if err != nil {
		return err
	}

	var payloads []dto.TokenTransferHistoryDTO

	txHash, gasUsed, gasPrice, err := w.ethClient.TransferNativeToken(ctx, w.chainID, w.masterWalletPrivateKey, address, requiredGas)
	if err != nil {
		return err
	}
	fee := utils.CalculateFee(gasUsed, gasPrice)

	nativeAmount, err := utils.ConvertSmallestUnitToFloatToken(requiredGas.String(), constants.NativeTokenDecimalPlaces)
	if err != nil {
		logger.GetLogger().Errorf("Failed to convert native token amount: %v", err)
		return err
	}
	payloads = append(payloads, dto.TokenTransferHistoryDTO{
		Network:         string(w.network),
		TransactionHash: txHash.Hex(),
		FromAddress:     w.masterWalletAddress,
		ToAddress:       address,
		TokenAmount:     nativeAmount,
		Status:          true,
		Symbol:          nativeTokenSymbol,
		ErrorMessage:    "",
		Fee:             fee,
		Type:            constants.Withdraw,
	})
	logger.GetLogger().Infof("Native token sent to %s for gas. Transaction hash: %s", address, txHash.Hex())

	txHash, gasUsed, gasPrice, err = w.ethClient.TransferToken(ctx, w.chainID, w.tokenContractAddress, privateKeyHex, w.masterWalletAddress, walletInfo.TokenAmount)
	if err != nil {
		return err
	}
	fee = utils.CalculateFee(gasUsed, gasPrice)

	tokenAmount, err := utils.ConvertSmallestUnitToFloatToken(walletInfo.TokenAmount.String(), decimals)
	if err != nil {
		logger.GetLogger().Errorf("Failed to convert native token amount: %v", err)
		return err
	}
	payloads = append(payloads, dto.TokenTransferHistoryDTO{
		Network:         string(w.network),
		TransactionHash: txHash.Hex(),
		FromAddress:     address,
		ToAddress:       w.masterWalletAddress,
		TokenAmount:     tokenAmount,
		Status:          true,
		Symbol:          constants.USDT,
		ErrorMessage:    "",
		Fee:             fee,
		Type:            constants.Withdraw,
	})
	logger.GetLogger().Infof("USDT transferred from %s to master wallet. Transaction hash: %s", address, txHash.Hex())

	err = w.tokenTransferUCase.CreateTokenTransferHistories(w.ctx, payloads)
	if err != nil {
		logger.GetLogger().Errorf("Failed to create token transfer histories: %v", err)
		return err
	}

	return nil
}

func (w *paymentWalletWithdrawWorker) calculateRequiredGas(ctx context.Context, address string, walletInfo walletInfo) (*big.Int, error) {
	estimatedGas, err := w.ethClient.EstimateGasERC20(
		common.HexToAddress(w.tokenContractAddress),
		common.HexToAddress(address),
		common.HexToAddress(w.masterWalletAddress),
		walletInfo.TokenAmount,
	)
	if err != nil {
		return nil, err
	}

	gasPrice, err := w.ethClient.SuggestGasPrice(ctx)
	if err != nil {
		return nil, err
	}

	requiredGas := new(big.Int).Mul(big.NewInt(int64(estimatedGas)), gasPrice)
	buffer := new(big.Int).Div(new(big.Int).Mul(requiredGas, big.NewInt(15)), big.NewInt(100))
	requiredGas.Add(requiredGas, buffer)

	if walletInfo.NativeTokenAmount != nil && walletInfo.NativeTokenAmount.Cmp(requiredGas) < 0 {
		requiredGas.Sub(requiredGas, walletInfo.NativeTokenAmount)
	}
	return requiredGas, nil
}
