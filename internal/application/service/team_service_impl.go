package service

import (
	"github.com/ELMOCPC/Backend/internal/application/usecase"
	"github.com/ELMOCPC/Backend/internal/domain/entity"
	"github.com/ELMOCPC/Backend/internal/domain/exception"
	repository "github.com/ELMOCPC/Backend/internal/domain/repository/postgres"
	"github.com/ELMOCPC/Backend/internal/infrastructure/database"
)

type TeamService struct {
	teamRepository          repository.TeamRepository
	userRepository          repository.UserRepository
	teamInvitationTokenRepo repository.TeamInvitationTokenRepository
	db                      database.Database
}

func NewTeamService(
	teamRepository repository.TeamRepository,
	userRepository repository.UserRepository,
	teamInvitationTokenRepo repository.TeamInvitationTokenRepository,
	db database.Database,
) *TeamService {
	return &TeamService{
		teamRepository:          teamRepository,
		userRepository:          userRepository,
		teamInvitationTokenRepo: teamInvitationTokenRepo,
		db:                      db,
	}
}

func (ts *TeamService) CreateTeam(userID uint, name, description string) (*entity.Team, error) {
	existingTeam, err := ts.teamRepository.FindTeamByUserID(ts.db, userID)
	if err != nil {
		return nil, err
	}
	if existingTeam != nil {
		conflictErr := exception.ConflictErrors{}
		conflictErr.Add("team", "user_already_has_team")
		return nil, conflictErr
	}

	team := &entity.Team{
		Name:        name,
		Description: description,
		CreatorID:   userID,
		Status:      entity.TeamStatusDraft,
	}

	if err := ts.teamRepository.CreateTeam(ts.db, team); err != nil {
		return nil, err
	}

	if err := ts.teamRepository.AddTeamMember(ts.db, team.ID, userID); err != nil {
		return nil, err
	}

	user, err := ts.userRepository.FindUserByID(ts.db, userID)
	if err != nil {
		return nil, err
	}
	if user != nil {
		user.TeamID = &team.ID
		if err := ts.userRepository.UpdateUser(ts.db, user); err != nil {
			return nil, err
		}
	}

	return team, nil
}

func (ts *TeamService) GetTeamByID(teamID uint) (*entity.Team, error) {
	return ts.teamRepository.FindTeamByID(ts.db, teamID)
}

func (ts *TeamService) GetUserTeam(userID uint) (*entity.Team, error) {
	user, err := ts.userRepository.FindUserByID(ts.db, userID)
	if err != nil {
		return nil, err
	}
	if user == nil || user.TeamID == nil {
		return nil, nil
	}
	return ts.teamRepository.FindTeamByID(ts.db, *user.TeamID)
}

func (ts *TeamService) UpdateTeam(teamID, userID uint, name, description string) (*entity.Team, error) {
	team, err := ts.teamRepository.FindTeamByID(ts.db, teamID)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, exception.NotFoundError{Item: "team"}
	}

	if team.CreatorID != userID {
		return nil, exception.ForbiddenError{Resource: "team", Message: "only team creator can update team details"}
	}

	// Check if team is still in draft status
	if team.Status != entity.TeamStatusDraft {
		return nil, exception.ForbiddenError{Resource: "team", Message: "cannot modify a submitted team"}
	}

	team.Name = name
	team.Description = description

	if err := ts.teamRepository.UpdateTeam(ts.db, team); err != nil {
		return nil, err
	}

	return team, nil
}

