package service

import (
	"github.com/ELMOCPC/Backend/internal/application/usecase"
	"github.com/ELMOCPC/Backend/internal/domain/entity"
	"github.com/ELMOCPC/Backend/internal/domain/exception"
	repository "github.com/ELMOCPC/Backend/internal/domain/repository/postgres"
	"github.com/ELMOCPC/Backend/internal/infrastructure/database"
)

type TeamService struct {
	teamRepository           repository.TeamRepository
	userRepository           repository.UserRepository
	teamInvitationTokenRepo  repository.TeamInvitationTokenRepository
	db                       database.Database
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

// CreateTeam creates a new team with the given user as creator
func (ts *TeamService) CreateTeam(userID uint, name, description string) (*entity.Team, error) {
	// Check if user already has a team
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

	// Add creator as a member
	if err := ts.teamRepository.AddTeamMember(ts.db, team.ID, userID); err != nil {
		return nil, err
	}

	// Update user's team
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

// GetTeamByID retrieves a team by its ID
func (ts *TeamService) GetTeamByID(teamID uint) (*entity.Team, error) {
	return ts.teamRepository.FindTeamByID(ts.db, teamID)
}

// GetUserTeam retrieves the team of a user
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

// UpdateTeam updates team details (only if not submitted)
func (ts *TeamService) UpdateTeam(teamID, userID uint, name, description string) (*entity.Team, error) {
	team, err := ts.teamRepository.FindTeamByID(ts.db, teamID)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, exception.NotFoundError{Item: "team"}
	}

	// Check if user is the creator
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

// DeleteTeam deletes a team (only creator can delete, and only if in draft status)
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

	// Check if team is still in draft status
	if team.Status != entity.TeamStatusDraft {
		return exception.ForbiddenError{Resource: "team", Message: "cannot delete a submitted team"}
	}

	// Delete the team
	if err := ts.teamRepository.DeleteTeam(ts.db, teamID); err != nil {
		return err
	}

	// Remove team reference from all members
	members, err := ts.teamRepository.GetTeamMembers(ts.db, teamID)
	if err == nil {
		for _, member := range members {
			member.TeamID = nil
			_ = ts.userRepository.UpdateUser(ts.db, member)
		}
	}

	return nil
}

// InviteUserByEmail invites a user to the team by email
func (ts *TeamService) InviteUserByEmail(teamID, inviterID uint, email string) (*entity.TeamInvitation, error) {
	team, err := ts.teamRepository.FindTeamByID(ts.db, teamID)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, exception.NotFoundError{Item: "team"}
	}

	// Check if inviter is team creator
	if team.CreatorID != inviterID {
		return nil, exception.ForbiddenError{Resource: "team", Message: "only team creator can invite members"}
	}

	// Check if team is in draft status
	if team.Status != entity.TeamStatusDraft {
		return nil, exception.ForbiddenError{Resource: "team", Message: "cannot invite members to a submitted team"}
	}

	// Check if user exists before inviting
	invitedUser, err := ts.userRepository.FindUserByEmail(ts.db, email)
	if err != nil {
		return nil, err
	}
	if invitedUser == nil {
		conflictErr := exception.ConflictErrors{}
		conflictErr.Add("email", "user_not_found")
		return nil, conflictErr
	}

	// Check if email is already in team
	isMember, err := ts.teamRepository.IsTeamMember(ts.db, teamID, invitedUser.ID)
	if err != nil {
		return nil, err
	}
	if isMember {
		conflictErr := exception.ConflictErrors{}
		conflictErr.Add("email", "already_team_member")
		return nil, conflictErr
	}

	// Check if invitation already exists for this email
	existingInvite, err := ts.teamRepository.FindInvitationByTeamAndEmail(ts.db, teamID, email)
	if err != nil {
		return nil, err
	}
	if existingInvite != nil && existingInvite.Status == entity.InvitationStatusPending {
		conflictErr := exception.ConflictErrors{}
		conflictErr.Add("email", "invitation_already_sent")
		return nil, conflictErr
	}

	// Check team member count
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

// AcceptInvitation accepts an invitation
func (ts *TeamService) AcceptInvitation(invitationID, userID uint) error {
	invitation, err := ts.teamRepository.FindInvitationByID(ts.db, invitationID)
	if err != nil {
		return err
	}
	if invitation == nil {
		return exception.NotFoundError{Item: "invitation"}
	}

	// Check if user's email matches the invitation email
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

	// Check if invitation is pending
	if invitation.Status != entity.InvitationStatusPending {
		conflictErr := exception.ConflictErrors{}
		conflictErr.Add("invitation", "invalid_status")
		return conflictErr
	}

	// Check if user already has a team
	if user.TeamID != nil {
		conflictErr := exception.ConflictErrors{}
		conflictErr.Add("user", "already_in_team")
		return conflictErr
	}

	// Get team and check status
	team, err := ts.teamRepository.FindTeamByID(ts.db, invitation.TeamID)
	if err != nil {
		return err
	}
	if team == nil {
		return exception.NotFoundError{Item: "team"}
	}

	// Cannot accept invitation if team is submitted
	if team.Status != entity.TeamStatusDraft {
		return exception.ForbiddenError{Resource: "team", Message: "cannot join a submitted team"}
	}

	// Add user to team
	if err := ts.teamRepository.AddTeamMember(ts.db, invitation.TeamID, userID); err != nil {
		return err
	}

	// Update user's team
	user.TeamID = &invitation.TeamID
	if err := ts.userRepository.UpdateUser(ts.db, user); err != nil {
		return err
	}

	// Update invitation status
	invitation.Status = entity.InvitationStatusAccepted
	if err := ts.teamRepository.UpdateInvitation(ts.db, invitation); err != nil {
		return err
	}

	return nil
}

// RejectInvitation rejects an invitation
func (ts *TeamService) RejectInvitation(invitationID, userID uint) error {
	invitation, err := ts.teamRepository.FindInvitationByID(ts.db, invitationID)
	if err != nil {
		return err
	}
	if invitation == nil {
		return exception.NotFoundError{Item: "invitation"}
	}

	// Check if user's email matches the invitation email
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

	// Check if invitation is pending
	if invitation.Status != entity.InvitationStatusPending {
		conflictErr := exception.ConflictErrors{}
		conflictErr.Add("invitation", "invalid_status")
		return conflictErr
	}

	// Update invitation status
	invitation.Status = entity.InvitationStatusRejected
	if err := ts.teamRepository.UpdateInvitation(ts.db, invitation); err != nil {
		return err
	}

	return nil
}

// RemoveMember removes a member from the team (creator only)
func (ts *TeamService) RemoveMember(teamID, userID, memberToRemoveID uint) error {
	team, err := ts.teamRepository.FindTeamByID(ts.db, teamID)
	if err != nil {
		return err
	}
	if team == nil {
		return exception.NotFoundError{Item: "team"}
	}

	// Check if user is the creator
	if team.CreatorID != userID {
		return exception.ForbiddenError{Resource: "team", Message: "only team creator can remove members"}
	}

	// Prevent removing creator
	if team.CreatorID == memberToRemoveID {
		return exception.ForbiddenError{Resource: "team", Message: "cannot remove team creator"}
	}

	// Check if team is still in draft
	if team.Status != entity.TeamStatusDraft {
		return exception.ForbiddenError{Resource: "team", Message: "cannot remove members from a submitted team"}
	}

	// Check if member is actually in the team
	isMember, err := ts.teamRepository.IsTeamMember(ts.db, teamID, memberToRemoveID)
	if err != nil {
		return err
	}
	if !isMember {
		conflictErr := exception.ConflictErrors{}
		conflictErr.Add("member", "not_in_team")
		return conflictErr
	}

	// Remove member
	if err := ts.teamRepository.RemoveTeamMember(ts.db, teamID, memberToRemoveID); err != nil {
		return err
	}

	// Update user's team
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

// LeaveTeam allows a member to leave the team
func (ts *TeamService) LeaveTeam(teamID, userID uint) error {
	team, err := ts.teamRepository.FindTeamByID(ts.db, teamID)
	if err != nil {
		return err
	}
	if team == nil {
		return exception.NotFoundError{Item: "team"}
	}

	// Check if user is the creator
	if team.CreatorID == userID {
		conflictErr := exception.ConflictErrors{}
		conflictErr.Add("team", "team_leader_cannot_leave")
		return conflictErr
	}

	// Check if team is still in draft
	if team.Status != entity.TeamStatusDraft {
		return exception.ForbiddenError{Resource: "team", Message: "cannot leave a submitted team"}
	}

	// Remove member
	if err := ts.teamRepository.RemoveTeamMember(ts.db, teamID, userID); err != nil {
		return err
	}

	// Update user's team
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

// GetTeamMembers retrieves all members of a team
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

// SubmitTeam submits a team for validation (locks team)
func (ts *TeamService) SubmitTeam(teamID, userID uint) (*entity.Team, error) {
	team, err := ts.teamRepository.FindTeamByID(ts.db, teamID)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, exception.NotFoundError{Item: "team"}
	}

	// Check if user is the creator
	if team.CreatorID != userID {
		return nil, exception.ForbiddenError{Resource: "team", Message: "only team creator can submit team"}
	}

	// Check if team is in draft status
	if team.Status != entity.TeamStatusDraft {
		conflictErr := exception.ConflictErrors{}
		conflictErr.Add("team", "invalid_status")
		return nil, conflictErr
	}

	// Check if team has at least 1 member
	memberCount, err := ts.teamRepository.GetTeamMemberCount(ts.db, teamID)
	if err != nil {
		return nil, err
	}
	if memberCount == 0 {
		conflictErr := exception.ConflictErrors{}
		conflictErr.Add("team", "no_members")
		return nil, conflictErr
	}

	// Check if team has pending invitations (new token-based invitations)
	pendingTokens, err := ts.teamInvitationTokenRepo.FindTokensByTeamID(ts.db, teamID)
	if err != nil {
		return nil, err
	}

	// Count pending tokens (not accepted and not rejected)
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

	// Check if team has pending invitations (legacy invitations)
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

	// Update team status
	team.Status = entity.TeamStatusSubmitted
	if err := ts.teamRepository.UpdateTeam(ts.db, team); err != nil {
		return nil, err
	}

	return team, nil
}

// GetUserInvitations retrieves all invitations for a user's email
func (ts *TeamService) GetUserInvitations(email string) ([]*entity.TeamInvitation, error) {
	return ts.teamRepository.FindInvitationsByEmail(ts.db, email)
}

// GetPendingInvitationTokens retrieves all pending invitation tokens for a team
func (ts *TeamService) GetPendingInvitationTokens(teamID, userID uint) ([]*entity.TeamInvitationToken, error) {
	team, err := ts.teamRepository.FindTeamByID(ts.db, teamID)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, exception.NotFoundError{Item: "team"}
	}

	// Check if user is the creator
	if team.CreatorID != userID {
		return nil, exception.ForbiddenError{Resource: "team", Message: "only team creator can view pending invitations"}
	}

	// Get all tokens for the team
	tokens, err := ts.teamInvitationTokenRepo.FindTokensByTeamID(ts.db, teamID)
	if err != nil {
		return nil, err
	}

	// Filter for pending tokens (not accepted and not rejected)
	var pendingTokens []*entity.TeamInvitationToken
	for _, token := range tokens {
		if !token.Accepted && !token.Rejected {
			pendingTokens = append(pendingTokens, token)
		}
	}

	return pendingTokens, nil
}

// CancelInvitation cancels a pending invitation (only team creator can cancel)
func (ts *TeamService) CancelInvitation(invitationID, userID uint) error {
	invitation, err := ts.teamRepository.FindInvitationByID(ts.db, invitationID)
	if err != nil {
		return err
	}
	if invitation == nil {
		return exception.NotFoundError{Item: "invitation"}
	}

	// Get team to check if user is creator
	team, err := ts.teamRepository.FindTeamByID(ts.db, invitation.TeamID)
	if err != nil {
		return err
	}
	if team == nil {
		return exception.NotFoundError{Item: "team"}
	}

	// Check if user is the team creator
	if team.CreatorID != userID {
		return exception.ForbiddenError{Resource: "invitation", Message: "only team creator can cancel invitations"}
	}

	// Check if invitation is still pending
	if invitation.Status != entity.InvitationStatusPending {
		conflictErr := exception.ConflictErrors{}
		conflictErr.Add("invitation", "invalid_status")
		return conflictErr
	}

	// Delete the invitation
	if err := ts.teamRepository.DeleteInvitation(ts.db, invitationID); err != nil {
		return err
	}

	return nil
}

// CancelInvitationToken cancels a pending invitation token (only team creator can cancel)
func (ts *TeamService) CancelInvitationToken(tokenID, userID uint) error {
	// Find the token by ID
	token, err := ts.teamInvitationTokenRepo.FindTokenByID(ts.db, tokenID)
	if err != nil {
		return err
	}
	if token == nil {
		return exception.NotFoundError{Item: "invitation_token"}
	}

	// Get team to check if user is creator
	team, err := ts.teamRepository.FindTeamByID(ts.db, token.TeamID)
	if err != nil {
		return err
	}
	if team == nil {
		return exception.NotFoundError{Item: "team"}
	}

	// Check if user is the team creator
	if team.CreatorID != userID {
		return exception.ForbiddenError{Resource: "invitation_token", Message: "only team creator can cancel invitations"}
	}

	// Check if token is still pending (not accepted and not rejected)
	if token.Accepted || token.Rejected {
		conflictErr := exception.ConflictErrors{}
		conflictErr.Add("invitation_token", "invalid_status")
		return conflictErr
	}

	// Delete the token
	if err := ts.teamInvitationTokenRepo.DeleteToken(ts.db, tokenID); err != nil {
		return err
	}

	return nil
}

// CanModifyTeam checks if a user can modify a team
func (ts *TeamService) CanModifyTeam(teamID, userID uint) (bool, error) {
	team, err := ts.teamRepository.FindTeamByID(ts.db, teamID)
	if err != nil {
		return false, err
	}
	if team == nil {
		return false, nil
	}

	// Only creator can modify draft teams
	if team.CreatorID != userID {
		return false, nil
	}

	// Team must be in draft status
	if team.Status != entity.TeamStatusDraft {
		return false, nil
	}

	return true, nil
}

// ValidateTeam approves a submitted team (admin operation)
func (ts *TeamService) ValidateTeam(teamID uint) (*entity.Team, error) {
	team, err := ts.teamRepository.FindTeamByID(ts.db, teamID)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, exception.NotFoundError{Item: "team"}
	}

	// Check if team is in submitted status
	if team.Status != entity.TeamStatusSubmitted {
		conflictErr := exception.ConflictErrors{}
		conflictErr.Add("team", "invalid_status_for_validation")
		return nil, conflictErr
	}

	// Update team status to waiting_for_payment
	team.Status = entity.TeamStatusWaitingPayment
	if err := ts.teamRepository.UpdateTeam(ts.db, team); err != nil {
		return nil, err
	}

	return team, nil
}

