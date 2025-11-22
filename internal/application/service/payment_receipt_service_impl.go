package service

import (
	"mime/multipart"
	"path/filepath"
	"time"

	"github.com/ELMOCPC/Backend/bootstrap"
	"github.com/ELMOCPC/Backend/internal/application/usecase"
	"github.com/ELMOCPC/Backend/internal/domain/entity"
	"github.com/ELMOCPC/Backend/internal/domain/enum"
	"github.com/ELMOCPC/Backend/internal/domain/exception"
	repository "github.com/ELMOCPC/Backend/internal/domain/repository/postgres"
	"github.com/ELMOCPC/Backend/internal/domain/storage/s3"
	"github.com/ELMOCPC/Backend/internal/infrastructure/database"
	"github.com/google/uuid"
)

type PaymentReceiptServiceImpl struct {
	paymentReceiptRepo repository.PaymentReceiptRepository
	teamRepository     repository.TeamRepository
	userRepository     repository.UserRepository
	emailService       usecase.EmailService
	s3Storage          s3.S3Storage
	constants          *bootstrap.Constants
	db                 database.Database
}

func NewPaymentReceiptService(
	paymentReceiptRepo repository.PaymentReceiptRepository,
	teamRepository repository.TeamRepository,
	userRepository repository.UserRepository,
	emailService usecase.EmailService,
	s3Storage s3.S3Storage,
	constants *bootstrap.Constants,
	db database.Database,
) usecase.PaymentReceiptService {
	return &PaymentReceiptServiceImpl{
		paymentReceiptRepo: paymentReceiptRepo,
		teamRepository:     teamRepository,
		userRepository:     userRepository,
		emailService:       emailService,
		s3Storage:          s3Storage,
		constants:          constants,
		db:                 db,
	}
}

func (prs *PaymentReceiptServiceImpl) UploadReceipt(teamID, userID uint, file *multipart.FileHeader) (*entity.PaymentReceipt, error) {
	team, err := prs.teamRepository.FindTeamByID(prs.db, teamID)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, exception.NotFoundError{Item: "team"}
	}

	if team.Status != entity.TeamStatusWaitingPayment {
		validationErr := exception.ValidationErrors{}
		validationErr.Add("team", "invalid_status")
		return nil, validationErr
	}

	isMember, err := prs.teamRepository.IsTeamMember(prs.db, teamID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, exception.NotFoundError{Item: "team_member"}
	}

	ext := filepath.Ext(file.Filename)
	fileName := uuid.New().String() + ext
	key := prs.constants.BucketPath.GetReceiptPath(teamID, fileName)

	if err := prs.s3Storage.UploadObject(enum.BucketTypeReceipts, key, file); err != nil {
		return nil, err
	}

	receipt := &entity.PaymentReceipt{
		TeamID:          teamID,
		UploadedByID:    userID,
		ReceiptImageURL: key,
		Status:          entity.ReceiptStatusPending,
	}

	if err := prs.paymentReceiptRepo.CreateReceipt(prs.db, receipt); err != nil {
		return nil, err
	}

	team.Status = entity.TeamStatusReceiptPending
	if err := prs.teamRepository.UpdateTeam(prs.db, team); err != nil {
		return nil, err
	}

	members := team.Members
	uploader, err := prs.userRepository.FindUserByID(prs.db, userID)
	if err != nil {
		return nil, err
	}

	for _, member := range members {
		subject := "Payment Receipt Submitted - Team Waiting for Verification"
		body := "Dear " + member.FirstName + ",\n\n" +
			uploader.FirstName + " " + uploader.LastName + " has submitted a payment receipt for team \"" + team.Name + "\".\n\n" +
			"Your team is now waiting for admin verification of the payment.\n\n" +
			"Best regards,\n" +
			"ELMOCPC Team"

		_ = prs.emailService.SendEmail(member.Email, subject, body)
	}

	return receipt, nil
}

func (prs *PaymentReceiptServiceImpl) VerifyReceipt(receiptID uint, adminID uint, approved bool, notes string) (*entity.PaymentReceipt, error) {
	// Find receipt
	receipt, err := prs.paymentReceiptRepo.FindReceiptByID(prs.db, receiptID)
	if err != nil {
		return nil, err
	}
	if receipt == nil {
		return nil, exception.NotFoundError{Item: "payment_receipt"}
	}

	// Update receipt status
	now := time.Now()
	receipt.VerifiedAt = &now
	receipt.VerifiedByID = &adminID

	if approved {
		receipt.Status = entity.ReceiptStatusApproved
	} else {
		receipt.Status = entity.ReceiptStatusRejected
	}

	if notes != "" {
		if receipt.VerificationNotes == nil {
			receipt.VerificationNotes = make(entity.DataMap)
		}
		receipt.VerificationNotes["notes"] = notes
	}

	if err := prs.paymentReceiptRepo.UpdateReceipt(prs.db, receipt); err != nil {
		return nil, err
	}

	// Update team status based on approval
	team := receipt.Team
	if approved {
		team.Status = entity.TeamStatusAccepted
	} else {
		team.Status = entity.TeamStatusWaitingPayment
	}

	if err := prs.teamRepository.UpdateTeam(prs.db, team); err != nil {
		return nil, err
	}

	// Notify team members
	members := team.Members

	for _, member := range members {
		if approved {
			// Send receipt approved notification
			_ = prs.emailService.SendReceiptApprovedNotification(member.Email, member.FirstName, team.Name)
		} else {
			// Send receipt rejected notification
			subject := "Payment Receipt Rejected - Team " + team.Name
			body := "Dear " + member.FirstName + ",\n\n" +
				"Your team's payment receipt has been rejected and needs to be resubmitted.\n\n"
			if notes != "" {
				body += "Admin notes: " + notes + "\n\n"
			}
			body += "Please submit a new receipt for verification.\n\n" +
				"Best regards,\n" +
				"ELMOCPC Team"

			_ = prs.emailService.SendEmail(member.Email, subject, body)
		}
	}

	return receipt, nil
}

func (prs *PaymentReceiptServiceImpl) GetPendingReceipts() ([]*entity.PaymentReceipt, error) {
	return prs.paymentReceiptRepo.FindReceiptsByStatus(prs.db, entity.ReceiptStatusPending)
}

func (prs *PaymentReceiptServiceImpl) GetReceiptByTeamID(teamID uint) (*entity.PaymentReceipt, error) {
	return prs.paymentReceiptRepo.FindReceiptByTeamID(prs.db, teamID)
}

var _ usecase.PaymentReceiptService = (*PaymentReceiptServiceImpl)(nil)
