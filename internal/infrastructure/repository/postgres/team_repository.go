package postgres

import (
	"github.com/ELMOCPC/Backend/internal/domain/entity"
	"github.com/ELMOCPC/Backend/internal/domain/repository/postgres"
	"github.com/ELMOCPC/Backend/internal/infrastructure/database"
	"gorm.io/gorm"
)

type TeamRepository struct{}

func NewTeamRepository() *TeamRepository {
	return &TeamRepository{}
}

// Team CRUD operations

func (repo *TeamRepository) CreateTeam(db database.Database, team *entity.Team) error {
	return db.GetDB().Create(team).Error
}

func (repo *TeamRepository) FindTeamByID(db database.Database, id uint) (*entity.Team, error) {
	var team entity.Team
	result := db.GetDB().Preload("Creator").Preload("Members").Preload("Invitations").First(&team, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &team, nil
}

func (repo *TeamRepository) FindTeamByUserID(db database.Database, userID uint) (*entity.Team, error) {
	var team entity.Team
	result := db.GetDB().Preload("Creator").Preload("Members").Preload("Invitations").Where("creator_id = ?", userID).First(&team)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &team, nil
}

func (repo *TeamRepository) UpdateTeam(db database.Database, team *entity.Team) error {
	return db.GetDB().Save(team).Error
}

func (repo *TeamRepository) DeleteTeam(db database.Database, id uint) error {
	return db.GetDB().Delete(&entity.Team{}, id).Error
}

// Team members operations

func (repo *TeamRepository) AddTeamMember(db database.Database, teamID, userID uint) error {
	team := &entity.Team{}
	team.ID = teamID
	user := &entity.User{}
	user.ID = userID
	return db.GetDB().Model(team).Association("Members").Append(user)
}

func (repo *TeamRepository) RemoveTeamMember(db database.Database, teamID, userID uint) error {
	team := &entity.Team{}
	team.ID = teamID
	user := &entity.User{}
	user.ID = userID
	return db.GetDB().Model(team).Association("Members").Delete(user)
}

func (repo *TeamRepository) GetTeamMembers(db database.Database, teamID uint) ([]*entity.User, error) {
	var users []*entity.User
	result := db.GetDB().Where("id IN (SELECT user_id FROM team_members WHERE team_id = ?)", teamID).Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}

func (repo *TeamRepository) GetTeamMemberCount(db database.Database, teamID uint) (int64, error) {
	var count int64
	result := db.GetDB().Table("team_members").Where("team_id = ?", teamID).Count(&count)
	if result.Error != nil {
		return 0, result.Error
	}
	return count, nil
}

func (repo *TeamRepository) IsTeamMember(db database.Database, teamID, userID uint) (bool, error) {
	var count int64
	result := db.GetDB().Table("team_members").Where("team_id = ? AND user_id = ?", teamID, userID).Count(&count)
	if result.Error != nil {
		return false, result.Error
	}
	return count > 0, nil
}

// Team invitations operations

func (repo *TeamRepository) CreateInvitation(db database.Database, invitation *entity.TeamInvitation) error {
	return db.GetDB().Create(invitation).Error
}

func (repo *TeamRepository) FindInvitationByID(db database.Database, id uint) (*entity.TeamInvitation, error) {
	var invitation entity.TeamInvitation
	result := db.GetDB().Preload("Team").Preload("InvitedBy").First(&invitation, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &invitation, nil
}

func (repo *TeamRepository) FindInvitationByTeamAndEmail(db database.Database, teamID uint, email string) (*entity.TeamInvitation, error) {
	var invitation entity.TeamInvitation
	result := db.GetDB().Preload("Team").Preload("InvitedBy").Where("team_id = ? AND email = ?", teamID, email).First(&invitation)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &invitation, nil
}

func (repo *TeamRepository) FindInvitationsByTeamID(db database.Database, teamID uint) ([]*entity.TeamInvitation, error) {
	var invitations []*entity.TeamInvitation
	result := db.GetDB().Preload("Team").Preload("InvitedBy").Where("team_id = ?", teamID).Find(&invitations)
	if result.Error != nil {
		return nil, result.Error
	}
	return invitations, nil
}

func (repo *TeamRepository) FindInvitationsByEmail(db database.Database, email string) ([]*entity.TeamInvitation, error) {
	var invitations []*entity.TeamInvitation
	result := db.GetDB().Preload("Team").Preload("InvitedBy").Where("email = ?", email).Find(&invitations)
	if result.Error != nil {
		return nil, result.Error
	}
	return invitations, nil
}

func (repo *TeamRepository) FindPendingInvitationsByEmail(db database.Database, email string) ([]*entity.TeamInvitation, error) {
	var invitations []*entity.TeamInvitation
	result := db.GetDB().Preload("Team").Preload("InvitedBy").Where("email = ? AND status = ?", email, entity.InvitationStatusPending).Find(&invitations)
	if result.Error != nil {
		return nil, result.Error
	}
	return invitations, nil
}

func (repo *TeamRepository) UpdateInvitation(db database.Database, invitation *entity.TeamInvitation) error {
	return db.GetDB().Save(invitation).Error
}

func (repo *TeamRepository) DeleteInvitation(db database.Database, id uint) error {
	return db.GetDB().Delete(&entity.TeamInvitation{}, id).Error
}

// Find teams with filters

func (repo *TeamRepository) FindTeams(db database.Database, options *postgres.QueryOptions) ([]*entity.Team, int64, error) {
	var teams []*entity.Team
	var total int64

	query := db.GetDB().Preload("Creator").Preload("Members").Preload("Invitations")

	// Count total
	if result := query.Model(&entity.Team{}).Count(&total); result.Error != nil {
		return nil, 0, result.Error
	}

	// Apply options and find
	query = applyQueryOptions(query, options)
	result := query.Find(&teams)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	return teams, total, nil
}
