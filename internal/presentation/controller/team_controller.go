package controller

import (
	"net/http"
	"strconv"

	"github.com/ELMOCPC/Backend/internal/application/usecase"
	"github.com/ELMOCPC/Backend/internal/domain/entity"
	"github.com/ELMOCPC/Backend/internal/domain/exception"
	repositoryPostgres "github.com/ELMOCPC/Backend/internal/domain/repository/postgres"
	"github.com/ELMOCPC/Backend/internal/infrastructure/database"
	"github.com/ELMOCPC/Backend/internal/presentation/dto"
	"github.com/gin-gonic/gin"
)


type TeamController struct {
	teamService           usecase.TeamService
	teamInvitationService usecase.TeamInvitationService
	paymentReceiptService usecase.PaymentReceiptService
	userRepository        repositoryPostgres.UserRepository
	db                    database.Database
}

func NewTeamController(
	teamService usecase.TeamService,
	teamInvitationService usecase.TeamInvitationService,
	paymentReceiptService usecase.PaymentReceiptService,
	userRepository repositoryPostgres.UserRepository,
	db database.Database,
) *TeamController {
	return &TeamController{
		teamService:           teamService,
		teamInvitationService: teamInvitationService,
		paymentReceiptService: paymentReceiptService,
		userRepository:        userRepository,
		db:                    db,
	}
}

func (tc *TeamController) CreateTeam(ctx *gin.Context) {
	req := Validated[dto.CreateTeamRequest](ctx)
	userID := ctx.GetUint("ID")

	team, err := tc.teamService.CreateTeam(userID, req.Name, req.Description)
	if err != nil {
		panic(err)
	}

	response := tc.mapTeamResponse(team)
	Response(ctx, http.StatusCreated, "team created successfully", response)
}

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

func (tc *TeamController) UpdateTeam(ctx *gin.Context) {
	req := Validated[dto.UpdateTeamRequest](ctx)
	userID := ctx.GetUint("ID")

	teamID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		Response(ctx, http.StatusBadRequest, "invalid team id", nil)
		return
	}

	team, err := tc.teamService.UpdateTeam(uint(teamID), userID, req.Name, req.Description)
	if err != nil {
		panic(err)
	}

	response := tc.mapTeamResponse(team)
	Response(ctx, http.StatusOK, "team updated successfully", response)
}

func (tc *TeamController) DeleteTeam(ctx *gin.Context) {
	userID := ctx.GetUint("ID")

	teamID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		Response(ctx, http.StatusBadRequest, "invalid team id", nil)
		return
	}

	err = tc.teamService.DeleteTeam(uint(teamID), userID)
	if err != nil {
		panic(err)
	}

	Response(ctx, http.StatusOK, "team deleted successfully", nil)
}

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
		panic(err)
	}

	Response(ctx, http.StatusOK, "member removed successfully", nil)
}

func (tc *TeamController) LeaveTeam(ctx *gin.Context) {
	userID := ctx.GetUint("ID")

	teamID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		Response(ctx, http.StatusBadRequest, "invalid team id", nil)
		return
	}

	err = tc.teamService.LeaveTeam(uint(teamID), userID)
	if err != nil {
		panic(err)
	}

	Response(ctx, http.StatusOK, "team left successfully", nil)
}

func (tc *TeamController) GetTeamMembers(ctx *gin.Context) {
	teamID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		Response(ctx, http.StatusBadRequest, "invalid team id", nil)
		return
	}

	members, err := tc.teamService.GetTeamMembers(uint(teamID))
	if err != nil {
		panic(err)
	}

	memberResponses := make([]*dto.UserResponse, len(members))
	for i, member := range members {
		memberResponses[i] = &dto.UserResponse{
			ID:        member.ID,
			FirstName: member.FirstName,
			LastName:  member.LastName,
			Email:     member.Email,
			Phone:     member.Phone,
		}
	}

	Response(ctx, http.StatusOK, "members retrieved successfully", memberResponses)
}

