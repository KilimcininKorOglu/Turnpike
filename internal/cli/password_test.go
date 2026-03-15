package cli

import (
	"os"
	"testing"
)

func TestPromptPassword_PipedInput(t *testing.T) {
	// Create a pipe to simulate stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	// Write password to pipe
	_, err = w.WriteString("testpassword\n")
	if err != nil {
		t.Fatal(err)
	}
	w.Close()

	// Replace stdin temporarily
	origStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = origStdin }()

	password, err := PromptPassword()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if password != "testpassword" {
		t.Errorf("expected 'testpassword', got %q", password)
	}
}

func TestPromptPassword_PipedInputWithCRLF(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	_, err = w.WriteString("mypass\r\n")
	if err != nil {
		t.Fatal(err)
	}
	w.Close()

	origStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = origStdin }()

	password, err := PromptPassword()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if password != "mypass" {
		t.Errorf("expected 'mypass', got %q", password)
	}
}
