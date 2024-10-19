# onchain-handler
## Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
## Generate Go bindings to interact with smart contracts
1. mkdir -p ./contracts/abigen/erc20token && touch ./contracts/abigen/erc20token/ERC20Token.go
2. abigen --abi=./contracts/abis/LifePointToken.abi.json --pkg=erc20token --out=./contracts/abigen/erc20token/ERC20Token.go
## Swagger
http://localhost:8080/swagger/index.html
## How to run
1. make build
2. make run
