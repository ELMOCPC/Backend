package service

import (
	"errors"
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

// LoginAuth implements smtp.Auth for LOGIN authentication
type LoginAuth struct {
	username string
	password string
}

func (a *LoginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte{}, nil
}

func (a *LoginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch string(fromServer) {
		case "Username:":
			return []byte(a.username), nil
		case "Password:":
			return []byte(a.password), nil
		default:
			return nil, errors.New("unknown from server")
		}
	}
	return nil, nil
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

func (es *EmailService) SendTeamRejectionNotification(toEmail, toName, teamName string) error {
	subject := fmt.Sprintf("Team '%s' validation unsuccessful", teamName)
	body := fmt.Sprintf(`
Dear %s,

Unfortunately, your team "%s" did not pass the validation process.

Please review your team information and try again if you believe this is an error.

Best regards,
ELMOCPC Team
`, toName, teamName)
	return es.sendEmail(toEmail, subject, body)
}

func (es *EmailService) SendTeamApprovedNotification(toEmail, toName, teamName string) error {
	subject := fmt.Sprintf("Team '%s' approved - Please upload payment receipt", teamName)
	body := fmt.Sprintf(`
Dear %s,

Great news! Your team "%s" has been approved by our admins.

Next Step: Please upload your payment receipt for verification. Once your receipt is validated, your team will be officially accepted to the contest.

Best regards,
ELMOCPC Team
`, toName, teamName)
	return es.sendEmail(toEmail, subject, body)
}

func (es *EmailService) SendTeamResetNotification(toEmail, toName, teamName string) error {
	subject := fmt.Sprintf("Team '%s' reset - Please resubmit", teamName)
	body := fmt.Sprintf(`
Dear %s,

Your team "%s" has been reset and is now back in draft status.

You can now review your team information and resubmit when ready.

Best regards,
ELMOCPC Team
`, toName, teamName)
	return es.sendEmail(toEmail, subject, body)
}

func (es *EmailService) SendReceiptApprovedNotification(toEmail, toName, teamName string) error {
	subject := fmt.Sprintf("Congratulations! Team '%s' officially accepted", teamName)
	body := fmt.Sprintf(`
Dear %s,

Congratulations! Your team "%s" has been officially accepted to ELMOCPC.

Your payment receipt has been validated and approved. Your team is now ready to participate in all contest activities.

Best regards,
ELMOCPC Team
`, toName, teamName)
	return es.sendEmail(toEmail, subject, body)
}

func (es *EmailService) SendEmail(to, subject, body string) error {
	from := fmt.Sprintf("%s <%s>", es.config.SenderName, es.config.SenderEmail)
	message := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		from, to, subject, body)

	auth := &LoginAuth{
		username: es.config.SenderEmail,
		password: es.config.SenderPassword,
	}

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
