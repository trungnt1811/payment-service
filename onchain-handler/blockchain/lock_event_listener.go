package blockchain

import (
	"context"
	"fmt"
	"math/big"
	"path/filepath"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/genefriendway/onchain-handler/internal/interfaces"
	"github.com/genefriendway/onchain-handler/internal/model"
	"github.com/genefriendway/onchain-handler/internal/utils/log"
)

// Constants for the event types
const (
	DepositEvent  = "Deposit"
	WithdrawEvent = "Withdraw"
)

// LockEventData represents the event data for both Deposit and Withdraw events.
type LockEventData struct {
	User      common.Address
	LockID    uint64
	Amount    *big.Int
	TxHash    string
	Event     string // "Deposit" or "Withdraw"
	Timestamp *big.Int
	Duration  *big.Int
}

// LockEventListener listens for Deposit and Withdraw events in the TokenLock contract.
type LockEventListener struct {
	BaseEventListener *BaseEventListener
	Repo              interfaces.LockRepository
	ContractAddress   string
	ParsedABI         abi.ABI
}

// NewLockEventListener initializes the lock event listener.
func NewLockEventListener(
	baseEventListener *BaseEventListener,
	client *ethclient.Client,
	contractAddr string,
	repo interfaces.LockRepository,
) (*LockEventListener, error) {
	abiFilePath, err := filepath.Abs("./contracts/abis/TokenLock.abi.json")
	if err != nil {
		return nil, fmt.Errorf("failed to get ABI file path: %w", err)
	}

	parsedABI, err := loadABI(abiFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load ABI: %w", err)
	}

	return &LockEventListener{
		BaseEventListener: baseEventListener,
		Repo:              repo,
		ContractAddress:   contractAddr,
		ParsedABI:         parsedABI,
	}, nil
}

// parseAndProcessLockEvent handles Deposit and Withdraw event-specific logic.
func (listener *LockEventListener) parseAndProcessLockEvent(vLog types.Log) (interface{}, error) {
	event := struct {
		User           common.Address
		LockID         *big.Int
		Amount         *big.Int
		CurrentBalance *big.Int
		Timestamp      *big.Int
		LockDuration   *big.Int // Add duration for Deposit events
	}{}

	var eventName string
	var endDuration time.Time

	if listener.isDepositEvent(vLog) {
		err := listener.ParsedABI.UnpackIntoInterface(&event, DepositEvent, vLog.Data)
		eventName = DepositEvent
		if err != nil {
			return nil, fmt.Errorf("failed to unpack Deposit event: %w", err)
		}

		// Convert lock Timestamp to time.Time
		lockTimestamp := time.Unix(event.Timestamp.Int64(), 0)

		// Convert LockDuration to a time.Duration
		lockDuration := time.Duration(event.LockDuration.Int64()) * time.Second

		// Calculate the end duration as lock Timestamp + LockDuration
		endDuration = lockTimestamp.Add(lockDuration)
	} else if listener.isWithdrawEvent(vLog) {
		err := listener.ParsedABI.UnpackIntoInterface(&event, WithdrawEvent, vLog.Data)
		eventName = WithdrawEvent
		if err != nil {
			return nil, fmt.Errorf("failed to unpack Withdraw event: %w", err)
		}
	} else {
		// Log unknown event
		log.LG.Warnf("Unknown event in log: %v", vLog)
		return "Unknown lock event", nil
	}

	event.User = common.HexToAddress(vLog.Topics[1].Hex())
	lockID, err := parseHexToUint64(vLog.Topics[2].Hex())
	if err != nil {
		return nil, fmt.Errorf("failed to parse lock ID: %w", err)
	}

	// Prepare the event model
	eventModel := model.LockEvent{
		UserAddress:     event.User.Hex(),
		LockID:          lockID,
		TransactionHash: vLog.TxHash.Hex(),
		Amount:          event.Amount.String(),
		LockAction:      strings.ToUpper(eventName),
		Status:          1, // Assume successful processing
		LockTimestamp:   event.Timestamp.Uint64(),
	}

	// Additional fields for Deposit events
	if eventName == DepositEvent {
		eventModel.CurrentBalance = event.CurrentBalance.String()
		eventModel.LockDuration = event.LockDuration.Uint64()
		eventModel.EndDuration = endDuration
	}

	// Additional fields for Withdraw events
	if eventName == WithdrawEvent {
		eventModel.CurrentBalance = event.CurrentBalance.String()
	}

	// Store event in the repository
	err = listener.Repo.CreateLockEventHistory(context.Background(), eventModel)
	if err != nil {
		if isDuplicateTransactionError(err) {
			log.LG.Warnf("Duplicate transaction detected for TxHash %s: %v", vLog.TxHash.Hex(), err)
		} else {
			log.LG.Errorf("Failed to create lock event history for LockID %d: %v", lockID, err)
			return nil, err
		}
	}

	eventData := &LockEventData{
		User:      event.User,
		LockID:    lockID,
		Amount:    event.Amount,
		TxHash:    vLog.TxHash.Hex(),
		Event:     eventName,
		Timestamp: event.Timestamp,
	}

	if eventData.Event == DepositEvent {
		eventData.Duration = event.LockDuration
	}

	return eventData, nil
}

func (listener *LockEventListener) isDepositEvent(vLog types.Log) bool {
	// Check if this is a Deposit event by comparing the topic with the Deposit event signature
	return vLog.Topics[0].Hex() == listener.ParsedABI.Events[DepositEvent].ID.Hex()
}

func (listener *LockEventListener) isWithdrawEvent(vLog types.Log) bool {
	// Check if this is a Withdraw event by comparing the topic with the Withdraw event signature
	return vLog.Topics[0].Hex() == listener.ParsedABI.Events[WithdrawEvent].ID.Hex()
}

func (listener *LockEventListener) RegisterLockEventListener(ctx context.Context) {
	listener.BaseEventListener.registerEventListener(
		listener.ContractAddress,
		listener.parseAndProcessLockEvent,
	)
}
