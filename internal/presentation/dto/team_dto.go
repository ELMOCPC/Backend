package dto

// Request DTOs
type CreateTeamRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
}

type UpdateTeamRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
}

type InviteMemberRequest struct {
	FirstName     string `json:"first_name" validate:"required"`
	LastName      string `json:"last_name" validate:"required"`
	Email         string `json:"email" validate:"required,email"`
	Phone         string `json:"phone" validate:"required"`
	TshirtSize    string `json:"tshirt_size" validate:"required,oneof=XS S M L XL XXL"`
	University    string `json:"university"`
	StudentNumber string `json:"student_number"`
	NationalCode  string `json:"national_code" validate:"required"`
}

type InviteMembersRequest struct {
	Members []InviteMemberRequest `json:"members" validate:"required,min=1"`
}

type AcceptInvitationRequest struct {
	Token string `json:"token" validate:"required"`
}

type RejectInvitationRequest struct {
	Token string `json:"token" validate:"required"`
}

type VerifyReceiptRequest struct {
	Approved bool   `json:"approved"`
	Notes    string `json:"notes"`
}

type RejectTeamRequest struct {
}

// Response DTOs

type PendingInvitationResponse struct {
	ID        uint   `json:"id"`
	Token     string `json:"token"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	ExpiresAt string `json:"expires_at"`
}
type TeamResponse struct {
	ID              uint            `json:"id"`
	Name            string          `json:"name"`
	Description     string          `json:"description"`
	CreatorID       uint            `json:"creator_id"`
	Status          string          `json:"status"`
	Members         []*UserResponse `json:"members"`
	ReceiptImageURL string          `json:"receipt_image_url,omitempty"`
}

type UserResponse struct {
	ID        uint   `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
}

type InvitationTokenResponse struct {
	Token    string `json:"token"`
	TeamID   uint   `json:"team_id"`
	InvitedBy uint   `json:"invited_by"`
	Email    string `json:"email"`
}

type PaymentReceiptResponse struct {
	ID                uint                   `json:"id"`
	TeamID            uint                   `json:"team_id"`
	UploadedByID      uint                   `json:"uploaded_by_id"`
	UploadedByName    string                 `json:"uploaded_by_name"`
	ReceiptImageURL   string                 `json:"receipt_image_url"`
	Status            string                 `json:"status"`
	VerifiedByID      *uint                  `json:"verified_by_id,omitempty"`
	VerifiedByName    string                 `json:"verified_by_name,omitempty"`
	VerificationNotes map[string]interface{} `json:"verification_notes,omitempty"`
	VerifiedAt        *string                `json:"verified_at,omitempty"`
	CreatedAt         string                 `json:"created_at"`
}
