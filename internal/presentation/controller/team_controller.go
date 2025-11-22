package controller

import (
	"mime/multipart"
	"net/http"
	"strconv"
	"time"

	"github.com/ELMOCPC/Backend/internal/application/usecase"
	"github.com/ELMOCPC/Backend/internal/domain/entity"
	"github.com/ELMOCPC/Backend/internal/domain/enum"
	"github.com/ELMOCPC/Backend/internal/domain/exception"
	repositoryPostgres "github.com/ELMOCPC/Backend/internal/domain/repository/postgres"
	"github.com/ELMOCPC/Backend/internal/domain/storage/s3"
	"github.com/ELMOCPC/Backend/internal/infrastructure/database"
	"github.com/ELMOCPC/Backend/internal/presentation/dto"
	"github.com/gin-gonic/gin"
)

type TeamController struct {
	teamService           usecase.TeamService
	teamInvitationService usecase.TeamInvitationService
	paymentReceiptService usecase.PaymentReceiptService
	emailService          usecase.EmailService
	userRepository        repositoryPostgres.UserRepository
	s3Storage             s3.S3Storage
	db                    database.Database
}

func NewTeamController(
	teamService usecase.TeamService,
	teamInvitationService usecase.TeamInvitationService,
	paymentReceiptService usecase.PaymentReceiptService,
	emailService usecase.EmailService,
	userRepository repositoryPostgres.UserRepository,
	s3Storage s3.S3Storage,
	db database.Database,
) *TeamController {
	return &TeamController{
		teamService:           teamService,
		teamInvitationService: teamInvitationService,
		paymentReceiptService: paymentReceiptService,
		emailService:          emailService,
		userRepository:        userRepository,
		s3Storage:             s3Storage,
		db:                    db,
	}
}

func (tc *TeamController) GetTeamStatuses(ctx *gin.Context) {
	statuses := entity.GetAllTeamStatuses()

	response := make([]string, len(statuses))
	for i, status := range statuses {
		response[i] = string(status)
	}

	Response(ctx, http.StatusOK, "team statuses retrieved successfully", response)
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

func (tc *TeamController) SubmitTeam(ctx *gin.Context) {
	userID := ctx.GetUint("ID")

	teamID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		Response(ctx, http.StatusBadRequest, "invalid team id", nil)
		return
	}

	team, err := tc.teamService.SubmitTeam(uint(teamID), userID)
	if err != nil {
		panic(err)
	}

	response := tc.mapTeamResponse(team)
	Response(ctx, http.StatusOK, "team submitted successfully", response)
}

func (tc *TeamController) GetPendingTeamInvitations(ctx *gin.Context) {
	userID := ctx.GetUint("ID")

	teamID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		Response(ctx, http.StatusBadRequest, "invalid team id", nil)
		return
	}

	tokens, err := tc.teamService.GetPendingInvitationTokens(uint(teamID), userID)
	if err != nil {
		panic(err)
	}

	var responses []*dto.PendingInvitationResponse
	for _, token := range tokens {
		responses = append(responses, &dto.PendingInvitationResponse{
			ID:        token.ID,
			Token:     token.Token,
			Email:     token.Email,
			FirstName: token.FirstName,
			LastName:  token.LastName,
			ExpiresAt: token.ExpiresAt.Format("2006-01-02 15:04:05"),
		})
	}

	Response(ctx, http.StatusOK, "pending invitations retrieved successfully", responses)
}

