package tools

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMCPProtocol_RemoteWithoutAPIKeyRejected(t *testing.T) {
	s := NewMCPServer(MCPServerConfig{}, NewRegistry())
	req := httptest.NewRequest("POST", "/mcp", strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"initialize"}`))
	req.RemoteAddr = "203.0.113.10:1234"
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	s.ServeHTTP(rec, req)
	if !strings.Contains(rec.Body.String(), "APIKey") || strings.Contains(rec.Body.String(), "\"result\"") {
		t.Fatalf("expected remote request rejection, code=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestMCPProtocol_LoopbackWithoutAPIKeyAllowed(t *testing.T) {
	s := NewMCPServer(MCPServerConfig{}, NewRegistry())
	req := newLoopbackRequest("POST", "/mcp", strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"initialize"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	s.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("loopback initialize failed: code=%d body=%s", rec.Code, rec.Body.String())
	}
}
