package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/ELMOCPC/Backend/bootstrap"
	"github.com/ELMOCPC/Backend/internal/application/usecase"
	"github.com/ELMOCPC/Backend/internal/domain/entity"
	"github.com/ELMOCPC/Backend/internal/domain/enum"
	"github.com/ELMOCPC/Backend/internal/domain/exception"
	repository "github.com/ELMOCPC/Backend/internal/domain/repository/postgres"
	"github.com/ELMOCPC/Backend/internal/infrastructure/database"
)

type TeamInvitationServiceImpl struct {
	teamInvitationTokenRepo repository.TeamInvitationTokenRepository
	teamRepository          repository.TeamRepository
	userRepository          repository.UserRepository
	emailService            usecase.EmailService
	jwtService              usecase.JwtService
	db                      database.Database
	env                     *bootstrap.Env
}

func NewTeamInvitationService(
	teamInvitationTokenRepo repository.TeamInvitationTokenRepository,
	teamRepository repository.TeamRepository,
	userRepository repository.UserRepository,
	emailService usecase.EmailService,
	jwtService usecase.JwtService,
	db database.Database,
	env *bootstrap.Env,
) usecase.TeamInvitationService {
	return &TeamInvitationServiceImpl{
		teamInvitationTokenRepo: teamInvitationTokenRepo,
		teamRepository:          teamRepository,
		userRepository:          userRepository,
		emailService:            emailService,
		jwtService:              jwtService,
		db:                      db,
		env:                     env,
	}
}

func (tis *TeamInvitationServiceImpl) SendInvitations(teamID, invitedByID uint, members []usecase.InviteMemberData) ([]string, error) {
	// Get team
	team, err := tis.teamRepository.FindTeamByID(tis.db, teamID)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, exception.NotFoundError{Item: "team"}
	}

	// Verify that only the team creator can invite members
	if team.CreatorID != invitedByID {
		return nil, exception.ForbiddenError{Message: "only team creator can invite members"}
	}

	// Check if team is in draft status
	if team.Status != entity.TeamStatusDraft {
		return nil, exception.ForbiddenError{Message: "cannot invite members to a submitted team"}
	}

	// Check team member limit (max 3)
	if len(team.Members)+len(members) > 3 {
		validationErr := exception.ValidationErrors{}
		validationErr.Add("team", "maximum 3 members allowed")
		return nil, validationErr
	}

	var tokens []string

	for _, member := range members {
		// Check if phone is already linked to a different email
		existingUserByPhone, err := tis.userRepository.FindUserByPhone(tis.db, member.Phone)
		if err != nil {
			return nil, err
		}
		if existingUserByPhone != nil && existingUserByPhone.Email != member.Email {
			conflictErr := exception.ConflictErrors{}
			conflictErr.Add("phone", fmt.Sprintf("%s is already linked to another email", member.Phone))
			return nil, conflictErr
		}

		// Check if email exists with a different phone number
		existingUserByEmail, err := tis.userRepository.FindUserByEmail(tis.db, member.Email)
		if err != nil {
			return nil, err
		}
		if existingUserByEmail != nil && existingUserByEmail.Phone != member.Phone {
			conflictErr := exception.ConflictErrors{}
			conflictErr.Add("email", fmt.Sprintf("%s is registered with a different phone number", member.Email))
			return nil, conflictErr
		}

		// Check if user already has a team
		if existingUserByEmail != nil && existingUserByEmail.TeamID != nil {
			conflictErr := exception.ConflictErrors{}
			conflictErr.Add("email", fmt.Sprintf("%s is already in a team", member.Email))
			return nil, conflictErr
		}

		// Check if email is already member
		for _, teamMember := range team.Members {
			if teamMember.Email == member.Email {
				conflictErr := exception.ConflictErrors{}
				conflictErr.Add("email", fmt.Sprintf("%s is already a member", member.Email))
				return nil, conflictErr
			}
		}

		// Check for pending invitations
		existingTokens, err := tis.teamInvitationTokenRepo.FindTokensByEmail(tis.db, member.Email)
		if err != nil {
			return nil, err
		}
		for _, existingToken := range existingTokens {
			if existingToken.TeamID == teamID && !existingToken.Accepted && !existingToken.Rejected {
				conflictErr := exception.ConflictErrors{}
				conflictErr.Add("email", fmt.Sprintf("%s already has a pending invitation", member.Email))
				return nil, conflictErr
			}
		}

		// Generate invitation token
		token := tis.generateToken()

		// Create invitation token with member details
		invitationToken := &entity.TeamInvitationToken{
			TeamID:        teamID,
			InvitedByID:   invitedByID,
			Email:         member.Email,
			Token:         token,
			FirstName:     member.FirstName,
			LastName:      member.LastName,
			Phone:         member.Phone,
			TshirtSize:    member.TshirtSize,
			University:    member.University,
			StudentNumber: member.StudentNumber,
			NationalCode:  member.NationalCode,
			ExpiresAt:     time.Now().Add(7 * 24 * time.Hour), // 7 days
			Accepted:      false,
			Rejected:      false,
		}

		if err := tis.teamInvitationTokenRepo.CreateToken(tis.db, invitationToken); err != nil {
			return nil, err
		}

		// Send email with invitation link
		inviter, err := tis.userRepository.FindUserByID(tis.db, invitedByID)
		if err != nil {
			return nil, err
		}

		invitationLink := fmt.Sprintf(
			"https://elmocpc.ir/accept-invite?token=%s",
			token,
		)

		subject := fmt.Sprintf("You're invited to join team: %s", team.Name)
		body := fmt.Sprintf(`
Dear %s,

%s %s has invited you to join the team "%s" in ELMOCPC.

Please click the link below to accept or reject the invitation:
%s

This link will expire in 7 days.

If you did not expect this invitation, you can ignore this email.

Best regards,
ELMOCPC Team
`, member.FirstName, inviter.FirstName, inviter.LastName, team.Name, invitationLink)

		if err := tis.emailService.SendEmail(member.Email, subject, body); err != nil {
			return nil, err
		}

		tokens = append(tokens, token)
	}

	return tokens, nil
}