func (tc *TeamController) CancelTeamInvitation(ctx *gin.Context) {
	userID := ctx.GetUint("ID")

	teamID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		Response(ctx, http.StatusBadRequest, "invalid team id", nil)
		return
	}

	token := ctx.Param("tokenId")
	if token == "" {
		Response(ctx, http.StatusBadRequest, "invalid token", nil)
		return
	}

	err = tc.teamService.CancelInvitationToken(uint(teamID), token, userID)
	if err != nil {
		panic(err)
	}

	Response(ctx, http.StatusOK, "invitation cancelled successfully", nil)
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

	// Get receipt URL if exists
	var receiptURL string
	receipt, err := tc.paymentReceiptService.GetReceiptByTeamID(team.ID)
	if err == nil && receipt != nil && receipt.ReceiptImageURL != "" {
		presignedURL, err := tc.s3Storage.GetPresignedURL(enum.BucketTypeReceipts, receipt.ReceiptImageURL, 24*time.Hour)
		if err == nil {
			receiptURL = presignedURL
		}
	}

	return &dto.TeamResponse{
		ID:              team.ID,
		Name:            team.Name,
		Description:     team.Description,
		CreatorID:       team.CreatorID,
		Status:          string(team.Status),
		Members:         members,
		ReceiptImageURL: receiptURL,
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
			NationalCode:  m.NationalCode,
		}
	}

	tokens, err := tc.teamInvitationService.SendInvitations(uint(teamID), userID, members)
	if err != nil {
		switch e := err.(type) {
		case exception.NotFoundError:
			Response(ctx, http.StatusNotFound, "team not found", nil)
		case exception.ForbiddenError:
			Response(ctx, http.StatusForbidden, e.Message, nil)
		case *exception.ConflictErrors:
			panic(err)
		case exception.ValidationErrors:
			panic(err)
		default:
			panic(err)
		}
		return
	}

	invitations := make([]map[string]interface{}, len(members))
	for i, member := range members {
		invitations[i] = map[string]interface{}{
			"token": tokens[i],
			"email": member.Email,
		}
	}

	response := map[string]interface{}{
		"team_id":     teamID,
		"invitations": invitations,
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
	type UploadReceiptParams struct {
		TeamID  uint                  `uri:"id" validate:"required"`
		Receipt *multipart.FileHeader `form:"receipt" validate:"required"`
	}

	params := Validated[UploadReceiptParams](ctx)
	userID := ctx.GetUint("ID")

	receipt, err := tc.paymentReceiptService.UploadReceipt(params.TeamID, userID, params.Receipt)
	if err != nil {
		panic(err)
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

func (tc *TeamController) GetPendingTeams(ctx *gin.Context) {
	teams, err := tc.teamService.GetPendingTeams()
	if err != nil {
		panic(err)
	}

	var responses []*dto.TeamResponse
	for _, team := range teams {
		responses = append(responses, tc.mapTeamResponse(team))
	}

	Response(ctx, http.StatusOK, "pending teams retrieved successfully", responses)
}

func (tc *TeamController) ValidateTeam(ctx *gin.Context) {
	teamID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		Response(ctx, http.StatusBadRequest, "invalid team id", nil)
		return
	}

	team, err := tc.teamService.ValidateTeam(uint(teamID))
	if err != nil {
		panic(err)
	}

	for _, member := range team.Members {
		_ = tc.emailService.SendTeamApprovedNotification(member.Email, member.FirstName, team.Name)
	}

	response := tc.mapTeamResponse(team)
	Response(ctx, http.StatusOK, "team validated successfully", response)
}

func (tc *TeamController) RejectTeam(ctx *gin.Context) {
	teamID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		Response(ctx, http.StatusBadRequest, "invalid team id", nil)
		return
	}

	team, err := tc.teamService.RejectTeam(uint(teamID))
	if err != nil {
		panic(err)
	}

	for _, member := range team.Members {
		_ = tc.emailService.SendTeamRejectionNotification(member.Email, member.FirstName, team.Name)
	}

	response := tc.mapTeamResponse(team)
	Response(ctx, http.StatusOK, "team rejected successfully", response)
}

func (tc *TeamController) GetAllTeams(ctx *gin.Context) {
	status := ctx.Query("status")

	var teams []*entity.Team
	var err error

	if status != "" {
		teams, err = tc.teamService.GetTeamsByStatus(entity.TeamStatus(status))
	} else {
		teams, err = tc.teamService.GetAllTeams()
	}

	if err != nil {
		panic(err)
	}

	var responses []*dto.TeamResponse
	for _, team := range teams {
		responses = append(responses, tc.mapTeamResponse(team))
	}

	Response(ctx, http.StatusOK, "teams retrieved successfully", responses)
}

func (tc *TeamController) GetTeamByIDAdmin(ctx *gin.Context) {
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

	// Generate presigned URL for the receipt image
	receiptURL := receipt.ReceiptImageURL
	if receipt.ReceiptImageURL != "" {
		presignedURL, err := tc.s3Storage.GetPresignedURL(enum.BucketTypeReceipts, receipt.ReceiptImageURL, 24*time.Hour)
		if err == nil {
			receiptURL = presignedURL
		}
	}

	return &dto.PaymentReceiptResponse{
		ID:                receipt.ID,
		TeamID:            receipt.TeamID,
		UploadedByID:      receipt.UploadedByID,
		UploadedByName:    uploadedByName,
		ReceiptImageURL:   receiptURL,
		Status:            string(receipt.Status),
		VerifiedByID:      receipt.VerifiedByID,
		VerifiedByName:    verifiedByName,
		VerificationNotes: receipt.VerificationNotes,
		VerifiedAt:        verifiedAtStr,
		CreatedAt:         receipt.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}