func (ts *TeamService) DeleteTeam(teamID, userID uint) error {
	team, err := ts.teamRepository.FindTeamByID(ts.db, teamID)
	if err != nil {
		return err
	}
	if team == nil {
		return exception.NotFoundError{Item: "team"}
	}

	// Check if user is the creator
	if team.CreatorID != userID {
		return exception.ForbiddenError{Resource: "team", Message: "only team creator can delete team"}
	}

	if team.Status != entity.TeamStatusDraft {
		return exception.ForbiddenError{Resource: "team", Message: "cannot delete a submitted team"}
	}

	if err := ts.teamRepository.DeleteTeam(ts.db, teamID); err != nil {
		return err
	}

	members, err := ts.teamRepository.GetTeamMembers(ts.db, teamID)
	if err == nil {
		for _, member := range members {
			member.TeamID = nil
			_ = ts.userRepository.UpdateUser(ts.db, member)
		}
	}

	return nil
}

func (ts *TeamService) InviteUserByEmail(teamID, inviterID uint, email string) (*entity.TeamInvitation, error) {
	team, err := ts.teamRepository.FindTeamByID(ts.db, teamID)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, exception.NotFoundError{Item: "team"}
	}

	if team.CreatorID != inviterID {
		return nil, exception.ForbiddenError{Resource: "team", Message: "only team creator can invite members"}
	}

	if team.Status != entity.TeamStatusDraft {
		return nil, exception.ForbiddenError{Resource: "team", Message: "cannot invite members to a submitted team"}
	}

	invitedUser, err := ts.userRepository.FindUserByEmail(ts.db, email)
	if err != nil {
		return nil, err
	}
	if invitedUser == nil {
		conflictErr := exception.ConflictErrors{}
		conflictErr.Add("email", "user_not_found")
		return nil, conflictErr
	}

	isMember, err := ts.teamRepository.IsTeamMember(ts.db, teamID, invitedUser.ID)
	if err != nil {
		return nil, err
	}
	if isMember {
		conflictErr := exception.ConflictErrors{}
		conflictErr.Add("email", "already_team_member")
		return nil, conflictErr
	}

	existingInvite, err := ts.teamRepository.FindInvitationByTeamAndEmail(ts.db, teamID, email)
	if err != nil {
		return nil, err
	}
	if existingInvite != nil && existingInvite.Status == entity.InvitationStatusPending {
		conflictErr := exception.ConflictErrors{}
		conflictErr.Add("email", "invitation_already_sent")
		return nil, conflictErr
	}

	memberCount, err := ts.teamRepository.GetTeamMemberCount(ts.db, teamID)
	if err != nil {
		return nil, err
	}
	if memberCount >= 3 {
		conflictErr := exception.ConflictErrors{}
		conflictErr.Add("team", "team_full")
		return nil, conflictErr
	}

	invitation := &entity.TeamInvitation{
		TeamID:      teamID,
		Email:       email,
		Status:      entity.InvitationStatusPending,
		InvitedByID: inviterID,
	}

	if err := ts.teamRepository.CreateInvitation(ts.db, invitation); err != nil {
		return nil, err
	}

	return invitation, nil
}

func (ts *TeamService) AcceptInvitation(invitationID, userID uint) error {
	invitation, err := ts.teamRepository.FindInvitationByID(ts.db, invitationID)
	if err != nil {
		return err
	}
	if invitation == nil {
		return exception.NotFoundError{Item: "invitation"}
	}

	user, err := ts.userRepository.FindUserByID(ts.db, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return exception.NotFoundError{Item: "user"}
	}

	if user.Email != invitation.Email {
		return exception.ForbiddenError{Resource: "invitation", Message: "invitation email does not match user email"}
	}

	if invitation.Status != entity.InvitationStatusPending {
		conflictErr := exception.ConflictErrors{}
		conflictErr.Add("invitation", "invalid_status")
		return conflictErr
	}

	if user.TeamID != nil {
		conflictErr := exception.ConflictErrors{}
		conflictErr.Add("user", "already_in_team")
		return conflictErr
	}

	team, err := ts.teamRepository.FindTeamByID(ts.db, invitation.TeamID)
	if err != nil {
		return err
	}
	if team == nil {
		return exception.NotFoundError{Item: "team"}
	}

	if team.Status != entity.TeamStatusDraft {
		return exception.ForbiddenError{Resource: "team", Message: "cannot join a submitted team"}
	}

	if err := ts.teamRepository.AddTeamMember(ts.db, invitation.TeamID, userID); err != nil {
		return err
	}

	user.TeamID = &invitation.TeamID
	if err := ts.userRepository.UpdateUser(ts.db, user); err != nil {
		return err
	}

	invitation.Status = entity.InvitationStatusAccepted
	if err := ts.teamRepository.UpdateInvitation(ts.db, invitation); err != nil {
		return err
	}

	return nil
}

