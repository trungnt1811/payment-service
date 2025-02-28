package ucases

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/genefriendway/onchain-handler/conf"
	"github.com/genefriendway/onchain-handler/constants"
	repotypes "github.com/genefriendway/onchain-handler/internal/adapters/repositories/types"
	"github.com/genefriendway/onchain-handler/internal/delivery/dto"
	ucasetypes "github.com/genefriendway/onchain-handler/internal/domain/ucases/types"
	"github.com/genefriendway/onchain-handler/pkg/blockchain"
	"github.com/genefriendway/onchain-handler/pkg/logger"
	"github.com/genefriendway/onchain-handler/pkg/payment"
	"github.com/genefriendway/onchain-handler/pkg/utils"
	"github.com/genefriendway/onchain-handler/wire/providers"
)

type paymentWalletUCase struct {
	db                             *gorm.DB
	paymentWalletRepository        repotypes.PaymentWalletRepository
	paymentWalletBalanceRepository repotypes.PaymentWalletBalanceRepository
}

func NewPaymentWalletUCase(
	db *gorm.DB,
	paymentWalletRepository repotypes.PaymentWalletRepository,
	paymentWalletBalanceRepository repotypes.PaymentWalletBalanceRepository,
) ucasetypes.PaymentWalletUCase {
	return &paymentWalletUCase{
		db:                             db,
		paymentWalletRepository:        paymentWalletRepository,
		paymentWalletBalanceRepository: paymentWalletBalanceRepository,
	}
}

func (u *paymentWalletUCase) CreateAndGenerateWallet(ctx context.Context, mnemonic, passphrase, salt string, inUse bool) error {
	return u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		wallet, err := u.paymentWalletRepository.CreateNewWallet(tx, inUse)
		if err != nil {
			return fmt.Errorf("failed to create and generate wallet: %w", err)
		}
		logger.GetLogger().Debugf("Created wallet ID: %d, Address: %s", wallet.ID, wallet.Address)
		return nil
	})
}

func (u *paymentWalletUCase) IsRowExist(ctx context.Context) (bool, error) {
	return u.paymentWalletRepository.IsRowExist(ctx)
}

func (u *paymentWalletUCase) GetPaymentWalletByAddress(
	ctx context.Context, address string,
) (dto.PaymentWalletBalanceDTO, error) {
	//	Fetch wallet with balances filtered by address
	wallet, err := u.paymentWalletRepository.GetPaymentWalletWithBalancesByAddress(ctx, &address)
	if err != nil {
		return dto.PaymentWalletBalanceDTO{}, err
	}

	// Prepare network balances map
	networkBalances := make(map[string][]dto.TokenBalanceDTO)

	// Group balances by network
	for _, balance := range wallet.PaymentWalletBalances {
		networkBalances[balance.Network] = append(networkBalances[balance.Network], dto.TokenBalanceDTO{
			Symbol: balance.Symbol,
			Amount: balance.Balance,
		})
	}

	// Convert grouped balances to DTO format
	var networkBalanceDTOs []dto.NetworkBalanceDTO
	for network, tokenBalances := range networkBalances {
		networkBalanceDTOs = append(networkBalanceDTOs, dto.NetworkBalanceDTO{
			Network:       network,
			TokenBalances: tokenBalances,
		})
	}

	// Ensure `networkBalanceDTOs` is **not nil** (return `[]` instead of `null`)
	if networkBalanceDTOs == nil {
		networkBalanceDTOs = []dto.NetworkBalanceDTO{}
	}

	// Build and return the final DTO
	return dto.PaymentWalletBalanceDTO{
		ID:              wallet.ID,
		Address:         wallet.Address,
		NetworkBalances: networkBalanceDTOs,
	}, nil
}

func (u *paymentWalletUCase) AddPaymentWalletBalance(
	ctx context.Context,
	walletID uint64,
	newBalance string,
	network constants.NetworkType,
	symbol string,
) error {
	return u.paymentWalletBalanceRepository.AddPaymentWalletBalance(ctx, walletID, newBalance, network.String(), symbol)
}

func (u *paymentWalletUCase) SubtractPaymentWalletBalance(
	ctx context.Context,
	walletID uint64,
	amountToSubtract string,
	network constants.NetworkType,
	symbol string,
) error {
	return u.paymentWalletBalanceRepository.SubtractPaymentWalletBalance(ctx, walletID, amountToSubtract, network.String(), symbol)
}

func (u *paymentWalletUCase) GetPaymentWallets(ctx context.Context) ([]dto.PaymentWalletDTO, error) {
	wallets, err := u.paymentWalletRepository.GetPaymentWallets(ctx)
	if err != nil {
		return nil, err
	}
	var dtos []dto.PaymentWalletDTO
	for _, wallet := range wallets {
		dtos = append(dtos, wallet.ToDto())
	}
	return dtos, nil
}

