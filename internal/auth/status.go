package auth

// LoginStatus represents the authentication state of a session.
// It is a direct port of the LoginStatus enum from the C# WPF application.
type LoginStatus int

const (
	// StatusLoggedIn indicates a successful, active session.
	StatusLoggedIn LoginStatus = iota

	// StatusLoggedOut indicates the session has been terminated.
	StatusLoggedOut

	// StatusWrongPassword indicates authentication failed due to invalid credentials.
	StatusWrongPassword

	// StatusMaxLimit indicates the server rejected the login due to maximum session limit.
	StatusMaxLimit

	// StatusLogInAgain indicates the server requires a fresh login attempt.
	StatusLogInAgain

	// StatusDataLimit indicates the session was rejected due to a data quota limit.
	StatusDataLimit

	// StatusNoConnection indicates the network or server is unreachable.
	StatusNoConnection
)

// String returns the human-readable string representation of a LoginStatus value.
func (s LoginStatus) String() string {
	switch s {
	case StatusLoggedIn:
		return "LoggedIn"
	case StatusLoggedOut:
		return "LoggedOut"
	case StatusWrongPassword:
		return "WrongPassword"
	case StatusMaxLimit:
		return "MaxLimit"
	case StatusLogInAgain:
		return "LogInAgain"
	case StatusDataLimit:
		return "DataLimit"
	case StatusNoConnection:
		return "NoConnection"
	default:
		return "Unknown"
	}
}
