# onchain-handler
## Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
## Generate Go bindings to interact with smart contracts
1. mkdir -p ./contracts/abigen/lifepointtoken && touch ./contracts/abigen/lifepointtoken/LifePointToken.go
2. abigen --abi=./contracts/abis/LifePointToken.abi.json --pkg=lifepointtoken --out=./contracts/abigen/lifepointtoken/LifePointToken.go
## Swagger
http://localhost:8080/swagger/index.html
## How to run
1. make build
2. make run