func (u *paymentWalletUCase) GetPaymentWalletsWithBalances(ctx context.Context, network *constants.NetworkType) ([]dto.PaymentWalletBalanceDTO, error) {
	// Convert `network` to `*string`
	var parsedNetwork *string
	if network != nil {
		networkStr := network.String()
		parsedNetwork = &networkStr
	}

	// Fetch wallets & total balance per network
	wallets, err := u.paymentWalletRepository.GetPaymentWalletsWithBalances(ctx, 0, 0, parsedNetwork)
	if err != nil {
		return nil, err
	}

	// Prepare result DTOs
	var dtos []dto.PaymentWalletBalanceDTO

	for _, wallet := range wallets {
		networkBalances := make(map[string][]dto.TokenBalanceDTO) // Map to group balances by network

		// Group balances by network
		for _, balance := range wallet.PaymentWalletBalances {
			networkBalances[balance.Network] = append(networkBalances[balance.Network], dto.TokenBalanceDTO{
				Symbol: balance.Symbol,
				Amount: balance.Balance,
			})
		}

		// Convert grouped balances to DTOs
		var networkBalanceDTOs []dto.NetworkBalanceDTO
		for network, tokenBalances := range networkBalances {
			networkBalanceDTOs = append(networkBalanceDTOs, dto.NetworkBalanceDTO{
				Network:       network,
				TokenBalances: tokenBalances,
			})
		}

		// Build the PaymentWalletBalanceDTO
		dto := dto.PaymentWalletBalanceDTO{
			ID:              wallet.ID,
			Address:         wallet.Address,
			NetworkBalances: networkBalanceDTOs, // Include all network balances
		}
		dtos = append(dtos, dto)
	}

	return dtos, nil
}

func (u *paymentWalletUCase) GetPaymentWalletsWithBalancesPagination(
	ctx context.Context, page, size int, network *constants.NetworkType,
) (dto.PaginationDTOResponse, error) {
	// Setup pagination variables
	limit := size + 1 // Fetch one extra record to determine if there's a next page
	offset := (page - 1) * size

	// Convert `network` to `*string`
	var parsedNetwork *string
	if network != nil {
		networkStr := network.String()
		parsedNetwork = &networkStr
	}

	// Fetch total balance per network
	totalBalancePerNetwork, err := u.paymentWalletRepository.GetTotalBalancePerNetwork(ctx, parsedNetwork)
	if err != nil {
		return dto.PaginationDTOResponse{}, err
	}

	// Fetch wallets per network
	wallets, err := u.paymentWalletRepository.GetPaymentWalletsWithBalances(ctx, limit, offset, parsedNetwork)
	if err != nil {
		return dto.PaginationDTOResponse{}, err
	}

	// Prepare the result DTOs
	var dtos []any

	// Iterate through wallets and map them to DTOs
	for i, wallet := range wallets {
		if i >= size { // Stop at requested size to prevent including the extra record
			break
		}

		networkBalances := make(map[string][]dto.TokenBalanceDTO) // Map to group balances by network

		// Group balances by network
		for _, balance := range wallet.PaymentWalletBalances {
			networkBalances[balance.Network] = append(networkBalances[balance.Network], dto.TokenBalanceDTO{
				Symbol: balance.Symbol,
				Amount: balance.Balance,
			})
		}

		// Convert grouped balances to DTOs
		var networkBalanceDTOs []dto.NetworkBalanceDTO
		for network, tokenBalances := range networkBalances {
			networkBalanceDTOs = append(networkBalanceDTOs, dto.NetworkBalanceDTO{
				Network:       network,
				TokenBalances: tokenBalances,
			})
		}

		// Build the PaymentWalletBalanceDTO
		dto := dto.PaymentWalletBalanceDTO{
			ID:              wallet.ID,
			Address:         wallet.Address,
			NetworkBalances: networkBalanceDTOs,
		}
		dtos = append(dtos, dto)
	}

	// Determine if there's a next page
	nextPage := page
	if len(wallets) > size {
		nextPage += 1
	}

	// Return paginated response
	return dto.PaginationDTOResponse{
		NextPage:               nextPage,
		Page:                   page,
		Size:                   size,
		TotalBalancePerNetwork: totalBalancePerNetwork,
		Data:                   dtos,
	}, nil
}

