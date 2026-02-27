package integration_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/mhtecdev/blog-ai/tests/testutil"
)

func TestLoginWithValidCredentials(t *testing.T) {
	app := testutil.NewTestApp(t)
	app.SeedUser(t, "admin", "supersecret")

	// Note: rate limit is 3 per window; SeedUser already does a login internally.
	// We use a fresh form login here.
	resp := app.PostForm("/studio/login", map[string]string{
		"username": "admin",
		"password": "supersecret",
	}, nil)

	if resp.StatusCode != http.StatusSeeOther {
		t.Errorf("expected 303 redirect, got %d", resp.StatusCode)
	}
	loc := resp.Header.Get("Location")
	if !strings.Contains(loc, "/studio/dashboard") {
		t.Errorf("expected redirect to dashboard, got %s", loc)
	}
	var hasSessionCookie bool
	for _, c := range resp.Cookies() {
		if c.Name == "session_id" && c.Value != "" {
			hasSessionCookie = true
		}
	}
	if !hasSessionCookie {
		t.Error("no session_id cookie set after login")
	}
}

func TestLoginWithInvalidCredentials(t *testing.T) {
	app := testutil.NewTestApp(t)
	resp := app.PostForm("/studio/login", map[string]string{
		"username": "nobody",
		"password": "wrongpass",
	}, nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 for bad credentials, got %d", resp.StatusCode)
	}
}

func TestLogout(t *testing.T) {
	app := testutil.NewTestApp(t)
	cookie := app.SeedUser(t, "admin2", "password123")

	resp := app.PostForm("/studio/logout", nil, []*http.Cookie{cookie})
	if resp.StatusCode != http.StatusSeeOther {
		t.Errorf("expected 303 after logout, got %d", resp.StatusCode)
	}

	// Now the old session should be invalid
	resp2 := app.Do("GET", "/studio/dashboard", nil, map[string]string{
		"Cookie": "session_id=" + cookie.Value,
	})
	if resp2.StatusCode != http.StatusSeeOther {
		t.Errorf("expected redirect after logout cookie used, got %d", resp2.StatusCode)
	}
}

func TestProtectedRouteWithValidSession(t *testing.T) {
	app := testutil.NewTestApp(t)
	cookie := app.SeedUser(t, "adminx", "secret999")

	req := map[string]string{}
	resp := app.PostForm("/studio/dashboard", req, []*http.Cookie{cookie})
	// /studio/dashboard is a GET, but we test via Do
	_ = resp

	resp2 := app.Do("GET", "/studio/dashboard", nil, map[string]string{
		"Cookie": "session_id=" + cookie.Value,
	})
	// Should get 200 or rendered page, NOT a redirect
	if resp2.StatusCode == http.StatusSeeOther {
		t.Error("authenticated GET /studio/dashboard should not redirect")
	}
}
