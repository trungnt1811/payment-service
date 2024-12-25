# Onchain Handler
Onchain Handler is a comprehensive service designed to facilitate seamless interaction with blockchain networks such as Binance Smart Chain (BSC) and Avalanche (AVAX). This service supports various functionalities, including managing payment wallets, processing token transfers, handling USDT balances, and tracking blockchain events efficiently. It is tailored to provide robust support for on-chain operations with scalability and reliability in mind.

## Features

- **Blockchain Agnostic**: Supports multiple blockchain networks like BSC and AVAX.

- **Payment Wallet Management**: Secure handling of payment wallets with balances and transaction tracking.

- **Token Transfers**: Efficient token transfer mechanisms, including USDT and native tokens.

- **Block Listeners**: Real-time listening to blockchain events from a specific starting block.

- **HD Wallet Support**: Wallets are securely generated using a single mnemonic, passphrase, and salt, ensuring consistency and enhanced security. This approach removes the necessity of storing private keys in the database, as wallets can be deterministically regenerated when needed.

- **Error-Resilient Operations**: Built-in retry mechanisms ensure the system remains robust against network or RPC errors. In cases where services experience downtime, a catch-up worker is implemented to process missed events or transactions during the downtime period, ensuring data consistency and continuity.

- **Expired Order Handling**:
  - Orders in the EXPIRED state are monitored by a catch-up worker. If an order remains in the EXPIRED state beyond the configured cutoff time (ORDER_CUTOFF_TIME), it transitions to FAILED.
  - However, in cases where payment is received after the order has transitioned to EXPIRED but before it moves to FAILED, the order will be updated to SUCCESS if the payment amount is sufficient.
  - This mechanism ensures flexibility for scenarios such as delayed transactions or network latency, maintaining accuracy and fairness in order processing.

- **State Machine for Payment Orders**:

  - PENDING: Order is created and waiting for payment.
  - PROCESSING: Payment is detected, and the order is being processed.
  - SUCCESS: Payment is completed, and the order is fulfilled.
  - PARTIAL: Partial payment is received, but not sufficient to fulfill the order.
  - EXPIRED: Order is expired due to timeout.
  - FAILED: 
    - Orders transition to FAILED after remaining in the EXPIRED state for a period exceeding the configured cutoff time (ORDER_CUTOFF_TIME in environment variables).
    - This ensures proper handling of long-expired orders, maintaining system consistency and accuracy.

- **Payment Wallets Pool Optimization**:

  - Dynamically maintains a pool of reusable payment wallets to optimize wallet usage.
  - Payment wallets are automatically assigned to new orders and released back into the pool upon order completion or failure.
  - Reduces operational costs by minimizing wallet creation and promoting efficient wallet reuse.
  - Supports consolidation of token balances by transferring USDT from payment wallets to the receiving wallet, which then forwards the collected USDT to the master wallet.
  - Automatically distributes BNB and AVAX to payment wallets for gas fees, ensuring smooth transaction processing.
  - Integrated a configurable worker that can run either: Daily at 00:00 UTC, or Hourly, depending on the WITHDRAW_WORKER_INTERVAL configuration. The worker handles wallet-related operations, such as gas fee distribution and balance consolidation.

## Environment Variables
The following environment variables are required for the application to run. Set them in a .env file or your environment:

### General Configuration
| Variable                | Description                                                            | Default               |
|-------------------------|------------------------------------------------------------------------|-----------------------|
| `ENV`                   | Environment mode: `DEV`, `PROD`.                                       | `DEV`                |
| `LOG_LEVEL`             | Logging level: `debug`, `info`, `warn`, `error`.                       | `debug`                |
| `APP_NAME`              | Application name.                                                      | `onchain-handler`     |
| `APP_PORT`              | Port to run the application.                                           | `8080`                |

### Database Configuration
| Variable                | Description                                                            | Default               |
|-------------------------|------------------------------------------------------------------------|-----------------------|
| `DB_USER`               | Database username.                                                     | `postgres`            |
| `DB_PASSWORD`           | Database password.                                                     | `123456`              |
| `DB_HOST`               | Database host.                                                         | `localhost`           |
| `DB_PORT`               | Database port.                                                         | `5432`                |
| `DB_NAME`               | Database name.                                                         | `onchain-handler`     |

### Blockchain Configuration
| Variable                     | Description                                                                                | Default                                                      |
|------------------------------|--------------------------------------------------------------------------------------------|--------------------------------------------------------------|
| `BSC_RPC_URLS`               | List of Binance Smart Chain RPC URLs.                                                      | `https://rpc.ankr.com/bsc_testnet_chapel/...` (ask developer) |
| `BSC_CHAIN_ID`               | Binance Smart Chain ID.      | `0`     (ask developer)                                     |
| `BSC_START_BLOCK_LISTENER`   | Starting block for listening on BSC. **Avoid setting it too far back to prevent pruning.**                             | `0` (ask developer)          |
| `BSC_USDT_CONTRACT_ADDRESS`  | Contract address for USDT on Binance Smart Chain.                                          | `0x...` (ask developer)              |
| `AVAX_RPC_URLS`              | List of Avalanche RPC URLs.  | `https://rpc.ankr.com/avalanche_fuji/...` (ask developer)   |
| `AVAX_CHAIN_ID`              | Avalanche Chain ID.          | `0`  (ask developer)                                        |
| `AVAX_START_BLOCK_LISTENER`  | Starting block for listening on Avalanche. **Avoid setting it too far back to prevent pruning.** | `0` (ask developer)            |
| `AVAX_USDT_CONTRACT_ADDRESS` | Contract address for USDT on Avalanche.                                       | `0x...` (ask developer)                     |
| `GAS_BUFFER_MULTIPLIER`      | Multiplier to buffer estimated gas calculations.                              | `2`                            |

