package service

import (
	"fmt"
	"net/smtp"

	"github.com/ELMOCPC/Backend/internal/application/usecase"
)

type EmailConfig struct {
	SMTPServer string
	SMTPPort   string
	SenderEmail string
	SenderPassword string
	SenderName string
}

type EmailService struct {
	config *EmailConfig
}

func NewEmailService(config *EmailConfig) *EmailService {
	return &EmailService{
		config: config,
	}
}

// SendTeamInvitation sends an email inviting a user to join a team
func (es *EmailService) SendTeamInvitation(toEmail, toName, teamName, inviterName string, invitationLink string) error {
	subject := fmt.Sprintf("You're invited to join team: %s", teamName)
	body := fmt.Sprintf(`
Dear %s,

%s has invited you to join the team "%s" in ELMOCPC.

Please click the link below to accept or reject the invitation:
%s

If you did not expect this invitation, you can ignore this email.

Best regards,
ELMOCPC Team
`, toName, inviterName, teamName, invitationLink)

	return es.sendEmail(toEmail, subject, body)
}

// SendTeamValidationNotification sends an email notifying team members that their team was validated
func (es *EmailService) SendTeamValidationNotification(toEmail, toName, teamName string) error {
	subject := fmt.Sprintf("Team '%s' has been validated", teamName)
	body := fmt.Sprintf(`
Dear %s,

We're excited to inform you that your team "%s" has been validated successfully!

Your team is now registered and ready to participate in ELMOCPC competitions.

Best regards,
ELMOCPC Team
`, toName, teamName)

	return es.sendEmail(toEmail, subject, body)
}

// SendTeamRejectionNotification sends an email notifying team members that their team was rejected
func (es *EmailService) SendTeamRejectionNotification(toEmail, toName, teamName, rejectionReason string) error {
	subject := fmt.Sprintf("Team '%s' validation unsuccessful", teamName)
	body := fmt.Sprintf(`
Dear %s,

Unfortunately, your team "%s" did not pass the validation process.

Rejection Reason:
%s

Please review your team information and try again if you believe this is an error.

Best regards,
ELMOCPC Team
`, toName, teamName, rejectionReason)

	return es.sendEmail(toEmail, subject, body)
}

// sendEmail is a helper method to send emails via SMTP
func (es *EmailService) sendEmail(to, subject, body string) error {
	// Note: This is a basic implementation. For production, you might want to use
	// a package like github.com/go-mail/mail or integrate with a service like SendGrid

	if es.config.SMTPServer == "" {
		// If SMTP is not configured, log the email instead
		fmt.Printf("Email would be sent to: %s\nSubject: %s\nBody: %s\n", to, subject, body)
		return nil
	}

	message := fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s", to, subject, body)

	auth := smtp.PlainAuth("", es.config.SenderEmail, es.config.SenderPassword, es.config.SMTPServer)
	err := smtp.SendMail(
		es.config.SMTPServer+":"+es.config.SMTPPort,
		auth,
		es.config.SenderEmail,
		[]string{to},
		[]byte(message),
	)

	return err
}

// Ensure EmailService implements the EmailService interface
var _ usecase.EmailService = (*EmailService)(nil)
