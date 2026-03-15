package auth

import "testing"

func TestLoginStatusString(t *testing.T) {
	tests := []struct {
		status   LoginStatus
		expected string
	}{
		{StatusLoggedIn, "LoggedIn"},
		{StatusLoggedOut, "LoggedOut"},
		{StatusWrongPassword, "WrongPassword"},
		{StatusMaxLimit, "MaxLimit"},
		{StatusLogInAgain, "LogInAgain"},
		{StatusDataLimit, "DataLimit"},
		{StatusNoConnection, "NoConnection"},
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			got := tc.status.String()
			if got != tc.expected {
				t.Errorf("LoginStatus(%d).String() = %q, want %q", int(tc.status), got, tc.expected)
			}
		})
	}
}

func TestLoginStatusStringUnknown(t *testing.T) {
	unknown := LoginStatus(999)
	got := unknown.String()
	if got != "Unknown" {
		t.Errorf("LoginStatus(999).String() = %q, want %q", got, "Unknown")
	}
}

func TestLoginStatusIota(t *testing.T) {
	// Verify the iota ordering matches the C# enum order exactly.
	// Any change in ordering would be a breaking change for serialized data.
	if StatusLoggedIn != 0 {
		t.Errorf("StatusLoggedIn must be 0, got %d", int(StatusLoggedIn))
	}
	if StatusLoggedOut != 1 {
		t.Errorf("StatusLoggedOut must be 1, got %d", int(StatusLoggedOut))
	}
	if StatusWrongPassword != 2 {
		t.Errorf("StatusWrongPassword must be 2, got %d", int(StatusWrongPassword))
	}
	if StatusMaxLimit != 3 {
		t.Errorf("StatusMaxLimit must be 3, got %d", int(StatusMaxLimit))
	}
	if StatusLogInAgain != 4 {
		t.Errorf("StatusLogInAgain must be 4, got %d", int(StatusLogInAgain))
	}
	if StatusDataLimit != 5 {
		t.Errorf("StatusDataLimit must be 5, got %d", int(StatusDataLimit))
	}
	if StatusNoConnection != 6 {
		t.Errorf("StatusNoConnection must be 6, got %d", int(StatusNoConnection))
	}
}