func (tc *TeamController) mapTeamResponse(team *entity.Team) *dto.TeamResponse {
	members := make([]*dto.UserResponse, len(team.Members))
	for i, member := range team.Members {
		members[i] = &dto.UserResponse{
			ID:        member.ID,
			FirstName: member.FirstName,
			LastName:  member.LastName,
			Email:     member.Email,
			Phone:     member.Phone,
		}
	}

	return &dto.TeamResponse{
		ID:          team.ID,
		Name:        team.Name,
		Description: team.Description,
		CreatorID:   team.CreatorID,
		Status:      string(team.Status),
		Members:     members,
	}
}

func (tc *TeamController) InviteMembers(ctx *gin.Context) {
	req := Validated[dto.InviteMembersRequest](ctx)
	userID := ctx.GetUint("ID")
	teamID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		Response(ctx, http.StatusBadRequest, "invalid team id", nil)
		return
	}

	members := make([]usecase.InviteMemberData, len(req.Members))
	for i, m := range req.Members {
		members[i] = usecase.InviteMemberData{
			FirstName:     m.FirstName,
			LastName:      m.LastName,
			Email:         m.Email,
			Phone:         m.Phone,
			TshirtSize:    m.TshirtSize,
			University:    m.University,
			StudentNumber: m.StudentNumber,
		}
	}

	tokens, err := tc.teamInvitationService.SendInvitations(uint(teamID), userID, members)
	if err != nil {
		switch err.(type) {
		case exception.NotFoundError:
			Response(ctx, http.StatusNotFound, "team not found", nil)
		case *exception.ConflictErrors:
			panic(err)
		case exception.ValidationErrors:
			panic(err)
		default:
			panic(err)
		}
		return
	}

	response := map[string]interface{}{
		"tokens":  tokens,
		"team_id": teamID,
	}
	Response(ctx, http.StatusOK, "invitations sent successfully", response)
}

func (tc *TeamController) AcceptTeamInvitation(ctx *gin.Context) {
	req := Validated[dto.AcceptInvitationRequest](ctx)

	user, accessToken, refreshToken, err := tc.teamInvitationService.AcceptInvitation(req.Token)
	if err != nil {
		switch err.(type) {
		case exception.NotFoundError:
			Response(ctx, http.StatusNotFound, "invitation not found or expired", nil)
		case *exception.ConflictErrors:
			panic(err)
		case exception.ValidationErrors:
			panic(err)
		default:
			panic(err)
		}
		return
	}

	response := &dto.AuthResponse{
		ID:           user.ID,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		Email:        user.Email,
		Phone:        user.Phone,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	Response(ctx, http.StatusCreated, "invitation accepted successfully", response)
}

func (tc *TeamController) RejectTeamInvitation(ctx *gin.Context) {
	req := Validated[dto.RejectInvitationRequest](ctx)

	err := tc.teamInvitationService.RejectInvitation(req.Token)
	if err != nil {
		switch err.(type) {
		case exception.NotFoundError:
			Response(ctx, http.StatusNotFound, "invitation not found or expired", nil)
		case exception.ValidationErrors:
			panic(err)
		default:
			panic(err)
		}
		return
	}

	Response(ctx, http.StatusOK, "invitation rejected successfully", nil)
}

func (tc *TeamController) GetPendingInvitations(ctx *gin.Context) {
	email := ctx.Query("email")
	if email == "" {
		Response(ctx, http.StatusBadRequest, "email is required", nil)
		return
	}

	tokens, err := tc.teamInvitationService.GetPendingInvitations(email)
	if err != nil {
		panic(err)
	}

	var responses []*dto.InvitationTokenResponse
	for _, token := range tokens {
		responses = append(responses, &dto.InvitationTokenResponse{
			Token:     token.Token,
			TeamID:    token.TeamID,
			InvitedBy: token.InvitedByID,
			Email:     token.Email,
		})
	}

	Response(ctx, http.StatusOK, "pending invitations retrieved successfully", responses)
}

func (tc *TeamController) UploadPaymentReceipt(ctx *gin.Context) {
	req := Validated[dto.UploadReceiptRequest](ctx)
	userID := ctx.GetUint("ID")

	teamID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		Response(ctx, http.StatusBadRequest, "invalid team id", nil)
		return
	}

	receipt, err := tc.paymentReceiptService.UploadReceipt(uint(teamID), userID, req.ReceiptImageURL)
	if err != nil {
		switch err.(type) {
		case exception.NotFoundError:
			panic(err)
		case exception.ValidationErrors:
			panic(err)
		default:
			panic(err)
		}
		return
	}

	response := tc.mapReceiptResponse(receipt)
	Response(ctx, http.StatusCreated, "receipt uploaded successfully", response)
}

