package enum

type TshirtSize string

const (
	TshirtSizeM   TshirtSize = "M"
	TshirtSizeL   TshirtSize = "L"
	TshirtSizeXL  TshirtSize = "XL"
	TshirtSizeXXL TshirtSize = "XXL"
)

// IsValid checks if the tshirt size is valid
func (t TshirtSize) IsValid() bool {
	switch t {
	case TshirtSizeM, TshirtSizeL, TshirtSizeXL, TshirtSizeXXL:
		return true
	default:
		return false
	}
}

// String returns the string representation
func (t TshirtSize) String() string {
	return string(t)
}