func (u *paymentWalletUCase) GetReceivingWalletAddressWithBalances(
	ctx context.Context, mnemonic, passphrase, salt string,
) (string, map[constants.NetworkType]string, error) {
	// Get the receiving wallet address
	account, _, err := payment.GetReceivingWallet(mnemonic, passphrase, salt)
	if err != nil {
		return "", nil, err
	}
	walletAddress := account.Address.Hex()

	// Define supported networks
	networks := conf.GetNetworks()

	// Map to store balances per network
	balances := make(map[constants.NetworkType]string)

	// Fetch balance for each network
	for _, network := range networks {
		balance, err := u.getNativeBalanceOnchain(ctx, walletAddress, network)
		if err != nil {
			// Log the error but don't return it (continue with other networks)
			logger.GetLogger().Errorf("Failed to fetch balance for %s: %v", network, err)
			balances[network] = "error"
		} else {
			balances[network] = balance
		}
	}

	return walletAddress, balances, nil
}

func (u *paymentWalletUCase) SyncWalletBalance(ctx context.Context, walletAddress string, network constants.NetworkType) (string, error) {
	// Check if the wallet exists and get its ID
	walletID, err := u.paymentWalletRepository.GetWalletIDByAddress(ctx, walletAddress)
	if err != nil {
		return "", fmt.Errorf("failed to get wallet ID by address: %w", err)
	}

	// Fetch the USDT balance from the blockchain
	usdtAmount, err := u.getUSDTBalanceOnchain(ctx, walletAddress, network)
	if err != nil {
		return "", fmt.Errorf("failed to fetch on-chain balance: %w", err)
	}

	// Update the wallet balance
	err = u.paymentWalletBalanceRepository.UpsertPaymentWalletBalance(ctx, walletID, usdtAmount, network.String(), constants.USDT)
	if err != nil {
		return "", fmt.Errorf("failed to upsert wallet balance: %w", err)
	}

	// Return the balance as a float string
	return usdtAmount, nil
}

// getBalanceOnchain fetches and converts the on-chain USDT balance for a given wallet
func (u *paymentWalletUCase) getUSDTBalanceOnchain(ctx context.Context, walletAddress string, network constants.NetworkType) (string, error) {
	// Get USDT contract address by network
	usdtContractAddress, err := conf.GetTokenAddress(constants.USDT, network.String())
	if err != nil {
		return "", fmt.Errorf("failed to get USDT contract address: %w", err)
	}

	// Get RPC URLs and Ethereum Client based on network
	rpcUrls, err := conf.GetRpcUrls(network)
	if err != nil {
		return "", fmt.Errorf("failed to get RPC URLs: %w", err)
	}

	ethClient, err := providers.ProvideEthClient(network, rpcUrls)
	if err != nil {
		return "", fmt.Errorf("failed to initialize Ethereum client: %w", err)
	}

	// Fetch the USDT balance from the blockchain
	usdtBalance, err := ethClient.GetTokenBalance(ctx, usdtContractAddress, walletAddress)
	if err != nil {
		return "", fmt.Errorf("failed to fetch USDT balance: %w", err)
	}

	// Get token decimals from cache
	decimals, err := blockchain.GetTokenDecimalsFromCache(usdtContractAddress, network.String(), providers.ProvideCacheRepository(ctx))
	if err != nil {
		return "", fmt.Errorf("failed to get token decimals: %w", err)
	}

	// Convert balance from smallest unit
	usdtAmount, err := utils.ConvertSmallestUnitToFloatToken(usdtBalance.String(), decimals)
	if err != nil {
		return "", fmt.Errorf("failed to convert USDT balance to float: %w", err)
	}

	return usdtAmount, nil
}

// getNativeBalanceOnchain fetches and converts the on-chain native token balance for a given wallet
func (u *paymentWalletUCase) getNativeBalanceOnchain(ctx context.Context, walletAddress string, network constants.NetworkType) (string, error) {
	// Get RPC URLs and Ethereum Client based on network
	rpcUrls, err := conf.GetRpcUrls(network)
	if err != nil {
		return "", fmt.Errorf("failed to get RPC URLs: %w", err)
	}

	ethClient, err := providers.ProvideEthClient(network, rpcUrls)
	if err != nil {
		return "", fmt.Errorf("failed to initialize Ethereum client: %w", err)
	}

	// Fetch native token balance (e.g., ETH, BNB, AVAX) from the blockchain
	nativeBalance, err := ethClient.GetNativeTokenBalance(ctx, walletAddress)
	if err != nil {
		return "", fmt.Errorf("failed to fetch native token balance: %w", err)
	}

	// Convert balance from smallest unit (WEI, GWEI, etc.)
	nativeAmount, err := utils.ConvertSmallestUnitToFloatToken(nativeBalance.String(), constants.NativeTokenDecimalPlaces)
	if err != nil {
		return "", fmt.Errorf("failed to convert native token balance to float: %w", err)
	}

	return nativeAmount, nil
}
