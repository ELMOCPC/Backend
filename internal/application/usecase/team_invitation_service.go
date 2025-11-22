package usecase

import "github.com/ELMOCPC/Backend/internal/domain/entity"

type InviteMemberData struct {
	FirstName     string
	LastName      string
	Email         string
	Phone         string
	TshirtSize    string
	University    string
	StudentNumber string
	NationalCode  string
}

type TeamInvitationService interface {
	SendInvitations(teamID, invitedByID uint, members []InviteMemberData) ([]string, error)
	AcceptInvitation(token string) (*entity.User, string, string, error)
	RejectInvitation(token string) error
	GetPendingInvitations(email string) ([]*entity.TeamInvitationToken, error)
}
