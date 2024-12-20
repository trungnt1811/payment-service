# Onchain Handler
Onchain Handler is a comprehensive service designed to facilitate seamless interaction with blockchain networks such as Binance Smart Chain (BSC) and Avalanche (AVAX). This service supports various functionalities, including managing payment wallets, processing token transfers, handling USDT balances, and tracking blockchain events efficiently. It is tailored to provide robust support for on-chain operations with scalability and reliability in mind.

## Features

- Blockchain Agnostic: Supports multiple blockchain networks like BSC and AVAX.

- Payment Wallet Management: Secure handling of payment wallets with balances and transaction tracking.

- Token Transfers: Efficient token transfer mechanisms, including USDT and native tokens.

- Block Listeners: Real-time listening to blockchain events from a specific starting block.

- HD Wallet Support: Wallet generation using mnemonics without storing private keys in the database.

- Error-Resilient Operations: Built-in retry mechanisms for handling network or RPC errors.

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
1. `make build`
2. `make run`

## How to check unittest coverage
Move to the target folder and run the command:

```bash
go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out
```

## How to generate mock for unittest
1. Install `mockgen`:
```bash
go install go.uber.org/mock/mockgen@latest
```
2. Generate mock files:
```bash
mockgen -source=[path_inteface_file] -destination=[path_mock_file] -package=mocks
```

## Environment Variables
The following environment variables are required for the application to run. Set them in a .env file or your environment:

### General Configuration
| Variable                | Description                                                            | Example               |
|-------------------------|------------------------------------------------------------------------|-----------------------|
| `ENV`                  | Environment mode: `DEV`, `PROD`.                           | `DEV`                |
| `LOG_LEVEL`            | Logging level: `debug`, `info`, `warn`, `error`.                      | `debug`              |
| `APP_NAME`             | Application name.                                                     | `onchain-handler`    |
| `APP_PORT`             | Port to run the application.                                          | `8080`               |

### Database Configuration
| Variable                | Description                                                            | Example               |
|-------------------------|------------------------------------------------------------------------|-----------------------|
| `DB_USER`              | Database username.                                                    | `postgres`           |
| `DB_PASSWORD`          | Database password.                                                    | `123456`             |
| `DB_HOST`              | Database host.                                                        | `localhost`          |
| `DB_PORT`              | Database port.                                                        | `5432`               |
| `DB_NAME`              | Database name.                                                        | `onchain-handler`    |

### Blockchain Configuration
| Variable                     | Description                                                                                | Example                                                      |
|------------------------------|--------------------------------------------------------------------------------------------|--------------------------------------------------------------|
| `BSC_RPC_URLS`              | List of Binance Smart Chain RPC URLs.                                                     | `https://rpc.ankr.com/bsc/...`                              |
| `BSC_CHAIN_ID`              | Binance Smart Chain ID.                                                                   | `56`                                                         |
| `BSC_START_BLOCK_LISTENER`  | Starting block for listening on BSC. **Avoid setting it too far back to prevent pruning.** | `45011000`                                                   |
| `BSC_USDT_CONTRACT_ADDRESS` | Contract address for USDT on Binance Smart Chain.                                         | `0x55d398326f99059fF775485246999027B3197955`                 |
| `AVAX_RPC_URLS`             | List of Avalanche RPC URLs.                                                              | `https://rpc.ankr.com/avalanche/...`                        |
| `AVAX_CHAIN_ID`             | Avalanche Chain ID.                                                                      | `43114`                                                      |
| `AVAX_START_BLOCK_LISTENER` | Starting block for listening on Avalanche. **Avoid setting it too far back to prevent pruning.** | `54567000`                                                   |
| `AVAX_USDT_CONTRACT_ADDRESS`| Contract address for USDT on Avalanche.                                                  | `0x9702230A8Ea53601f5cD2dc00fDBc13d4dF4A8c7`                 |

### Additional Configuration
| Variable                     | Description                                                            | Example               |
|------------------------------|------------------------------------------------------------------------|-----------------------|
| `INIT_WALLET_COUNT`          | Initial count of wallets to be generated.                             | `10`                 |
| `EXPIRED_ORDER_TIME`         | Time (in minutes) for an order to move from `PEDNING` to `EXPIRED`.    | `15`                 |
| `ORDER_CUTOFF_TIME`          | Maximum duration (in minutes) for an order to move from `EXPIRED` to `FAILED`.          | `1440`               |
| `PAYMENT_COVERING`           | Discount amount applied to each order.                            | `1` (1 USDT)             |
| `MNEMONIC`                   | Secret mnemonic phrase for HD wallet derivation.                      | `net motor more...`  |
| `PASSPHRASE`                 | Passphrase for HD wallet derivation.                                  | `your passphrase`              |
| `SALT`                       | Salt for HD wallet derivation.                                        | `your salt`               |

## Notes
- **BSC_START_BLOCK_LISTENER** and **AVAX_START_BLOCK_LISTENER**: Ensure values are not too old to avoid issues with block pruning.
- **MNEMONIC, PASSPHRASE, SALT**: These are used to generate wallets through the HD path algorithm. Private keys do not need to be stored in the database.

