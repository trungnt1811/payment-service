## Setup development tools

### Install golangci-lint
```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### Generate Go bindings to interact with smart contracts
```bash
mkdir -p ./contracts/abigen/erc20token && touch ./contracts/abigen/erc20token/ERC20Token.go

abigen --abi=./contracts/abis/ERC20Token.abi.json --pkg=erc20token --out=./contracts/abigen/erc20token/ERC20Token.go
```

### Swagger
- `make swagger`
- http://localhost:8080/swagger/index.html

### How to run migration
- `make migrate`

### How to run linter
- `make lint`

### How to run service
- `make build`
- `make run`

### How to stop service
- `make stop`

### How to check unittest coverage
Move to the target folder and run the command:

```bash
go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out
```

### How to generate mock for unittest
1. Install `mockgen`:
```bash
go install go.uber.org/mock/mockgen@latest
```
2. Generate mock files:
```bash
mockgen -source=[path_inteface_file] -destination=[path_mock_file] -package=mocks
```  





