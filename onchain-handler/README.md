# onchain-handler
## Install golangci-lint
```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```
## Generate Go bindings to interact with smart contracts
```bash
mkdir -p ./contracts/abigen/erc20token && touch ./contracts/abigen/erc20token/ERC20Token.go

abigen --abi=./contracts/abis/ERC20Token.abi.json --pkg=erc20token --out=./contracts/abigen/erc20token/ERC20Token.go
```
## Swagger
http://localhost:8080/swagger/index.html
## How to run
1. make build
2. make run

## How to check unittest coverage
1. move to target folder
2. go test -coverprofile=coverage.out ./...