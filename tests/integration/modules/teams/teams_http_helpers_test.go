package teams

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
)

func newTeamsRouter(conn *postgres.Connection) http.Handler {
	cfg := config.Config{
		App: config.AppConfig{
			Name: "soccer-team-report",
			Env:  config.EnvTest,
		},
		Database: config.DatabaseConfig{},
		Auth: config.AuthConfig{
			JWTSecret:      "integration-test-secret",
			AccessTokenTTL: 15 * time.Minute,
		},
	}

	return ginrouter.New(cfg, conn, logger.New(cfg.App.Name, cfg.App.Env))
}

func createAccountAndLogin(t *testing.T, repo *authpersistence.AccountRepository, router http.Handler, userID, companyID int64, username, companyName, password string) string {
	t.Helper()

	hash, err := appcrypto.HashPassword(password)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	account := entities.CompanyAdminAccount{
		User: entities.User{
			ID:           userID,
			Username:     username,
			PasswordHash: hash,
		},
		Company: entities.Company{
			ID:   companyID,
			Name: companyName,
		},
	}

	if err := repo.Create(context.Background(), account); err != nil {
		t.Fatalf("create account: %v", err)
	}

	loginResponse := sendJSONRequest(t, router, http.MethodPost, "/api/v1/auth/login", "", map[string]any{
		"username": username,
		"password": password,
	})
	if loginResponse.Code != http.StatusOK {
		t.Fatalf("expected login status %d, got %d with body %s", http.StatusOK, loginResponse.Code, loginResponse.Body.String())
	}

	var payload struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(loginResponse.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal login response: %v", err)
	}
	if payload.AccessToken == "" {
		t.Fatal("expected access token from login")
	}

	return payload.AccessToken
}

func sendJSONRequest(t *testing.T, router http.Handler, method, path, token string, payload map[string]any) *httptest.ResponseRecorder {
	t.Helper()

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	return sendRequest(t, router, method, path, token, bytes.NewReader(body))
}

func sendRequest(t *testing.T, router http.Handler, method, path, token string, body *bytes.Reader) *httptest.ResponseRecorder {
	t.Helper()

	var requestBody *bytes.Reader
	if body == nil {
		requestBody = bytes.NewReader(nil)
	} else {
		requestBody = body
	}

	req := httptest.NewRequest(method, path, requestBody)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if method == http.MethodPost || method == http.MethodPut {
		req.Header.Set("Content-Type", "application/json")
	}
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	return res
}

type teamPayload struct {
	ID                    int64  `json:"id"`
	Name                  string `json:"name"`
	FoundedYear           int    `json:"founded_year"`
	HomebaseAddress       string `json:"homebase_address"`
	CityOfHomebaseAddress string `json:"city_of_homebase_address"`
}

func decodeTeamResponse(t *testing.T, response *httptest.ResponseRecorder) teamPayload {
	t.Helper()

	var team teamPayload
	if err := json.Unmarshal(response.Body.Bytes(), &team); err != nil {
		t.Fatalf("unmarshal team response: %v", err)
	}

	return team
}
