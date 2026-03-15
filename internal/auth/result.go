package auth

// LoginResult holds the outcome of a login or logout operation.
// It is a direct port of the LoginResult class from the C# WPF application.
type LoginResult struct {
	// Success indicates whether the operation completed successfully.
	Success bool

	// Status is the specific LoginStatus code for this result.
	Status LoginStatus

	// Message is a human-readable description of the result, suitable for display.
	Message string

	// ResponseData contains the raw response body from the server, if available.
	// This field is empty when not applicable (e.g. network failures).
	ResponseData string
}

// NewLoginResult constructs a LoginResult with all fields explicitly provided.
// This is the primary constructor, equivalent to the C# LoginResult(bool, LoginStatus, string, string).
func NewLoginResult(success bool, status LoginStatus, message string, responseData string) LoginResult {
	return LoginResult{
		Success:      success,
		Status:       status,
		Message:      message,
		ResponseData: responseData,
	}
}

// CreateSuccess constructs a successful LoginResult with no raw response data.
// Equivalent to the C# static factory LoginResult.CreateSuccess(LoginStatus, string).
func CreateSuccess(status LoginStatus, message string) LoginResult {
	return LoginResult{
		Success:      true,
		Status:       status,
		Message:      message,
		ResponseData: "",
	}
}

// CreateFailure constructs a failed LoginResult with no raw response data.
// Equivalent to the C# static factory LoginResult.CreateFailure(LoginStatus, string).
func CreateFailure(status LoginStatus, message string) LoginResult {
	return LoginResult{
		Success:      false,
		Status:       status,
		Message:      message,
		ResponseData: "",
	}
}
