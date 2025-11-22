package dto

type LoginRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type SendOTPRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type VerifyOTPRequest struct {
	Email string `json:"email" validate:"required,email"`
	Code  string `json:"code" validate:"required"`
}

type CompleteSignUpRequest struct {
	FirstName     string `json:"first_name" validate:"required"`
	LastName      string `json:"last_name" validate:"required"`
	Email         string `json:"email" validate:"required,email"`
	Phone         string `json:"phone" validate:"required"`
	TshirtSize    string `json:"tshirt_size" validate:"required,oneof=XS S M L XL XXL"`
	University    string `json:"university"`
	StudentNumber string `json:"student_number"`
	NationalCode  string `json:"national_code" validate:"required"`
}

type AuthResponse struct {
	ID           uint   `json:"id"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type OTPVerificationResponse struct {
	Message string `json:"message"`
}
