package workers

import (
	"context"
	"sync"
	"time"

	"github.com/genefriendway/onchain-handler/constants"
	infrainterfaces "github.com/genefriendway/onchain-handler/infra/interfaces"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
	"github.com/genefriendway/onchain-handler/pkg/blockchain"
	"github.com/genefriendway/onchain-handler/pkg/logger"
	"github.com/genefriendway/onchain-handler/pkg/utils"
)

type paymentWalletBalanceWorker struct {
	paymentWalletUCase   interfaces.PaymentWalletUCase
	cacheRepo            infrainterfaces.CacheRepository
	network              constants.NetworkType
	rpcURL               string
	tokenContractAddress string
	tokenSymbol          string
	isRunning            bool
	mu                   sync.Mutex
}

func NewPaymentWalletBalanceWorker(
	paymentWalletUCase interfaces.PaymentWalletUCase,
	cacheRepo infrainterfaces.CacheRepository,
	network constants.NetworkType,
	rpcURL string,
	tokenContractAddress string,
	tokenSymbol string,
) interfaces.Worker {
	return &paymentWalletBalanceWorker{
		paymentWalletUCase:   paymentWalletUCase,
		network:              network,
		rpcURL:               rpcURL,
		tokenContractAddress: tokenContractAddress,
		tokenSymbol:          tokenSymbol,
		cacheRepo:            cacheRepo,
	}
}

func (w *paymentWalletBalanceWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(constants.PaymentWalletBalanceFetchInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			go w.run(ctx)
		case <-ctx.Done():
			logger.GetLogger().Info("Shutting down paymentWalletBalanceWorker")
			return
		}
	}
}

func (w *paymentWalletBalanceWorker) run(ctx context.Context) {
	w.mu.Lock()
	if w.isRunning {
		logger.GetLogger().Warn("Previous paymentWalletBalanceWorker run still in progress, skipping this cycle")
		w.mu.Unlock()
		return
	}

	// Mark as running
	w.isRunning = true
	w.mu.Unlock()

	// Fetch wallets
	wallets, err := w.paymentWalletUCase.GetPaymentWallets(ctx)
	if err != nil {
		logger.GetLogger().Errorf("Failed to get payment wallets: %v", err)
		return
	}

	var walletIDs []uint64
	var addresses []string
	for _, wallet := range wallets {
		walletIDs = append(walletIDs, wallet.ID)
		addresses = append(addresses, wallet.Address)
	}

	// Get token decimals
	decimals, err := blockchain.GetTokenDecimalsFromCache(w.tokenContractAddress, string(w.network), w.cacheRepo)
	if err != nil {
		logger.GetLogger().Errorf("Failed to get token decimals: %v", err)
		return
	}

	// Fetch token balances
	mapAddressesTokenBalances, err := blockchain.GetTokenBalances(w.rpcURL, w.tokenContractAddress, decimals, addresses)
	if err != nil {
		logger.GetLogger().Errorf("Failed to get USDT token (%s) balances on %s: %v", w.tokenContractAddress, w.network, err)
		return
	}

	// Fetch native token balances
	mapAddressesNativeTokenBalances, err := blockchain.GetNativeTokenBalances(w.rpcURL, addresses)
	if err != nil {
		logger.GetLogger().Errorf("Failed to get native token balances on %s: %v", w.network, err)
		return
	}

	var tokenBalances []string
	var nativeTokenBalances []string
	for index := range walletIDs {
		tokenBalance := mapAddressesTokenBalances[addresses[index]]
		tokenBalances = append(tokenBalances, tokenBalance)
		nativeTokenBalance := mapAddressesNativeTokenBalances[addresses[index]]
		nativeTokenBalances = append(nativeTokenBalances, nativeTokenBalance)
	}

	// Upsert balances in the database
	err = w.paymentWalletUCase.UpsertPaymentWalletBalances(ctx, walletIDs, tokenBalances, w.network, w.tokenSymbol)
	if err != nil {
		logger.GetLogger().Errorf("Failed to batch upsert payment wallet balances: %v", err)
		return
	}

	// Get native token symbol
	nativeTokenSymbol, err := utils.GetNativeTokenSymbol(w.network)
	if err != nil {
		logger.GetLogger().Errorf("Failed to get native token symbol: %v", err)
		return
	}
	// Upsert balances in the database
	err = w.paymentWalletUCase.UpsertPaymentWalletBalances(ctx, walletIDs, nativeTokenBalances, w.network, nativeTokenSymbol)
	if err != nil {
		logger.GetLogger().Errorf("Failed to batch upsert payment wallet native token balances: %v", err)
		return
	}

	// Mark as not running
	w.mu.Lock()
	w.isRunning = false
	w.mu.Unlock()
}
