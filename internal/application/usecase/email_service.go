package usecase

type EmailService interface {
	SendEmail(to, subject, body string) error
	SendTeamInvitation(toEmail, toName, teamName, inviterName string, invitationLink string) error
	SendTeamValidationNotification(toEmail, toName, teamName string) error
	SendTeamRejectionNotification(toEmail, toName, teamName string) error
	SendTeamApprovedNotification(toEmail, toName, teamName string) error
	SendTeamResetNotification(toEmail, toName, teamName string) error
	SendReceiptApprovedNotification(toEmail, toName, teamName string) error
}