func (ts *TeamService) RejectInvitation(invitationID, userID uint) error {
	invitation, err := ts.teamRepository.FindInvitationByID(ts.db, invitationID)
	if err != nil {
		return err
	}
	if invitation == nil {
		return exception.NotFoundError{Item: "invitation"}
	}

	user, err := ts.userRepository.FindUserByID(ts.db, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return exception.NotFoundError{Item: "user"}
	}

	if user.Email != invitation.Email {
		return exception.ForbiddenError{Resource: "invitation", Message: "invitation email does not match user email"}
	}

	if invitation.Status != entity.InvitationStatusPending {
		conflictErr := exception.ConflictErrors{}
		conflictErr.Add("invitation", "invalid_status")
		return conflictErr
	}

	invitation.Status = entity.InvitationStatusRejected
	if err := ts.teamRepository.UpdateInvitation(ts.db, invitation); err != nil {
		return err
	}

	return nil
}

func (ts *TeamService) RemoveMember(teamID, userID, memberToRemoveID uint) error {
	team, err := ts.teamRepository.FindTeamByID(ts.db, teamID)
	if err != nil {
		return err
	}
	if team == nil {
		return exception.NotFoundError{Item: "team"}
	}

	if team.CreatorID != userID {
		return exception.ForbiddenError{Resource: "team", Message: "only team creator can remove members"}
	}

	if team.CreatorID == memberToRemoveID {
		return exception.ForbiddenError{Resource: "team", Message: "cannot remove team creator"}
	}

	if team.Status != entity.TeamStatusDraft {
		return exception.ForbiddenError{Resource: "team", Message: "cannot remove members from a submitted team"}
	}

	isMember, err := ts.teamRepository.IsTeamMember(ts.db, teamID, memberToRemoveID)
	if err != nil {
		return err
	}
	if !isMember {
		conflictErr := exception.ConflictErrors{}
		conflictErr.Add("member", "not_in_team")
		return conflictErr
	}

	if err := ts.teamRepository.RemoveTeamMember(ts.db, teamID, memberToRemoveID); err != nil {
		return err
	}

	member, err := ts.userRepository.FindUserByID(ts.db, memberToRemoveID)
	if err != nil {
		return err
	}
	if member != nil {
		member.TeamID = nil
		if err := ts.userRepository.UpdateUser(ts.db, member); err != nil {
			return err
		}
	}

	return nil
}

func (ts *TeamService) LeaveTeam(teamID, userID uint) error {
	team, err := ts.teamRepository.FindTeamByID(ts.db, teamID)
	if err != nil {
		return err
	}
	if team == nil {
		return exception.NotFoundError{Item: "team"}
	}

	if team.CreatorID == userID {
		conflictErr := exception.ConflictErrors{}
		conflictErr.Add("team", "team_leader_cannot_leave")
		return conflictErr
	}

	if team.Status != entity.TeamStatusDraft {
		return exception.ForbiddenError{Resource: "team", Message: "cannot leave a submitted team"}
	}

	if err := ts.teamRepository.RemoveTeamMember(ts.db, teamID, userID); err != nil {
		return err
	}

	user, err := ts.userRepository.FindUserByID(ts.db, userID)
	if err != nil {
		return err
	}
	if user != nil {
		user.TeamID = nil
		if err := ts.userRepository.UpdateUser(ts.db, user); err != nil {
			return err
		}
	}

	return nil
}

