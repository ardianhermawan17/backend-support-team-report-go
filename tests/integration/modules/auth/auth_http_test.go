package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	ginrouter "backend-sport-team-report-go/internal/api/gin/router"
	"backend-sport-team-report-go/internal/config"
	"backend-sport-team-report-go/internal/modules/auth/domain/entities"
	authpersistence "backend-sport-team-report-go/internal/modules/auth/infrastructure/persistence"
	"backend-sport-team-report-go/internal/platform/database/postgres"
	"backend-sport-team-report-go/internal/shared/logger"
	appcrypto "backend-sport-team-report-go/pkg/crypto"
	"backend-sport-team-report-go/tests/integration/testenv"
	"backend-sport-team-report-go/tests/integration/testhelpers"
)

func TestAuthLoginAndMeFlow(t *testing.T) {
	env := testenv.StartPostgres(t)
	conn := env.OpenConnection(t)
	repo := authpersistence.NewAccountRepository(conn)
	router := newAuthRouter(t, conn)

	password := "correct-horse-battery"
	hash, err := appcrypto.HashPassword(password)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	account := entities.CompanyAdminAccount{
		User: entities.User{
			ID:           7100000000101,
			Username:     "admin-login",
			Email:        "admin-login@example.test",
			PasswordHash: hash,
		},
		Company: entities.Company{
			ID:   7200000000101,
			Name: "Login FC",
		},
	}

	if err := repo.Create(context.Background(), account); err != nil {
		t.Fatalf("create account: %v", err)
	}

	loginBody := bytes.NewBufferString(`{"username":"admin-login","password":"correct-horse-battery"}`)
	loginRequest := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", loginBody)
	loginRequest.Header.Set("Content-Type", "application/json")
	loginResponse := httptest.NewRecorder()

	router.ServeHTTP(loginResponse, loginRequest)

	if loginResponse.Code != http.StatusOK {
		t.Fatalf("expected login status %d, got %d with body %s", http.StatusOK, loginResponse.Code, loginResponse.Body.String())
	}

	var loginPayload struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		User        struct {
			ID       int64  `json:"id"`
			Username string `json:"username"`
			Email    string `json:"email"`
		} `json:"user"`
		Company struct {
			ID   int64  `json:"id"`
			Name string `json:"name"`
		} `json:"company"`
	}
	if err := json.Unmarshal(loginResponse.Body.Bytes(), &loginPayload); err != nil {
		t.Fatalf("unmarshal login response: %v", err)
	}

	if loginPayload.AccessToken == "" {
		t.Fatal("expected access token in login response")
	}

	if loginPayload.TokenType != "Bearer" {
		t.Fatalf("expected Bearer token type, got %q", loginPayload.TokenType)
	}

	if loginPayload.User.ID != account.User.ID || loginPayload.User.Username != account.User.Username || loginPayload.User.Email != account.User.Email {
		t.Fatalf("unexpected user in login response: %#v", loginPayload.User)
	}

	if loginPayload.Company.ID != account.Company.ID || loginPayload.Company.Name != account.Company.Name {
		t.Fatalf("unexpected company in login response: %#v", loginPayload.Company)
	}

	meRequest := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	meRequest.Header.Set("Authorization", "Bearer "+loginPayload.AccessToken)
	meResponse := httptest.NewRecorder()

	router.ServeHTTP(meResponse, meRequest)

	if meResponse.Code != http.StatusOK {
		t.Fatalf("expected me status %d, got %d with body %s", http.StatusOK, meResponse.Code, meResponse.Body.String())
	}

	var mePayload struct {
		User struct {
			ID       int64  `json:"id"`
			Username string `json:"username"`
			Email    string `json:"email"`
		} `json:"user"`
	}
	if err := json.Unmarshal(meResponse.Body.Bytes(), &mePayload); err != nil {
		t.Fatalf("unmarshal me response: %v", err)
	}
	if mePayload.User.Email != account.User.Email {
		t.Fatalf("expected me email %q, got %q", account.User.Email, mePayload.User.Email)
	}
}

func TestAuthLoginRejectsInvalidCredentials(t *testing.T) {
	env := testenv.StartPostgres(t)
	conn := env.OpenConnection(t)
	repo := authpersistence.NewAccountRepository(conn)
	router := newAuthRouter(t, conn)

	hash, err := appcrypto.HashPassword("correct-password")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	account := entities.CompanyAdminAccount{
		User: entities.User{
			ID:           7100000000102,
			Username:     "admin-invalid",
			Email:        "admin-invalid@example.test",
			PasswordHash: hash,
		},
		Company: entities.Company{
			ID:   7200000000102,
			Name: "Invalid FC",
		},
	}

	if err := repo.Create(context.Background(), account); err != nil {
		t.Fatalf("create account: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(`{"username":"admin-invalid","password":"wrong-password"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized status %d, got %d with body %s", http.StatusUnauthorized, response.Code, response.Body.String())
	}
}

