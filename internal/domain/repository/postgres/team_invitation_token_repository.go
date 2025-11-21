package postgres

import (
	"github.com/ELMOCPC/Backend/internal/domain/entity"
	"github.com/ELMOCPC/Backend/internal/infrastructure/database"
)

type TeamInvitationTokenRepository interface {
	CreateToken(db database.Database, token *entity.TeamInvitationToken) error
	FindTokenByToken(db database.Database, tokenStr string) (*entity.TeamInvitationToken, error)
	FindTokensByTeamID(db database.Database, teamID uint) ([]*entity.TeamInvitationToken, error)
	FindTokensByEmail(db database.Database, email string) ([]*entity.TeamInvitationToken, error)
	UpdateToken(db database.Database, token *entity.TeamInvitationToken) error
	DeleteToken(db database.Database, id uint) error
	DeleteExpiredTokens(db database.Database) error
}
