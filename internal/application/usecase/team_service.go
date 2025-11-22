package usecase

import "github.com/ELMOCPC/Backend/internal/domain/entity"

type TeamService interface {
	// Create and manage teams
	CreateTeam(userID uint, name, description string) (*entity.Team, error)
	GetTeamByID(teamID uint) (*entity.Team, error)
	GetUserTeam(userID uint) (*entity.Team, error)
	UpdateTeam(teamID, userID uint, name, description string) (*entity.Team, error)
	DeleteTeam(teamID, userID uint) error

	// Team members management
	InviteUserByEmail(teamID, inviterID uint, email string) (*entity.TeamInvitation, error)
	AcceptInvitation(invitationID, userID uint) error
	RejectInvitation(invitationID, userID uint) error
	RemoveMember(teamID, userID, memberToRemoveID uint) error
	LeaveTeam(teamID, userID uint) error
	GetTeamMembers(teamID uint) ([]*entity.User, error)

	// Team invitations and submission
	SubmitTeam(teamID, userID uint) (*entity.Team, error)
	GetUserInvitations(email string) ([]*entity.TeamInvitation, error)
	GetPendingInvitationTokens(teamID, userID uint) ([]*entity.TeamInvitationToken, error)
	CancelInvitation(invitationID, userID uint) error
	CancelInvitationToken(teamID uint, token string, userID uint) error

	// Team validation (admin operations)
	ValidateTeam(teamID uint) (*entity.Team, error)
	RejectTeam(teamID uint) (*entity.Team, error)
	GetPendingTeams() ([]*entity.Team, error)
	GetAllTeams() ([]*entity.Team, error)
	GetTeamsByStatus(status entity.TeamStatus) ([]*entity.Team, error)
	CanModifyTeam(teamID, userID uint) (bool, error)
}
