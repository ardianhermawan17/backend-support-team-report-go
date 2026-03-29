package auth

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"backend-sport-team-report-go/internal/config"
	"backend-sport-team-report-go/tests/integration/testenv"
)

func TestAuthLoginRateLimit(t *testing.T) {
	env := testenv.StartPostgres(t)
	conn := env.OpenConnection(t)
	router := newAuthRouterWithSecurity(t, conn, config.SecurityConfig{
		RateLimit: config.RateLimitConfig{
			Login: config.RateLimitRule{
				Window:      time.Minute,
				MaxRequests: 2,
			},
		},
	})

	for attempt := 1; attempt <= 2; attempt++ {
		response := performAuthRequest(router, http.MethodPost, "/api/v1/auth/login", `{"username":"missing-user","password":"wrong-password"}`, "198.51.100.10:3456")
		if response.Code != http.StatusUnauthorized {
			t.Fatalf("attempt %d: expected unauthorized status %d, got %d with body %s", attempt, http.StatusUnauthorized, response.Code, response.Body.String())
		}
	}

	throttled := performAuthRequest(router, http.MethodPost, "/api/v1/auth/login", `{"username":"missing-user","password":"wrong-password"}`, "198.51.100.10:3456")
	if throttled.Code != http.StatusTooManyRequests {
		t.Fatalf("expected throttled status %d, got %d with body %s", http.StatusTooManyRequests, throttled.Code, throttled.Body.String())
	}
	if throttled.Header().Get("Retry-After") == "" {
		t.Fatal("expected Retry-After header when login is throttled")
	}
}

func performAuthRequest(router http.Handler, method, path, body, remoteAddr string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = remoteAddr

	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	return res
}
