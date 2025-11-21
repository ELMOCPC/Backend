package enum

type TshirtSize string

const (
	TshirtSizeXS TshirtSize = "XS"
	TshirtSizeS  TshirtSize = "S"
	TshirtSizeM  TshirtSize = "M"
	TshirtSizeL  TshirtSize = "L"
	TshirtSizeXL TshirtSize = "XL"
	TshirtSizeXXL TshirtSize = "XXL"
)

// IsValid checks if the tshirt size is valid
func (t TshirtSize) IsValid() bool {
	switch t {
	case TshirtSizeXS, TshirtSizeS, TshirtSizeM, TshirtSizeL, TshirtSizeXL, TshirtSizeXXL:
		return true
	default:
		return false
	}
}

// String returns the string representation
func (t TshirtSize) String() string {
	return string(t)
}