func TestAuthMeRejectsSoftDeletedAccount(t *testing.T) {
	env := testenv.StartPostgres(t)
	conn := env.OpenConnection(t)
	repo := authpersistence.NewAccountRepository(conn)
	router := newAuthRouter(t, conn)

	hash, err := appcrypto.HashPassword("active-password")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	account := entities.CompanyAdminAccount{
		User: entities.User{
			ID:           7100000000103,
			Username:     "admin-soft-delete-login",
			Email:        "admin-soft-delete-login@example.test",
			PasswordHash: hash,
		},
		Company: entities.Company{
			ID:   7200000000103,
			Name: "Soft Delete FC",
		},
	}

	if err := repo.Create(context.Background(), account); err != nil {
		t.Fatalf("create account: %v", err)
	}

	loginRequest := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(`{"username":"admin-soft-delete-login","password":"active-password"}`))
	loginRequest.Header.Set("Content-Type", "application/json")
	loginResponse := httptest.NewRecorder()
	router.ServeHTTP(loginResponse, loginRequest)

	if loginResponse.Code != http.StatusOK {
		t.Fatalf("expected login status %d, got %d with body %s", http.StatusOK, loginResponse.Code, loginResponse.Body.String())
	}

	var loginPayload struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(loginResponse.Body.Bytes(), &loginPayload); err != nil {
		t.Fatalf("unmarshal login response: %v", err)
	}

	if _, err := env.DB.Exec(`UPDATE users SET deleted_at = NOW() WHERE id = $1`, account.User.ID); err != nil {
		t.Fatalf("soft delete user: %v", err)
	}

	meRequest := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	meRequest.Header.Set("Authorization", "Bearer "+loginPayload.AccessToken)
	meResponse := httptest.NewRecorder()

	router.ServeHTTP(meResponse, meRequest)

	if meResponse.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized status %d, got %d with body %s", http.StatusUnauthorized, meResponse.Code, meResponse.Body.String())
	}
}

func TestAuthLoginRejectsUnknownJSONFields(t *testing.T) {
	env := testenv.StartPostgres(t)
	conn := env.OpenConnection(t)
	router := newAuthRouter(t, conn)

	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(`{"username":"admin-login","password":"secret","role":"owner"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request status %d, got %d with body %s", http.StatusBadRequest, response.Code, response.Body.String())
	}
}

func TestAuthLoginRateLimitRejectsRepeatedRequests(t *testing.T) {
	env := testenv.StartPostgres(t)
	conn := env.OpenConnection(t)
	repo := authpersistence.NewAccountRepository(conn)
	security := testhelpers.DefaultTestConfig().Security
	security.RateLimit.Login.Window = time.Minute
	security.RateLimit.Login.MaxRequests = 2
	router := newAuthRouterWithSecurity(t, conn, security)

	hash, err := appcrypto.HashPassword("correct-password")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	account := entities.CompanyAdminAccount{
		User: entities.User{
			ID:           7100000000104,
			Username:     "admin-throttle",
			Email:        "admin-throttle@example.test",
			PasswordHash: hash,
		},
		Company: entities.Company{
			ID:   7200000000104,
			Name: "Throttle FC",
		},
	}

	if err := repo.Create(context.Background(), account); err != nil {
		t.Fatalf("create account: %v", err)
	}

	for attempt := 1; attempt <= 3; attempt++ {
		request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(`{"username":"admin-throttle","password":"wrong-password"}`))
		request.Header.Set("Content-Type", "application/json")
		request.RemoteAddr = "198.51.100.20:1234"
		response := httptest.NewRecorder()

		router.ServeHTTP(response, request)

		switch attempt {
		case 1, 2:
			if response.Code != http.StatusUnauthorized {
				t.Fatalf("expected unauthorized status %d on attempt %d, got %d with body %s", http.StatusUnauthorized, attempt, response.Code, response.Body.String())
			}
		case 3:
			if response.Code != http.StatusTooManyRequests {
				t.Fatalf("expected too many requests status %d on attempt %d, got %d with body %s", http.StatusTooManyRequests, attempt, response.Code, response.Body.String())
			}
			if response.Header().Get("Retry-After") == "" {
				t.Fatal("expected Retry-After header on throttled login response")
			}
		}
	}
}

func newAuthRouter(t *testing.T, conn *postgres.Connection) http.Handler {
	t.Helper()
	return newAuthRouterWithSecurity(t, conn, testhelpers.DefaultTestConfig().Security)
}

func newAuthRouterWithSecurity(t *testing.T, conn *postgres.Connection, security config.SecurityConfig) http.Handler {
	t.Helper()

	cfg := testhelpers.DefaultTestConfig()
	cfg.Security = security
	return ginrouter.New(cfg, conn, logger.New(cfg.App.Name, cfg.App.Env))
}