func (tis *TeamInvitationServiceImpl) AcceptInvitation(token string) (*entity.User, string, string, error) {
	// Find invitation token
	invitationToken, err := tis.teamInvitationTokenRepo.FindTokenByToken(tis.db, token)
	if err != nil {
		return nil, "", "", err
	}
	if invitationToken == nil {
		return nil, "", "", exception.NotFoundError{Item: "invitation"}
	}

	// Check if token is expired
	if time.Now().After(invitationToken.ExpiresAt) {
		validationErr := exception.ValidationErrors{}
		validationErr.Add("token", "expired")
		return nil, "", "", validationErr
	}

	// Check if already accepted or rejected
	if invitationToken.Accepted {
		validationErr := exception.ValidationErrors{}
		validationErr.Add("invitation", "already accepted")
		return nil, "", "", validationErr
	}
	if invitationToken.Rejected {
		validationErr := exception.ValidationErrors{}
		validationErr.Add("invitation", "already rejected")
		return nil, "", "", validationErr
	}

	email := invitationToken.Email

	// Check if user exists
	existingUser, err := tis.userRepository.FindUserByEmail(tis.db, email)
	if err != nil {
		return nil, "", "", err
	}

	var user *entity.User

	if existingUser != nil {
		// User already exists, validate and add to team
		user = existingUser

		// Check if user is already in a team
		if user.TeamID != nil {
			conflictErr := exception.ConflictErrors{}
			conflictErr.Add("user", "already in a team")
			return nil, "", "", conflictErr
		}

		// Validate that the phone number matches
		if user.Phone != invitationToken.Phone {
			conflictErr := exception.ConflictErrors{}
			conflictErr.Add("phone", "phone number does not match existing account")
			return nil, "", "", conflictErr
		}
	} else {
		// Create new user with stored details
		// Check if phone is unique
		existingUserByPhone, err := tis.userRepository.FindUserByPhone(tis.db, invitationToken.Phone)
		if err != nil {
			return nil, "", "", err
		}
		if existingUserByPhone != nil {
			conflictErr := exception.ConflictErrors{}
			conflictErr.Add("phone", "already exists")
			return nil, "", "", conflictErr
		}

		user = &entity.User{
			FirstName:     invitationToken.FirstName,
			LastName:      invitationToken.LastName,
			Email:         email,
			Phone:         invitationToken.Phone,
			Password:      "", // No password, OTP-based authentication
			TshirtSize:    enum.TshirtSize(invitationToken.TshirtSize),
			University:    invitationToken.University,
			StudentNumber: invitationToken.StudentNumber,
			NationalCode:  invitationToken.NationalCode,
			EmailVerified: true,
		}

		if err := tis.userRepository.CreateUser(tis.db, user); err != nil {
			return nil, "", "", err
		}
	}

	// Add user to team
	if err := tis.teamRepository.AddTeamMember(tis.db, invitationToken.TeamID, user.ID); err != nil {
		return nil, "", "", err
	}

	// Update user's TeamID
	user.TeamID = &invitationToken.TeamID
	if err := tis.userRepository.UpdateUser(tis.db, user); err != nil {
		return nil, "", "", err
	}

	// Mark invitation token as accepted
	now := time.Now()
	invitationToken.Accepted = true
	invitationToken.UsedAt = &now
	if err := tis.teamInvitationTokenRepo.UpdateToken(tis.db, invitationToken); err != nil {
		return nil, "", "", err
	}

	// Generate tokens
	accessToken, refreshToken, err := tis.jwtService.GenerateToken(user.ID)
	if err != nil {
		return nil, "", "", err
	}

	return user, accessToken, refreshToken, nil
}

func (tis *TeamInvitationServiceImpl) RejectInvitation(token string) error {
	// Find invitation token
	invitationToken, err := tis.teamInvitationTokenRepo.FindTokenByToken(tis.db, token)
	if err != nil {
		return err
	}
	if invitationToken == nil {
		return exception.NotFoundError{Item: "invitation"}
	}

	// Check if token is expired
	if time.Now().After(invitationToken.ExpiresAt) {
		validationErr := exception.ValidationErrors{}
		validationErr.Add("token", "expired")
		return validationErr
	}

	// Check if already accepted or rejected
	if invitationToken.Accepted {
		validationErr := exception.ValidationErrors{}
		validationErr.Add("invitation", "already accepted")
		return validationErr
	}
	if invitationToken.Rejected {
		validationErr := exception.ValidationErrors{}
		validationErr.Add("invitation", "already rejected")
		return validationErr
	}

	// Mark invitation token as rejected
	now := time.Now()
	invitationToken.Rejected = true
	invitationToken.UsedAt = &now
	if err := tis.teamInvitationTokenRepo.UpdateToken(tis.db, invitationToken); err != nil {
		return err
	}

	return nil
}

func (tis *TeamInvitationServiceImpl) GetPendingInvitations(email string) ([]*entity.TeamInvitationToken, error) {
	return tis.teamInvitationTokenRepo.FindTokensByEmail(tis.db, email)
}

func (tis *TeamInvitationServiceImpl) generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

var _ usecase.TeamInvitationService = (*TeamInvitationServiceImpl)(nil)