func (ts *TeamService) GetTeamMembers(teamID uint) ([]*entity.User, error) {
	team, err := ts.teamRepository.FindTeamByID(ts.db, teamID)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, exception.NotFoundError{Item: "team"}
	}

	return ts.teamRepository.GetTeamMembers(ts.db, teamID)
}

func (ts *TeamService) SubmitTeam(teamID, userID uint) (*entity.Team, error) {
	team, err := ts.teamRepository.FindTeamByID(ts.db, teamID)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, exception.NotFoundError{Item: "team"}
	}

	if team.CreatorID != userID {
		return nil, exception.ForbiddenError{Resource: "team", Message: "only team creator can submit team"}
	}

	if team.Status != entity.TeamStatusDraft {
		conflictErr := exception.ConflictErrors{}
		conflictErr.Add("team", "invalid_status")
		return nil, conflictErr
	}

	memberCount, err := ts.teamRepository.GetTeamMemberCount(ts.db, teamID)
	if err != nil {
		return nil, err
	}
	if memberCount == 0 {
		conflictErr := exception.ConflictErrors{}
		conflictErr.Add("team", "no_members")
		return nil, conflictErr
	}

	pendingTokens, err := ts.teamInvitationTokenRepo.FindTokensByTeamID(ts.db, teamID)
	if err != nil {
		return nil, err
	}

	pendingCount := 0
	for _, token := range pendingTokens {
		if !token.Accepted && !token.Rejected {
			pendingCount++
		}
	}

	if pendingCount > 0 {
		conflictErr := exception.ConflictErrors{}
		conflictErr.Add("team", "pending_invitations")
		return nil, conflictErr
	}

	invitations, err := ts.teamRepository.FindInvitationsByTeamID(ts.db, teamID)
	if err != nil {
		return nil, err
	}

	for _, invitation := range invitations {
		if invitation.Status == entity.InvitationStatusPending {
			conflictErr := exception.ConflictErrors{}
			conflictErr.Add("team", "pending_invitations")
			return nil, conflictErr
		}
	}

	team.Status = entity.TeamStatusSubmitted
	if err := ts.teamRepository.UpdateTeam(ts.db, team); err != nil {
		return nil, err
	}

	return team, nil
}

func (ts *TeamService) GetUserInvitations(email string) ([]*entity.TeamInvitation, error) {
	return ts.teamRepository.FindInvitationsByEmail(ts.db, email)
}

func (ts *TeamService) GetPendingInvitationTokens(teamID, userID uint) ([]*entity.TeamInvitationToken, error) {
	team, err := ts.teamRepository.FindTeamByID(ts.db, teamID)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, exception.NotFoundError{Item: "team"}
	}

	if team.CreatorID != userID {
		return nil, exception.ForbiddenError{Resource: "team", Message: "only team creator can view pending invitations"}
	}

	tokens, err := ts.teamInvitationTokenRepo.FindTokensByTeamID(ts.db, teamID)
	if err != nil {
		return nil, err
	}

	var pendingTokens []*entity.TeamInvitationToken
	for _, token := range tokens {
		if !token.Accepted && !token.Rejected {
			pendingTokens = append(pendingTokens, token)
		}
	}

	return pendingTokens, nil
}

func (ts *TeamService) CancelInvitation(invitationID, userID uint) error {
	invitation, err := ts.teamRepository.FindInvitationByID(ts.db, invitationID)
	if err != nil {
		return err
	}
	if invitation == nil {
		return exception.NotFoundError{Item: "invitation"}
	}

	team, err := ts.teamRepository.FindTeamByID(ts.db, invitation.TeamID)
	if err != nil {
		return err
	}
	if team == nil {
		return exception.NotFoundError{Item: "team"}
	}

	if team.CreatorID != userID {
		return exception.ForbiddenError{Resource: "invitation", Message: "only team creator can cancel invitations"}
	}

	if invitation.Status != entity.InvitationStatusPending {
		conflictErr := exception.ConflictErrors{}
		conflictErr.Add("invitation", "invalid_status")
		return conflictErr
	}

	if err := ts.teamRepository.DeleteInvitation(ts.db, invitationID); err != nil {
		return err
	}

	return nil
}

