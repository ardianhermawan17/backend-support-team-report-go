package teams

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	ginrouter "backend-sport-team-report-go/internal/api/gin/router"
	"backend-sport-team-report-go/internal/config"
	authpersistence "backend-sport-team-report-go/internal/modules/auth/infrastructure/persistence"
	"backend-sport-team-report-go/internal/platform/database/postgres"
	"backend-sport-team-report-go/internal/shared/logger"
	"backend-sport-team-report-go/tests/integration/testhelpers"
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
	return testhelpers.CreateAccountAndLogin(t, repo, router, userID, companyID, username, companyName, password)
}

func sendJSONRequest(t *testing.T, router http.Handler, method, path, token string, payload map[string]any) *httptest.ResponseRecorder {
	t.Helper()
	return testhelpers.SendJSONRequest(t, router, method, path, token, payload)
}

func sendRequest(t *testing.T, router http.Handler, method, path, token string, body *bytes.Reader) *httptest.ResponseRecorder {
	t.Helper()
	return testhelpers.SendRequest(t, router, method, path, token, body)
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
