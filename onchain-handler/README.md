# Onchain Handler
Onchain Handler is a comprehensive service designed to facilitate seamless interaction with blockchain networks such as Binance Smart Chain (BSC) and Avalanche (AVAX). This service supports various functionalities, including managing payment wallets, processing token transfers, handling USDT balances, and tracking blockchain events efficiently. It is tailored to provide robust support for on-chain operations with scalability and reliability in mind.

## Features

- **Blockchain Agnostic**: Supports multiple blockchain networks like BSC and AVAX.

- **Payment Wallet Management**: Secure handling of payment wallets with balances and transaction tracking.

- **Token Transfers**: Efficient token transfer mechanisms, including USDT and native tokens.

- **Block Listeners**: Real-time listening to blockchain events from a specific starting block.

- **HD Wallet Support**: Wallet generation using mnemonics without storing private keys in the database.

- **Error-Resilient Operations**: Built-in retry mechanisms for handling network or RPC errors.

- **Expired Order Handling**: Worker processes expired orders and transitions them to a FAILED state based on the configured expiration time.

- **State Machine for Payment Orders**:

- - PENDING: Order is created and waiting for payment.

- - PROCESSING: Payment is detected, and the order is being processed.

- - SUCCESS: Payment is completed, and the order is fulfilled.

- - PARTIAL: Partial payment is received, but not sufficient to fulfill the order.

- - EXPIRED: Order is expired due to timeout.

- - FAILED: Order is marked as failed after expiration or manual intervention.

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
| Variable                | Description                                                            | Default               |
|-------------------------|------------------------------------------------------------------------|-----------------------|
| `ENV`                  | Environment mode: `DEV`, `PROD`.                           | `PROD`                |
| `LOG_LEVEL`            | Logging level: `debug`, `info`, `warn`, `error`.                      | `info`              |
| `APP_NAME`             | Application name.                                                     | `onchain-handler`    |
| `APP_PORT`             | Port to run the application.                                          | `8080`               |

### Database Configuration
| Variable                | Description                                                            | Default               |
|-------------------------|------------------------------------------------------------------------|-----------------------|
| `DB_USER`              | Database username.                                                    | `postgres`           |
| `DB_PASSWORD`          | Database password.                                                    | `123456`             |
| `DB_HOST`              | Database host.                                                        | `localhost`          |
| `DB_PORT`              | Database port.                                                        | `5432`               |
| `DB_NAME`              | Database name.                                                        | `onchain-handler`    |

### Blockchain Configuration
| Variable                     | Description                                                                                | Example                                                      |
|------------------------------|--------------------------------------------------------------------------------------------|--------------------------------------------------------------|
| `BSC_RPC_URLS`              | List of Binance Smart Chain RPC URLs.                                                     | `https://rpc.ankr.com/bsc/...` (ask developer)                            |
| `BSC_CHAIN_ID`              | Binance Smart Chain ID.                                                                   | `56`                                                         |
| `BSC_START_BLOCK_LISTENER`  | Starting block for listening on BSC. **Avoid setting it too far back to prevent pruning.** | `45011000` (ask developer)                                                   |
| `BSC_USDT_CONTRACT_ADDRESS` | Contract address for USDT on Binance Smart Chain.                                         | `0x55d398326f99059fF775485246999027B3197955`             |
| `AVAX_RPC_URLS`             | List of Avalanche RPC URLs.                                                              | `https://rpc.ankr.com/avalanche/...` (ask developer)                        |
| `AVAX_CHAIN_ID`             | Avalanche Chain ID.                                                                      | `43114`                                                      |
| `AVAX_START_BLOCK_LISTENER` | Starting block for listening on Avalanche. **Avoid setting it too far back to prevent pruning.** | `54567000` (ask developer)                                                 |
| `AVAX_USDT_CONTRACT_ADDRESS`| Contract address for USDT on Avalanche.                                                  | `0x9702230A8Ea53601f5cD2dc00fDBc13d4dF4A8c7`                 |

