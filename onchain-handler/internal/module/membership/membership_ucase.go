package membership

import (
	"context"

	"github.com/genefriendway/onchain-handler/internal/dto"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
)

type membershipUCase struct {
	MembershipRepository interfaces.MembershipRepository
}

func NewMembershipUCase(membershipRepository interfaces.MembershipRepository) interfaces.MembershipUCase {
	return &membershipUCase{
		MembershipRepository: membershipRepository,
	}
}

func (u *membershipUCase) GetMembershipEventByOrderID(ctx context.Context, orderID uint64) (*dto.MembershipEventDTO, error) {
	membershipEvent, err := u.MembershipRepository.GetMembershipEventByOrderID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if membershipEvent == nil {
		return nil, nil
	}
	membershipEventDTO := membershipEvent.ToDto()
	return &membershipEventDTO, nil
}

// GetMembershipEventsByOrderIDs retrieves a list of membership events by their order IDs.
func (u *membershipUCase) GetMembershipEventsByOrderIDs(ctx context.Context, orderIDs []uint64) ([]dto.MembershipEventDTO, error) {
	// Retrieve membership events from the repository using the provided order IDs.
	membershipEvents, err := u.MembershipRepository.GetMembershipEventsByOrderIDs(ctx, orderIDs)
	if err != nil {
		return nil, err
	}

	// Check if no events were found.
	if len(membershipEvents) == 0 {
		return nil, nil
	}

	// Convert the retrieved membership events to DTOs.
	var membershipEventDTOs []dto.MembershipEventDTO
	for _, event := range membershipEvents {
		membershipEventDTO := event.ToDto() // Assuming ToDto() method converts model to DTO.
		membershipEventDTOs = append(membershipEventDTOs, membershipEventDTO)
	}

	return membershipEventDTOs, nil
}
