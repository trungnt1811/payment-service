package ucases

import (
	"context"

	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/internal/domain"
	"github.com/genefriendway/onchain-handler/internal/dto"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
)

type paymentWalletUCase struct {
	paymentWalletRepository        interfaces.PaymentWalletRepository
	paymentWalletBalanceRepository interfaces.PaymentWalletBalanceRepository
}

func NewPaymentWalletUCase(
	paymentWalletRepository interfaces.PaymentWalletRepository,
	paymentWalletBalanceRepository interfaces.PaymentWalletBalanceRepository,
) interfaces.PaymentWalletUCase {
	return &paymentWalletUCase{
		paymentWalletRepository:        paymentWalletRepository,
		paymentWalletBalanceRepository: paymentWalletBalanceRepository,
	}
}

func (u *paymentWalletUCase) CreatePaymentWallets(ctx context.Context, payloads []dto.PaymentWalletPayloadDTO) error {
	var wallets []domain.PaymentWallet
	for _, payload := range payloads {
		wallet := domain.PaymentWallet{
			ID:      payload.ID,
			Address: payload.Address,
			InUse:   payload.InUse,
		}
		wallets = append(wallets, wallet)
	}
	return u.paymentWalletRepository.CreatePaymentWallets(ctx, wallets)
}

func (u *paymentWalletUCase) IsRowExist(ctx context.Context) (bool, error) {
	return u.paymentWalletRepository.IsRowExist(ctx)
}

func (u *paymentWalletUCase) GetPaymentWalletByAddress(ctx context.Context, address string) (dto.PaymentWalletDTO, error) {
	wallet, err := u.paymentWalletRepository.GetPaymentWalletByAddress(ctx, address)
	if err != nil {
		return dto.PaymentWalletDTO{}, err
	}
	return wallet.ToDto(), nil
}

func (u *paymentWalletUCase) UpsertPaymentWalletBalances(
	ctx context.Context,
	walletIDs []uint64,
	newBalances []string,
	network constants.NetworkType,
	symbol string,
) error {
	return u.paymentWalletBalanceRepository.UpsertPaymentWalletBalances(ctx, walletIDs, newBalances, string(network), symbol)
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

func (u *paymentWalletUCase) GetPaymentWalletsWithBalances(ctx context.Context, nonZeroOnly bool, network *constants.NetworkType) ([]dto.PaymentWalletBalanceDTO, error) {
	// Convert `network` to `*string` if it's not nil
	var parsedNetwork *string
	if network != nil {
		networkStr := string(*network)
		parsedNetwork = &networkStr
	} else {
		parsedNetwork = nil
	}

	// Fetch wallets with balances from the repository
	wallets, err := u.paymentWalletRepository.GetPaymentWalletsWithBalances(ctx, nonZeroOnly, parsedNetwork)
	if err != nil {
		return nil, err
	}

	// Prepare the result DTOs
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
