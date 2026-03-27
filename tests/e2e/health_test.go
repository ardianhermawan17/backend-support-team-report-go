package e2e

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	ginrouter "backend-sport-team-report-go/internal/api/gin/router"
	"backend-sport-team-report-go/internal/config"
	"backend-sport-team-report-go/internal/shared/logger"
)

func TestHealthEndpoint(t *testing.T) {
	cfg := config.Config{App: config.AppConfig{Name: "soccer-team-report", Env: config.EnvTest}}
	log := logger.New(cfg.App.Name, cfg.App.Env)
	router := ginrouter.New(cfg, nil, log)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, res.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if body["database"] != "down" {
		t.Fatalf("expected database down, got %#v", body["database"])
	}
}
