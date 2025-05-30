package models

// KYCRequest represents the request structure for KYC verification
type KYCRequest struct {
	Email string `form:"email" json:"email" validate:"required,email"`
}

// KYCResponse represents the response structure for KYC verification
type KYCResponse struct {
	Success    bool    `json:"success"`
	Verified   bool    `json:"verified"`
	Similarity float32 `json:"similarity,omitempty"`
	Message    string  `json:"message"`
	Error      string  `json:"error,omitempty"`
}

// FaceValidationCriteria defines the minimum requirements for face validation
type FaceValidationCriteria struct {
	MinConfidence float32
	MinBrightness float32
	MinSharpness  float32
	MinSimilarity float32
}

// DefaultFaceValidationCriteria returns the default validation criteria
func DefaultFaceValidationCriteria() FaceValidationCriteria {
	return FaceValidationCriteria{
		MinConfidence: 90.0,
		MinBrightness: 50.0,
		MinSharpness:  50.0,
		MinSimilarity: 70.0,
	}
}

// VerificationResult holds the complete verification result
type VerificationResult struct {
	Verified   bool
	Similarity float32
	Message    string
}
