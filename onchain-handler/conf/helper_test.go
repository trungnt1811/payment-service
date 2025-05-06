package conf

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/genefriendway/onchain-handler/constants"
)

var (
	avaxUSDTAddressMock = "avax-usdt-address"
	avaxUSDCAddressMock = "avax-usdc-address"
	bscUSDTAddressMock  = "bsc-usdt-address"
	bscUSDCAddressMock  = "bsc-usdc-address"
)

func TestGetTokenSymbol(t *testing.T) {
	// Setup mock configuration
	configuration.Blockchain.AvaxNetwork.AvaxUSDTContractAddress = avaxUSDTAddressMock
	configuration.Blockchain.BscNetwork.BscUSDTContractAddress = bscUSDTAddressMock
	configuration.Blockchain.AvaxNetwork.AvaxUSDCContractAddress = avaxUSDCAddressMock
	configuration.Blockchain.BscNetwork.BscUSDCContractAddress = bscUSDCAddressMock

	tests := []struct {
		name           string
		tokenAddress   string
		expectedSymbol string
		expectingError bool
		expectedErrMsg string
	}{
		{
			name:           "Valid AVAX USDT address",
			tokenAddress:   avaxUSDTAddressMock,
			expectedSymbol: constants.USDT,
		},
		{
			name:           "Valid BSC USDT address",
			tokenAddress:   bscUSDTAddressMock,
			expectedSymbol: constants.USDT,
		},
		{
			name:           "Valid AVAX USDC address",
			tokenAddress:   avaxUSDCAddressMock,
			expectedSymbol: constants.USDC,
		},
		{
			name:           "Valid BSC USDC address",
			tokenAddress:   bscUSDCAddressMock,
			expectedSymbol: constants.USDC,
		},
		{
			name:           "Unknown token address",
			tokenAddress:   "unknown-address",
			expectingError: true,
			expectedErrMsg: "unknown token address: unknown-address",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			symbol, err := GetTokenSymbol(tt.tokenAddress)

			if tt.expectingError {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectedErrMsg)
				assert.Empty(t, symbol)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedSymbol, symbol)
			}
		})
	}
}

func TestGetTokenAddress(t *testing.T) {
	// Setup mock configuration
	configuration.Blockchain.AvaxNetwork.AvaxUSDTContractAddress = avaxUSDTAddressMock
	configuration.Blockchain.AvaxNetwork.AvaxUSDCContractAddress = avaxUSDCAddressMock
	configuration.Blockchain.BscNetwork.BscUSDTContractAddress = bscUSDTAddressMock
	configuration.Blockchain.BscNetwork.BscUSDCContractAddress = bscUSDCAddressMock

	tests := []struct {
		name           string
		symbol         string
		network        string
		expectedAddr   string
		expectingError bool
		expectedErrMsg string
	}{
		{
			name:         "Valid AVAX USDT address",
			symbol:       constants.USDT,
			network:      constants.AvaxCChain.String(),
			expectedAddr: avaxUSDTAddressMock,
		},
		{
			name:         "Valid AVAX USDC address",
			symbol:       constants.USDC,
			network:      constants.AvaxCChain.String(),
			expectedAddr: avaxUSDCAddressMock,
		},
		{
			name:         "Valid BSC USDT address",
			symbol:       constants.USDT,
			network:      constants.Bsc.String(),
			expectedAddr: bscUSDTAddressMock,
		},
		{
			name:         "Valid BSC USDC address",
			symbol:       constants.USDC,
			network:      constants.Bsc.String(),
			expectedAddr: bscUSDCAddressMock,
		},
		{
			name:           "Unknown token symbol for valid network",
			symbol:         "DAI",
			network:        constants.Bsc.String(),
			expectingError: true,
			expectedErrMsg: fmt.Sprintf("unknown token symbol for network %s: %s", constants.Bsc.String(), "DAI"),
		},
		{
			name:           "Unsupported network",
			symbol:         constants.USDT,
			network:        "POLYGON",
			expectingError: true,
			expectedErrMsg: "unsupported network: POLYGON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, err := GetTokenAddress(tt.symbol, tt.network)

			if tt.expectingError {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectedErrMsg)
				assert.Empty(t, addr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedAddr, addr)
			}
		})
	}
}
