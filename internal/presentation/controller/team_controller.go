package controller

import (
	"net/http"
	"strconv"

	"github.com/ELMOCPC/Backend/internal/application/usecase"
	"github.com/ELMOCPC/Backend/internal/domain/entity"
	"github.com/ELMOCPC/Backend/internal/domain/exception"
	"github.com/gin-gonic/gin"
)

// Request DTOs
type CreateTeamRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
}

type UpdateTeamRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
}

type InviteUserRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// Response DTOs
type TeamResponse struct {
	ID          uint                     `json:"id"`
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	CreatorID   uint                     `json:"creator_id"`
	Status      string                   `json:"status"`
	Members     []*UserResponse          `json:"members"`
	Invitations []*InvitationResponse    `json:"invitations,omitempty"`
}

type UserResponse struct {
	ID        uint   `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
}

type InvitationResponse struct {
	ID      uint   `json:"id"`
	Email   string `json:"email"`
	Status  string `json:"status"`
	TeamID  uint   `json:"team_id"`
}

type TeamController struct {
	teamService usecase.TeamService
}

func NewTeamController(teamService usecase.TeamService) *TeamController {
	return &TeamController{
		teamService: teamService,
	}
}

// CreateTeam - POST /teams
func (tc *TeamController) CreateTeam(ctx *gin.Context) {
	req := Validated[CreateTeamRequest](ctx)
	userID := ctx.GetUint("ID")

	team, err := tc.teamService.CreateTeam(userID, req.Name, req.Description)
	if err != nil {
		// Let recovery middleware handle all error types and formatting
		panic(err)
	}

	response := tc.mapTeamResponse(team)
	Response(ctx, http.StatusCreated, "team created successfully", response)
}

// GetTeam - GET /teams/:id
func (tc *TeamController) GetTeam(ctx *gin.Context) {
	teamID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		Response(ctx, http.StatusBadRequest, "invalid team id", nil)
		return
	}

	team, err := tc.teamService.GetTeamByID(uint(teamID))
	if err != nil {
		panic(err)
	}
	if team == nil {
		panic(exception.NotFoundError{Item: "team"})
	}

	response := tc.mapTeamResponse(team)
	Response(ctx, http.StatusOK, "team retrieved successfully", response)
}

// GetUserTeam - GET /teams/user/current
func (tc *TeamController) GetUserTeam(ctx *gin.Context) {
	userID := ctx.GetUint("ID")

	team, err := tc.teamService.GetUserTeam(userID)
	if err != nil {
		panic(err)
	}
	if team == nil {
		Response(ctx, http.StatusOK, "user has no team", nil)
		return
	}

	response := tc.mapTeamResponse(team)
	Response(ctx, http.StatusOK, "team retrieved successfully", response)
}

// UpdateTeam - PUT /teams/:id
func (tc *TeamController) UpdateTeam(ctx *gin.Context) {
	req := Validated[UpdateTeamRequest](ctx)
	userID := ctx.GetUint("ID")

	teamID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		Response(ctx, http.StatusBadRequest, "invalid team id", nil)
		return
	}

	team, err := tc.teamService.UpdateTeam(uint(teamID), userID, req.Name, req.Description)
	if err != nil {
		// Let recovery middleware handle all error types and formatting
		panic(err)
	}

	response := tc.mapTeamResponse(team)
	Response(ctx, http.StatusOK, "team updated successfully", response)
}

// InviteUser - POST /teams/:id/invitations
func (tc *TeamController) InviteUser(ctx *gin.Context) {
	req := Validated[InviteUserRequest](ctx)
	userID := ctx.GetUint("ID")

	teamID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		Response(ctx, http.StatusBadRequest, "invalid team id", nil)
		return
	}

	invitation, err := tc.teamService.InviteUserByEmail(uint(teamID), userID, req.Email)
	if err != nil {
		// Let recovery middleware handle all error types and formatting
		panic(err)
	}

	response := &InvitationResponse{
		ID:     invitation.ID,
		Email:  invitation.Email,
		Status: string(invitation.Status),
		TeamID: invitation.TeamID,
	}
	Response(ctx, http.StatusCreated, "invitation sent successfully", response)
}

// AcceptInvitation - POST /invitations/:id/accept
func (tc *TeamController) AcceptInvitation(ctx *gin.Context) {
	userID := ctx.GetUint("ID")

	invitationID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		Response(ctx, http.StatusBadRequest, "invalid invitation id", nil)
		return
	}

	err = tc.teamService.AcceptInvitation(uint(invitationID), userID)
	if err != nil {
		// Let recovery middleware handle all error types and formatting
		panic(err)
	}

	Response(ctx, http.StatusOK, "invitation accepted successfully", nil)
}

// RejectInvitation - POST /invitations/:id/reject
func (tc *TeamController) RejectInvitation(ctx *gin.Context) {
	userID := ctx.GetUint("ID")

	invitationID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		Response(ctx, http.StatusBadRequest, "invalid invitation id", nil)
		return
	}

	err = tc.teamService.RejectInvitation(uint(invitationID), userID)
	if err != nil {
		// Let recovery middleware handle all error types and formatting
		panic(err)
	}

	Response(ctx, http.StatusOK, "invitation rejected successfully", nil)
}

// GetUserInvitations - GET /invitations
func (tc *TeamController) GetUserInvitations(ctx *gin.Context) {
	// Get user to fetch email
	// This is a simplified approach - in practice you'd pass the email from the service or context
	invitations, err := tc.teamService.GetUserInvitations("")
	if err != nil {
		panic(err)
	}

	invitationResponses := make([]*InvitationResponse, len(invitations))
	for i, inv := range invitations {
		invitationResponses[i] = &InvitationResponse{
			ID:     inv.ID,
			Email:  inv.Email,
			Status: string(inv.Status),
			TeamID: inv.TeamID,
		}
	}

	Response(ctx, http.StatusOK, "invitations retrieved successfully", invitationResponses)
}

// CancelInvitation - DELETE /invitations/:id
func (tc *TeamController) CancelInvitation(ctx *gin.Context) {
	userID := ctx.GetUint("ID")

	invitationID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		Response(ctx, http.StatusBadRequest, "invalid invitation id", nil)
		return
	}

	err = tc.teamService.CancelInvitation(uint(invitationID), userID)
	if err != nil {
		// Let recovery middleware handle all error types and formatting
		panic(err)
	}

	Response(ctx, http.StatusOK, "invitation cancelled successfully", nil)
}

// RemoveMember - DELETE /teams/:id/members/:memberId
func (tc *TeamController) RemoveMember(ctx *gin.Context) {
	userID := ctx.GetUint("ID")

	teamID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		Response(ctx, http.StatusBadRequest, "invalid team id", nil)
		return
	}

	memberID, err := strconv.ParseUint(ctx.Param("memberId"), 10, 32)
	if err != nil {
		Response(ctx, http.StatusBadRequest, "invalid member id", nil)
		return
	}

	err = tc.teamService.RemoveMember(uint(teamID), userID, uint(memberID))
	if err != nil {
		// Let recovery middleware handle all error types and formatting
		panic(err)
	}

	Response(ctx, http.StatusOK, "member removed successfully", nil)
}

// LeaveTeam - POST /teams/:id/leave
func (tc *TeamController) LeaveTeam(ctx *gin.Context) {
	userID := ctx.GetUint("ID")

	teamID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		Response(ctx, http.StatusBadRequest, "invalid team id", nil)
		return
	}

	err = tc.teamService.LeaveTeam(uint(teamID), userID)
	if err != nil {
		// Let recovery middleware handle all error types and formatting
		panic(err)
	}

	Response(ctx, http.StatusOK, "team left successfully", nil)
}

// GetTeamMembers - GET /teams/:id/members
func (tc *TeamController) GetTeamMembers(ctx *gin.Context) {
	teamID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		Response(ctx, http.StatusBadRequest, "invalid team id", nil)
		return
	}

	members, err := tc.teamService.GetTeamMembers(uint(teamID))
	if err != nil {
		// Let recovery middleware handle all error types and formatting
		panic(err)
	}

	memberResponses := make([]*UserResponse, len(members))
	for i, member := range members {
		memberResponses[i] = &UserResponse{
			ID:        member.ID,
			FirstName: member.FirstName,
			LastName:  member.LastName,
			Email:     member.Email,
			Phone:     member.Phone,
		}
	}

	Response(ctx, http.StatusOK, "members retrieved successfully", memberResponses)
}

// SubmitTeam - POST /teams/:id/submit
func (tc *TeamController) SubmitTeam(ctx *gin.Context) {
	userID := ctx.GetUint("ID")

	teamID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		Response(ctx, http.StatusBadRequest, "invalid team id", nil)
		return
	}

	team, err := tc.teamService.SubmitTeam(uint(teamID), userID)
	if err != nil {
		// Let recovery middleware handle all error types and formatting
		panic(err)
	}

	response := tc.mapTeamResponse(team)
	Response(ctx, http.StatusOK, "team submitted successfully", response)
}

// Helper method to map Team entity to TeamResponse
func (tc *TeamController) mapTeamResponse(team *entity.Team) *TeamResponse {
	members := make([]*UserResponse, len(team.Members))
	for i, member := range team.Members {
		members[i] = &UserResponse{
			ID:        member.ID,
			FirstName: member.FirstName,
			LastName:  member.LastName,
			Email:     member.Email,
			Phone:     member.Phone,
		}
	}

	invitations := make([]*InvitationResponse, len(team.Invitations))
	for i, inv := range team.Invitations {
		invitations[i] = &InvitationResponse{
			ID:     inv.ID,
			Email:  inv.Email,
			Status: string(inv.Status),
			TeamID: inv.TeamID,
		}
	}

	return &TeamResponse{
		ID:          team.ID,
		Name:        team.Name,
		Description: team.Description,
		CreatorID:   team.CreatorID,
		Status:      string(team.Status),
		Members:     members,
		Invitations: invitations,
	}
}
