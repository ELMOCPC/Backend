package postgres

import (
	"time"

	"github.com/ELMOCPC/Backend/internal/domain/entity"
	domainRepo "github.com/ELMOCPC/Backend/internal/domain/repository/postgres"
	"github.com/ELMOCPC/Backend/internal/infrastructure/database"
	"gorm.io/gorm"
)

type TeamInvitationTokenRepositoryImpl struct{}

func NewTeamInvitationTokenRepository() domainRepo.TeamInvitationTokenRepository {
	return &TeamInvitationTokenRepositoryImpl{}
}

func (repo *TeamInvitationTokenRepositoryImpl) CreateToken(db database.Database, token *entity.TeamInvitationToken) error {
	result := db.GetDB().Create(token)
	return result.Error
}

func (repo *TeamInvitationTokenRepositoryImpl) FindTokenByID(db database.Database, id uint) (*entity.TeamInvitationToken, error) {
	var token entity.TeamInvitationToken
	result := db.GetDB().First(&token, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &token, nil
}

func (repo *TeamInvitationTokenRepositoryImpl) FindTokenByToken(db database.Database, tokenStr string) (*entity.TeamInvitationToken, error) {
	var token entity.TeamInvitationToken
	result := db.GetDB().
		Where("token = ? AND expires_at > NOW() AND accepted = ?", tokenStr, false).
		First(&token)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &token, nil
}

func (repo *TeamInvitationTokenRepositoryImpl) FindTokensByTeamID(db database.Database, teamID uint) ([]*entity.TeamInvitationToken, error) {
	var tokens []*entity.TeamInvitationToken
	result := db.GetDB().
		Where("team_id = ?", teamID).
		Order("created_at DESC").
		Find(&tokens)
	return tokens, result.Error
}

func (repo *TeamInvitationTokenRepositoryImpl) FindTokensByEmail(db database.Database, email string) ([]*entity.TeamInvitationToken, error) {
	var tokens []*entity.TeamInvitationToken
	result := db.GetDB().
		Where("email = ? AND expires_at > NOW() AND accepted = ?", email, false).
		Find(&tokens)
	return tokens, result.Error
}

func (repo *TeamInvitationTokenRepositoryImpl) UpdateToken(db database.Database, token *entity.TeamInvitationToken) error {
	result := db.GetDB().Save(token)
	return result.Error
}

func (repo *TeamInvitationTokenRepositoryImpl) DeleteToken(db database.Database, id uint) error {
	result := db.GetDB().Delete(&entity.TeamInvitationToken{}, id)
	return result.Error
}

func (repo *TeamInvitationTokenRepositoryImpl) DeleteExpiredTokens(db database.Database) error {
	result := db.GetDB().Where("expires_at < ?", time.Now()).Delete(&entity.TeamInvitationToken{})
	return result.Error
}
