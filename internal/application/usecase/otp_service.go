package usecase

type OTPService interface {
	GenerateOTP(email string) (string, error)
	ValidateOTP(email, code string) error
	SendOTPEmail(email string) error
}
