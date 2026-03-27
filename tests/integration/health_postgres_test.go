package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	ginrouter "backend-sport-team-report-go/internal/api/gin/router"
	"backend-sport-team-report-go/internal/config"
	"backend-sport-team-report-go/internal/shared/logger"
	"backend-sport-team-report-go/tests/integration/testenv"
)

func TestHealthEndpointWithRealPostgres(t *testing.T) {
	env := testenv.StartPostgres(t)
	conn := env.OpenConnection(t)
	cfg := config.Config{
		App: config.AppConfig{
			Name: "soccer-team-report",
			Env:  config.EnvTest,
		},
	}

	router := ginrouter.New(cfg, conn, logger.New(cfg.App.Name, cfg.App.Env))
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if body["database"] != "up" {
		t.Fatalf("expected database up, got %#v", body["database"])
	}
}
