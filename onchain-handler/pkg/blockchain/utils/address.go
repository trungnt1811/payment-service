package utils

import "github.com/ethereum/go-ethereum/common"

// ConvertToCommonAddresses converts string addresses to common.Address type
func ConvertToCommonAddresses(recipients []string) []common.Address {
	var addresses []common.Address
	for _, recipient := range recipients {
		addresses = append(addresses, common.HexToAddress(recipient))
	}
	return addresses
}

// IsValidEthAddress checks if a given string is a valid Ethereum address
func IsValidEthAddress(address string) bool {
	return common.IsHexAddress(address)
}
