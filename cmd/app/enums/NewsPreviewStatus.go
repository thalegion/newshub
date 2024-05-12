package enums

// Define a type for the enum
type NewsPreviewStatus string

// Define constants for the enum
const (
	NewsPreviewStatusNew         NewsPreviewStatus = "new"
	NewsPreviewStatusFullContent NewsPreviewStatus = "full_content"
)

// Check if the status is valid
func (s NewsPreviewStatus) IsValid() bool {
	switch s {
	case NewsPreviewStatusNew:
		return true
	default:
		return false
	}
}

// String representation for the enum
func (s NewsPreviewStatus) String() string {
	return string(s)
}
