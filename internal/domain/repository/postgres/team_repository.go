package postgres

import (
	"github.com/ELMOCPC/Backend/internal/domain/entity"
	"github.com/ELMOCPC/Backend/internal/infrastructure/database"
)

type TeamRepository interface {
	// Team CRUD operations
	CreateTeam(db database.Database, team *entity.Team) error
	FindTeamByID(db database.Database, id uint) (*entity.Team, error)
	FindTeamByUserID(db database.Database, userID uint) (*entity.Team, error)
	UpdateTeam(db database.Database, team *entity.Team) error
	DeleteTeam(db database.Database, id uint) error

	// Team members operations
	AddTeamMember(db database.Database, teamID, userID uint) error
	RemoveTeamMember(db database.Database, teamID, userID uint) error
	GetTeamMembers(db database.Database, teamID uint) ([]*entity.User, error)
	GetTeamMemberCount(db database.Database, teamID uint) (int64, error)
	IsTeamMember(db database.Database, teamID, userID uint) (bool, error)

	// Team invitations operations
	CreateInvitation(db database.Database, invitation *entity.TeamInvitation) error
	FindInvitationByID(db database.Database, id uint) (*entity.TeamInvitation, error)
	FindInvitationByTeamAndEmail(db database.Database, teamID uint, email string) (*entity.TeamInvitation, error)
	FindInvitationsByTeamID(db database.Database, teamID uint) ([]*entity.TeamInvitation, error)
	FindInvitationsByEmail(db database.Database, email string) ([]*entity.TeamInvitation, error)
	FindPendingInvitationsByEmail(db database.Database, email string) ([]*entity.TeamInvitation, error)
	UpdateInvitation(db database.Database, invitation *entity.TeamInvitation) error
	DeleteInvitation(db database.Database, id uint) error

	// Find teams with filters
	FindTeams(db database.Database, options *QueryOptions) ([]*entity.Team, int64, error)
	FindTeamsByStatus(db database.Database, status entity.TeamStatus) ([]*entity.Team, error)
}
