package lock

import (
	"context"

	"github.com/genefriendway/onchain-handler/internal/dto"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
)

type lockUCase struct {
	LockRepository interfaces.LockRepository
}

func NewLockUCase(lockRepository interfaces.LockRepository) interfaces.LockUCase {
	return &lockUCase{
		LockRepository: lockRepository,
	}
}

func (s *lockUCase) GetLatestLockEventsByUserAddress(ctx context.Context, userAddress string, page, size int) (dto.LockEventDTOResponse, error) {
	// Fetch the latest lock events
	listLatestLockEvents, err := s.LockRepository.GetLatestLockEventsByUserAddress(ctx, userAddress, page, size)
	if err != nil {
		return dto.LockEventDTOResponse{}, err
	}

	var listLockEventsDTO []dto.LockEventDTO
	var lockIDs []uint64

	// Convert to DTO and gather lock IDs
	for i := range listLatestLockEvents {
		if i >= size {
			break
		}
		listLockEventsDTO = append(listLockEventsDTO, listLatestLockEvents[i].ToDto())
		lockIDs = append(lockIDs, listLatestLockEvents[i].LockID)
	}

	// Fetch deposit lock events for the gathered lock IDs
	listLockDepositEvents, err := s.LockRepository.GetDepositLockEventByLockIDs(ctx, lockIDs)
	if err != nil {
		return dto.LockEventDTOResponse{}, err
	}

	// Map deposit amounts by lock ID
	depositMap := make(map[uint64]string)
	for _, depositEvent := range listLockDepositEvents {
		depositMap[depositEvent.LockID] = depositEvent.Amount
	}

	// Assign deposit amounts to corresponding lock events in the DTO
	for i := range listLockEventsDTO {
		if amount, found := depositMap[listLockEventsDTO[i].LockID]; found {
			listLockEventsDTO[i].DepositAmount = amount
		}
	}

	// Determine the next page if there are more events
	nextPage := page
	if len(listLatestLockEvents) > size {
		nextPage += 1
	}

	// Return the response DTO
	return dto.LockEventDTOResponse{
		NextPage: nextPage,
		Page:     page,
		Size:     size,
		Data:     listLockEventsDTO,
	}, nil
}