### Additional Configuration
| Variable                     | Description                                                            | Default               |
|------------------------------|------------------------------------------------------------------------|-----------------------|
| `INIT_WALLET_COUNT`          | Initial count of wallets to be generated.                             | `10`                 |
| `EXPIRED_ORDER_TIME`         | Time (in minutes) for an order to move from `PEDNING` to `EXPIRED`.    | `15`                 |
| `ORDER_CUTOFF_TIME`          | Maximum duration (in minutes) for an order to move from `EXPIRED` to `FAILED`.          | `1440`               |
| `PAYMENT_COVERING`           | Discount amount applied to each order.                            | `1` (1 USDT)             |
| `MNEMONIC`                   | Secret mnemonic phrase for HD wallet derivation.                      | `net motor more...` (ask devops) |
| `PASSPHRASE`                 | Passphrase for HD wallet derivation.                                  | `your passphrase` (ask devops)              |
| `SALT`                       | Salt for HD wallet derivation.                                        | `your salt` (ask devops)              |

# Receiving Wallet Documentation

## Overview

The **Receiving Wallet** is a centralized wallet designed to:
1. **Collect USDT** from all Payment Wallets daily.
2. Consolidate the collected **USDT** and transfer it to the **Master Wallet**.
3. **Distribute Gas Fees** (BNB for Binance Smart Chain and AVAX for Avalanche) to Payment Wallets to ensure sufficient gas for future transactions.

A worker automates these processes daily at **00:00 UTC**.

---

## Daily Wallet Usage and Transaction Fee Estimation

### Daily Wallet Usage
- **Average daily reuse**: 10 Payment Wallets.
- **Each Payment Wallet** performs at least one **USDT transfer** per day, requiring gas for these transactions.

### Transaction Fees Per Network
| **Network**     | **Gas Fee per USDT Transfer** | 
|------------------|-------------------------------|
| Binance Smart Chain (BSC) | **0.00006 BNB**       | 
| Avalanche (AVAX)          | **0.0028 AVAX**      | 

### Monthly Estimation (30 Days)
For **10 Payment Wallets reused daily**, the estimated gas fees are:

| **Network**     | **Gas Fee per Day**             | **Monthly Gas Fee** |
|------------------|---------------------------------|----------------------|
| Binance Smart Chain (BSC) | **10 × 0.00006 BNB = 0.0006 BNB** | **0.018 BNB**   |
| Avalanche (AVAX)          | **10 × 0.0028 AVAX = 0.028 AVAX**| **0.84 AVAX** |

---

## Gas Balance Top-Up for Receiving Wallet

### Monthly Top-Up Recommendation
| **Network**     | **Estimated Gas (Monthly)** | **Recommended Top-Up (30% Buffer)** |
|------------------|-----------------------------|--------------------------------------|
| Binance Smart Chain (BSC) | **0.018 BNB**         | **0.0234 BNB**                      | 
| Avalanche (AVAX)          | **0.84 AVAX**        | **1.092 AVAX**                     |

**Notes**:
- The recommended top-up includes a **30% buffer** for unexpected additional transactions.
- Gas fees and token prices are subject to change. Adjust the top-up values periodically to align with current rates.

---

## Payment Wallets Withdrawing Worker

### Worker Schedule
A worker runs daily at **00:00 UTC** to:
1. **Collect USDT** from all Payment Wallets into the Receiving Wallet.
2. **Transfer USDT** from the Receiving Wallet to the Master Wallet.
3. **Distribute Gas Fees** (BNB and AVAX) to Payment Wallets to ensure they are operational.

### Workflow Summary
1. **Gas Redistribution**:
   - The Receiving Wallet sends BNB and AVAX to Payment Wallets for gas fee provisioning.
2. **USDT Collection**:
   - The Receiving Wallet collects USDT balances from Payment Wallets.
3. **Master Wallet Transfer**:
   - The Receiving Wallet transfers the total collected USDT to the Master Wallet.

---

## Notes
- **BSC_START_BLOCK_LISTENER** and **AVAX_START_BLOCK_LISTENER**:
  - Ensure the starting block is not too far in the past to avoid issues with pruned nodes.
- **Gas Fee Recommendations**:
  - Top up the Receiving Wallet monthly with **0.078 BNB** and **0.156** for seamless operations.
- **Payment Wallets Withdrawing Worker**:
  - Runs daily to minimize manual intervention and ensure all Payment Wallets are operational with sufficient gas.





