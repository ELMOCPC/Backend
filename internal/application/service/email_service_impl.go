package service

import (
	"fmt"
	"net/smtp"

	"github.com/ELMOCPC/Backend/internal/application/usecase"
)

type EmailConfig struct {
	SMTPServer     string
	SMTPPort       string
	SenderEmail    string
	SenderPassword string
	SenderName     string
}

type EmailService struct {
	config *EmailConfig
}

func NewEmailService(config *EmailConfig) *EmailService {
	return &EmailService{
		config: config,
	}
}

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

func (es *EmailService) SendEmail(to, subject, body string) error {
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

func (es *EmailService) sendEmail(to, subject, body string) error {
	return es.SendEmail(to, subject, body)
}

var _ usecase.EmailService = (*EmailService)(nil)
