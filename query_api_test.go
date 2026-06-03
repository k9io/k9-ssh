package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	// Redirect logging to discard so tests don't require a live syslog connection.
	log.SetOutput(io.Discard)
	os.Exit(m.Run())
}

func testConfig(serverURL string) *Configuration {
	c := &Configuration{}
	c.System.MachineGroup = "testgroup"
	c.System.RunAs = "key9"
	c.System.ConnectionTimeout = 5
	c.Authentication.APIKey = "test-key"
	c.Authentication.CompanyUUID = "test-uuid"
	c.Urls.QuerySSHKeys = serverURL + "/keys/"
	c.Urls.QueryAllUsers = serverURL + "/users"
	return c
}

// --- validUsername ---

func TestValidUsername(t *testing.T) {
	valid := []string{"alice", "bob123", "test_user", "_svc", "a", "user-name"}
	for _, u := range valid {
		if !validUsername.MatchString(u) {
			t.Errorf("expected %q to be valid", u)
		}
	}

	invalid := []string{
		"",
		"../etc/passwd",
		"foo/bar",
		"user?x=1",
		"user#hash",
		"User",                  // uppercase
		"123user",               // starts with digit
		strings.Repeat("a", 33), // too long
	}
	for _, u := range invalid {
		if validUsername.MatchString(u) {
			t.Errorf("expected %q to be invalid", u)
		}
	}
}

// --- QueryAPI ---

func TestQueryAPI_ValidKey(t *testing.T) {
	// A real Ed25519 authorized_keys line for testing.
	pubKey := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GkZE test@example.com"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"public_key":"%s"}`+"\n", pubKey)
	}))
	defer srv.Close()

	Config = testConfig(srv.URL)

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	QueryAPI("alice", "", true)

	w.Close()
	os.Stdout = old
	out, _ := io.ReadAll(r)

	if !strings.Contains(string(out), "ssh-ed25519") {
		t.Errorf("expected public key in stdout, got: %q", string(out))
	}
}

func TestQueryAPI_APIErrorResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"error":"invalid api key"}`)
	}))
	defer srv.Close()

	Config = testConfig(srv.URL)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	QueryAPI("alice", "", true)

	w.Close()
	os.Stdout = old
	out, _ := io.ReadAll(r)

	if strings.TrimSpace(string(out)) != "" {
		t.Errorf("expected no output on API error, got: %q", string(out))
	}
}

func TestQueryAPI_Non200Status(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, `{"error":"unauthorized"}`)
	}))
	defer srv.Close()

	Config = testConfig(srv.URL)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	QueryAPI("alice", "", true)

	w.Close()
	os.Stdout = old
	out, _ := io.ReadAll(r)

	if strings.TrimSpace(string(out)) != "" {
		t.Errorf("expected no output on non-200 status, got: %q", string(out))
	}
}

func TestQueryAPI_InvalidUsername(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))
	defer srv.Close()

	Config = testConfig(srv.URL)
	QueryAPI("../../etc/passwd", "", true)

	if called {
		t.Error("API should not be called for invalid username")
	}
}

func TestQueryAPI_MalformedPublicKey(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"public_key":"not-a-valid-key"}`)
	}))
	defer srv.Close()

	Config = testConfig(srv.URL)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	QueryAPI("alice", "", true)

	w.Close()
	os.Stdout = old
	out, _ := io.ReadAll(r)

	if strings.TrimSpace(string(out)) != "" {
		t.Errorf("expected malformed key to be rejected, got: %q", string(out))
	}
}