func (tc *TeamController) VerifyPaymentReceipt(ctx *gin.Context) {
	req := Validated[dto.VerifyReceiptRequest](ctx)
	adminID := ctx.GetUint("ID")

	receiptID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		Response(ctx, http.StatusBadRequest, "invalid receipt id", nil)
		return
	}

	receipt, err := tc.paymentReceiptService.VerifyReceipt(uint(receiptID), adminID, req.Approved, req.Notes)
	if err != nil {
		switch err.(type) {
		case exception.NotFoundError:
			panic(err)
		default:
			panic(err)
		}
		return
	}

	response := tc.mapReceiptResponse(receipt)
	Response(ctx, http.StatusOK, "receipt verified successfully", response)
}

func (tc *TeamController) GetPendingReceipts(ctx *gin.Context) {
	receipts, err := tc.paymentReceiptService.GetPendingReceipts()
	if err != nil {
		panic(err)
	}

	var responses []*dto.PaymentReceiptResponse
	for _, receipt := range receipts {
		responses = append(responses, tc.mapReceiptResponse(receipt))
	}

	Response(ctx, http.StatusOK, "pending receipts retrieved successfully", responses)
}

func (tc *TeamController) GetTeamReceipt(ctx *gin.Context) {
	teamID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		Response(ctx, http.StatusBadRequest, "invalid team id", nil)
		return
	}

	receipt, err := tc.paymentReceiptService.GetReceiptByTeamID(uint(teamID))
	if err != nil {
		panic(err)
	}

	if receipt == nil {
		Response(ctx, http.StatusNotFound, "no receipt found for this team", nil)
		return
	}

	response := tc.mapReceiptResponse(receipt)
	Response(ctx, http.StatusOK, "receipt retrieved successfully", response)
}

func (tc *TeamController) mapReceiptResponse(receipt *entity.PaymentReceipt) *dto.PaymentReceiptResponse {
	var verifiedAtStr *string
	if receipt.VerifiedAt != nil {
		t := receipt.VerifiedAt.Format("2006-01-02 15:04:05")
		verifiedAtStr = &t
	}

	uploadedByName := ""
	if receipt.UploadedBy != nil {
		uploadedByName = receipt.UploadedBy.FirstName + " " + receipt.UploadedBy.LastName
	}

	verifiedByName := ""
	if receipt.VerifiedBy != nil {
		verifiedByName = receipt.VerifiedBy.FirstName + " " + receipt.VerifiedBy.LastName
	}

	return &dto.PaymentReceiptResponse{
		ID:                receipt.ID,
		TeamID:            receipt.TeamID,
		UploadedByID:      receipt.UploadedByID,
		UploadedByName:    uploadedByName,
		ReceiptImageURL:   receipt.ReceiptImageURL,
		Status:            string(receipt.Status),
		VerifiedByID:      receipt.VerifiedByID,
		VerifiedByName:    verifiedByName,
		VerificationNotes: receipt.VerificationNotes,
		VerifiedAt:        verifiedAtStr,
		CreatedAt:         receipt.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}
