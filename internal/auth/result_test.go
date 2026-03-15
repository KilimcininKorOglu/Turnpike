package auth

import "testing"

func TestNewLoginResult(t *testing.T) {
	t.Run("success with response data", func(t *testing.T) {
		r := NewLoginResult(true, StatusLoggedIn, "Login successful", "<xml>ok</xml>")
		if !r.Success {
			t.Error("Success must be true")
		}
		if r.Status != StatusLoggedIn {
			t.Errorf("Status = %v, want %v", r.Status, StatusLoggedIn)
		}
		if r.Message != "Login successful" {
			t.Errorf("Message = %q, want %q", r.Message, "Login successful")
		}
		if r.ResponseData != "<xml>ok</xml>" {
			t.Errorf("ResponseData = %q, want %q", r.ResponseData, "<xml>ok</xml>")
		}
	})

	t.Run("failure with response data", func(t *testing.T) {
		r := NewLoginResult(false, StatusWrongPassword, "Invalid credentials", "<xml>err</xml>")
		if r.Success {
			t.Error("Success must be false")
		}
		if r.Status != StatusWrongPassword {
			t.Errorf("Status = %v, want %v", r.Status, StatusWrongPassword)
		}
		if r.Message != "Invalid credentials" {
			t.Errorf("Message = %q, want %q", r.Message, "Invalid credentials")
		}
		if r.ResponseData != "<xml>err</xml>" {
			t.Errorf("ResponseData = %q, want %q", r.ResponseData, "<xml>err</xml>")
		}
	})

	t.Run("empty response data is preserved", func(t *testing.T) {
		r := NewLoginResult(false, StatusNoConnection, "No connection", "")
		if r.ResponseData != "" {
			t.Errorf("ResponseData = %q, want empty string", r.ResponseData)
		}
	})
}

func TestCreateSuccess(t *testing.T) {
	t.Run("sets Success to true", func(t *testing.T) {
		r := CreateSuccess(StatusLoggedIn, "Authenticated")
		if !r.Success {
			t.Error("Success must be true")
		}
	})

	t.Run("sets Status correctly", func(t *testing.T) {
		r := CreateSuccess(StatusLoggedIn, "Authenticated")
		if r.Status != StatusLoggedIn {
			t.Errorf("Status = %v, want %v", r.Status, StatusLoggedIn)
		}
	})

	t.Run("sets Message correctly", func(t *testing.T) {
		r := CreateSuccess(StatusLoggedIn, "Authenticated")
		if r.Message != "Authenticated" {
			t.Errorf("Message = %q, want %q", r.Message, "Authenticated")
		}
	})

	t.Run("ResponseData is empty", func(t *testing.T) {
		r := CreateSuccess(StatusLoggedIn, "Authenticated")
		if r.ResponseData != "" {
			t.Errorf("ResponseData = %q, want empty string", r.ResponseData)
		}
	})

	t.Run("empty message is accepted", func(t *testing.T) {
		r := CreateSuccess(StatusLoggedIn, "")
		if r.Message != "" {
			t.Errorf("Message = %q, want empty string", r.Message)
		}
	})

	t.Run("uses correct status for each success variant", func(t *testing.T) {
		successStatuses := []LoginStatus{StatusLoggedIn, StatusLoggedOut, StatusLogInAgain}
		for _, s := range successStatuses {
			r := CreateSuccess(s, "")
			if r.Status != s {
				t.Errorf("CreateSuccess(%v).Status = %v, want %v", s, r.Status, s)
			}
			if !r.Success {
				t.Errorf("CreateSuccess(%v).Success must be true", s)
			}
		}
	})
}

func TestCreateFailure(t *testing.T) {
	t.Run("sets Success to false", func(t *testing.T) {
		r := CreateFailure(StatusWrongPassword, "Wrong password")
		if r.Success {
			t.Error("Success must be false")
		}
	})

	t.Run("sets Status correctly", func(t *testing.T) {
		r := CreateFailure(StatusMaxLimit, "Max session limit reached")
		if r.Status != StatusMaxLimit {
			t.Errorf("Status = %v, want %v", r.Status, StatusMaxLimit)
		}
	})

	t.Run("sets Message correctly", func(t *testing.T) {
		r := CreateFailure(StatusNoConnection, "Server unreachable")
		if r.Message != "Server unreachable" {
			t.Errorf("Message = %q, want %q", r.Message, "Server unreachable")
		}
	})

	t.Run("ResponseData is empty", func(t *testing.T) {
		r := CreateFailure(StatusWrongPassword, "Wrong password")
		if r.ResponseData != "" {
			t.Errorf("ResponseData = %q, want empty string", r.ResponseData)
		}
	})

	t.Run("uses correct status for each failure variant", func(t *testing.T) {
		failureStatuses := []LoginStatus{
			StatusWrongPassword,
			StatusMaxLimit,
			StatusLogInAgain,
			StatusDataLimit,
			StatusNoConnection,
		}
		for _, s := range failureStatuses {
			r := CreateFailure(s, "error")
			if r.Status != s {
				t.Errorf("CreateFailure(%v).Status = %v, want %v", s, r.Status, s)
			}
			if r.Success {
				t.Errorf("CreateFailure(%v).Success must be false", s)
			}
		}
	})
}

func TestLoginResultZeroValue(t *testing.T) {
	// The zero value of LoginResult should be a safe, unambiguous state.
	var r LoginResult
	if r.Success {
		t.Error("zero-value LoginResult.Success must be false")
	}
	if r.Status != StatusLoggedIn {
		t.Errorf("zero-value LoginResult.Status = %v, want %v (iota 0)", r.Status, StatusLoggedIn)
	}
	if r.Message != "" {
		t.Errorf("zero-value LoginResult.Message = %q, want empty string", r.Message)
	}
	if r.ResponseData != "" {
		t.Errorf("zero-value LoginResult.ResponseData = %q, want empty string", r.ResponseData)
	}
}