### Additional Configuration
| Variable                     | Description                                                            | Default               |
|------------------------------|------------------------------------------------------------------------|-----------------------|
| `INIT_WALLET_COUNT`          | Initial count of wallets to be generated.                              | `10`                  |
| `EXPIRED_ORDER_TIME`         | Time (in minutes) for an order to move from `PEDNING` to `EXPIRED`.    | `15`                  |
| `ORDER_CUTOFF_TIME`          | Maximum duration (in minutes) for an order to move from `EXPIRED` to `FAILED`.                 | `1440`|
| `PAYMENT_COVERING`           | Discount amount applied to each order.                                 | `1` (1 USDT)          |
| `MNEMONIC`                   | Secret mnemonic phrase for HD wallet derivation.                       | `your mnemonic` (ask devops)           |
| `PASSPHRASE`                 | Passphrase for HD wallet derivation.                                   | `your passphrase` (ask devops)           |
| `SALT`                       | Salt for HD wallet derivation.                                         | `your salt` (ask devops)                |
| `MASTER_WALLET_ADDRESS`      | The address of the master wallet where funds from receiving wallets are consolidated. Ensure this is securely configured.| `your master wallet address` (ask devops) |
| `WITHDRAW_WORKER_INTERVAL`   | Interval for the paymentWalletWithdrawWorker to run. Accepts `hourly` or `daily`.                       | `hourly`                                                                |

## Receiving Wallet Documentation

### Overview

The **Receiving Wallet** is a centralized wallet designed to:
1. **Collect USDT** from all Payment Wallets.
2. Consolidate the collected **USDT** and transfer it to the **Master Wallet**.
3. **Distribute Gas Fees** (BNB for Binance Smart Chain and AVAX for Avalanche) to Payment Wallets to ensure sufficient gas for future transactions.

A worker automates these processes daily at **00:00 UTC** or optionally configurable to run hourly based on operational needs.

---

### Daily Wallet Usage and Transaction Fee Estimation

#### Daily Wallet Usage
- **Average daily reuse**: 10 Payment Wallets.
- **Each Payment Wallet** performs at least one **USDT transfer** per day, requiring gas for these transactions.

#### Transaction Fees Per Network
| **Network**      | **Gas Fee per USDT Transfer** | 
|------------------|-------------------------------|
| Binance Smart Chain (BSC) | **0.00016 BNB**      | 
| Avalanche (AVAX)          | **0.0028 AVAX**      | 

#### Monthly Estimation (30 Days)
For **10 Payment Wallets reused daily**, the estimated gas fees are:

| **Network**      | **Gas Fee per Day**             | **Monthly Gas Fee**  |
|------------------|---------------------------------|----------------------|
| Binance Smart Chain (BSC) | **10 × 0.00016 BNB = 0.0016 BNB** | **0.048 BNB** |
| Avalanche (AVAX)          | **10 × 0.0028 AVAX = 0.028 AVAX** | **0.84 AVAX** |

---

### Gas Balance Top-Up for Receiving Wallet

#### Monthly Top-Up Recommendation
| **Network**      | **Estimated Gas (Monthly)** | **Recommended Top-Up (30% Buffer)**  |
|------------------|-----------------------------|--------------------------------------|
| Binance Smart Chain (BSC) | **0.048 BNB**      | **0.0624 BNB**                       | 
| Avalanche (AVAX)          | **0.84 AVAX**      | **1.092 AVAX**                       |

**Notes**:
- The recommended top-up includes a **30% buffer** for unexpected additional transactions.
- Gas fees and token prices are subject to change. Adjust the top-up values periodically to align with current rates.

---

### Payment Wallets Withdrawing Worker

#### Worker Schedule
The worker is configured to run either:
  Daily at 00:00 UTC, or
  Hourly (based on the WITHDRAW_WORKER_INTERVAL configuration) to:
1. **Collect USDT** from all Payment Wallets into the Receiving Wallet.
2. **Transfer USDT** from the Receiving Wallet to the Master Wallet.
3. **Distribute Gas Fees** (BNB and AVAX) to Payment Wallets to ensure they are operational.

#### Workflow Summary
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
  - Top up the Receiving Wallet monthly with **0.0624 BNB** and **1.092 AVAX** for seamless operations.
- **Payment Wallets Withdrawing Worker**:
  - Runs daily to minimize manual intervention and ensure all Payment Wallets are operational with sufficient gas.
