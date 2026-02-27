package integration_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/mhtecdev/blog-ai/tests/testutil"
)

func TestSecurityHeaders(t *testing.T) {
	app := testutil.NewTestApp(t)
	resp := app.Get("/")
	if resp.StatusCode == 500 {
		t.Fatal("home page returned 500")
	}

	h := resp.Header
	assertHeader(t, h, "X-Content-Type-Options", "nosniff")
	assertHeader(t, h, "X-Frame-Options", "DENY")
	assertHeaderContains(t, h, "Referrer-Policy", "strict-origin")
	if h.Get("Content-Security-Policy") == "" {
		t.Error("Content-Security-Policy header is missing")
	}
}

func TestRateLimitOnLogin(t *testing.T) {
	app := testutil.NewTestApp(t)

	// First N attempts (limit is 3) should not be rate-limited (may be 401 for bad creds)
	for i := 0; i < 3; i++ {
		resp := app.PostForm("/studio/login", map[string]string{
			"username": "wrong",
			"password": "wrong",
		}, nil)
		if resp.StatusCode == http.StatusTooManyRequests {
			t.Errorf("attempt %d was rate limited before limit reached", i+1)
		}
	}

	// Next attempt (4th) should be rate limited
	resp := app.PostForm("/studio/login", map[string]string{
		"username": "wrong",
		"password": "wrong",
	}, nil)
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Errorf("expected 429 after limit, got %d", resp.StatusCode)
	}
	if resp.Header.Get("Retry-After") == "" {
		t.Error("Retry-After header missing on 429 response")
	}
}

func TestCSPPresent(t *testing.T) {
	app := testutil.NewTestApp(t)
	for _, path := range []string{"/", "/timeline", "/categories"} {
		resp := app.Get(path)
		if resp.Header.Get("Content-Security-Policy") == "" {
			t.Errorf("CSP header missing on %s", path)
		}
	}
}

func TestUnauthenticatedRedirect(t *testing.T) {
	app := testutil.NewTestApp(t)
	for _, path := range []string{"/studio/dashboard", "/studio/posts", "/studio/metrics"} {
		resp := app.Get(path)
		if resp.StatusCode != http.StatusSeeOther {
			t.Errorf("expected redirect on %s, got %d", path, resp.StatusCode)
		}
		loc := resp.Header.Get("Location")
		if !strings.Contains(loc, "/studio/login") {
			t.Errorf("expected redirect to /studio/login on %s, got %s", path, loc)
		}
	}
}

func assertHeader(t *testing.T, h http.Header, key, expected string) {
	t.Helper()
	if got := h.Get(key); got != expected {
		t.Errorf("header %s: expected %q, got %q", key, expected, got)
	}
}

func assertHeaderContains(t *testing.T, h http.Header, key, substr string) {
	t.Helper()
	if got := h.Get(key); !strings.Contains(got, substr) {
		t.Errorf("header %s: expected to contain %q, got %q", key, substr, got)
	}
}