func (ts *TeamService) CancelInvitationToken(teamID uint, token string, userID uint) error {
	invitationToken, err := ts.teamInvitationTokenRepo.FindTokenByToken(ts.db, token)
	if err != nil {
		return err
	}
	if invitationToken == nil {
		return exception.NotFoundError{Item: "invitation_token"}
	}

	if invitationToken.TeamID != teamID {
		return exception.NotFoundError{Item: "invitation_token"}
	}

	team, err := ts.teamRepository.FindTeamByID(ts.db, teamID)
	if err != nil {
		return err
	}
	if team == nil {
		return exception.NotFoundError{Item: "team"}
	}

	if team.CreatorID != userID {
		return exception.ForbiddenError{Resource: "invitation_token", Message: "only team creator can cancel invitations"}
	}

	if invitationToken.Accepted || invitationToken.Rejected {
		conflictErr := exception.ConflictErrors{}
		conflictErr.Add("invitation_token", "invalid_status")
		return conflictErr
	}

	if err := ts.teamInvitationTokenRepo.DeleteToken(ts.db, invitationToken.ID); err != nil {
		return err
	}

	return nil
}

func (ts *TeamService) CanModifyTeam(teamID, userID uint) (bool, error) {
	team, err := ts.teamRepository.FindTeamByID(ts.db, teamID)
	if err != nil {
		return false, err
	}
	if team == nil {
		return false, nil
	}

	if team.CreatorID != userID {
		return false, nil
	}

	if team.Status != entity.TeamStatusDraft {
		return false, nil
	}

	return true, nil
}

func (ts *TeamService) ValidateTeam(teamID uint) (*entity.Team, error) {
	team, err := ts.teamRepository.FindTeamByID(ts.db, teamID)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, exception.NotFoundError{Item: "team"}
	}

	if team.Status != entity.TeamStatusSubmitted {
		conflictErr := exception.ConflictErrors{}
		conflictErr.Add("team", "invalid_status")
		return nil, conflictErr
	}

	team.Status = entity.TeamStatusWaitingPayment
	if err := ts.teamRepository.UpdateTeam(ts.db, team); err != nil {
		return nil, err
	}

	return team, nil
}

func (ts *TeamService) RejectTeam(teamID uint) (*entity.Team, error) {
	team, err := ts.teamRepository.FindTeamByID(ts.db, teamID)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, exception.NotFoundError{Item: "team"}
	}

	if team.Status != entity.TeamStatusSubmitted {
		conflictErr := exception.ConflictErrors{}
		conflictErr.Add("team", "invalid_status")
		return nil, conflictErr
	}

	team.Status = entity.TeamStatusRejected
	if err := ts.teamRepository.UpdateTeam(ts.db, team); err != nil {
		return nil, err
	}

	return team, nil
}

func (ts *TeamService) GetPendingTeams() ([]*entity.Team, error) {
	return ts.teamRepository.FindTeamsByStatus(ts.db, entity.TeamStatusSubmitted)
}

func (ts *TeamService) GetAllTeams() ([]*entity.Team, error) {
	teams, _, err := ts.teamRepository.FindTeams(ts.db, nil)
	return teams, err
}

func (ts *TeamService) GetTeamsByStatus(status entity.TeamStatus) ([]*entity.Team, error) {
	return ts.teamRepository.FindTeamsByStatus(ts.db, status)
}

var _ usecase.TeamService = (*TeamService)(nil)