// RejectTeam rejects a submitted team (admin operation)
func (ts *TeamService) RejectTeam(teamID uint, rejectionReason string) (*entity.Team, error) {
	team, err := ts.teamRepository.FindTeamByID(ts.db, teamID)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, exception.NotFoundError{Item: "team"}
	}

	// Check if team is in submitted status
	if team.Status != entity.TeamStatusSubmitted {
		conflictErr := exception.ConflictErrors{}
		conflictErr.Add("team", "invalid_status_for_rejection")
		return nil, conflictErr
	}

	// Update team status to rejected
	team.Status = entity.TeamStatusRejected
	if err := ts.teamRepository.UpdateTeam(ts.db, team); err != nil {
		return nil, err
	}

	return team, nil
}

// GetPendingTeams retrieves all teams pending admin validation
func (ts *TeamService) GetPendingTeams() ([]*entity.Team, error) {
	return ts.teamRepository.FindTeamsByStatus(ts.db, entity.TeamStatusSubmitted)
}

// GetAllTeams retrieves all teams (admin operation)
func (ts *TeamService) GetAllTeams() ([]*entity.Team, error) {
	// Use FindTeams with nil options to get all teams
	teams, _, err := ts.teamRepository.FindTeams(ts.db, nil)
	return teams, err
}

// GetTeamsByStatus retrieves teams by specific status (admin operation)
func (ts *TeamService) GetTeamsByStatus(status entity.TeamStatus) ([]*entity.Team, error) {
	return ts.teamRepository.FindTeamsByStatus(ts.db, status)
}

// Ensure TeamService implements the TeamService interface
var _ usecase.TeamService = (*TeamService)(nil)
