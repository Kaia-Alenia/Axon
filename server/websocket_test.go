package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"github.com/gorilla/websocket"
)

func TestWebSocketOrigin(t *testing.T) {
	// Set up the token to bypass token checks and focus on Origin checks
	activeToken = "test_token"
	
	// Create a test server with our websocket handler
	server := httptest.NewServer(http.HandlerFunc(handleWebSocket))
	defer server.Close()

	// Replace http:// with ws://
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/?token=test_token"

	// Test 1: Malicious Origin header should fail
	dialer := websocket.Dialer{}
	headers := http.Header{}
	headers.Add("Origin", "http://evil.com")
	
	_, resp, err := dialer.Dial(wsURL, headers)
	if err == nil {
		t.Fatalf("Expected connection to fail with malicious Origin, but it succeeded")
	}
	if resp != nil && resp.StatusCode != http.StatusForbidden {
		t.Errorf("Expected 403 Forbidden, got %d", resp.StatusCode)
	}

	// Test 2: Missing Origin header should succeed (Simulating Android app / Non-browser client)
	// We need to bypass the IP check though. The `httptest` server makes connections from 127.0.0.1, 
	// which `isAddressLocal` treats as local, bypassing the token check actually, 
	// but providing the token is fine anyway.
	dialer2 := websocket.Dialer{}
	// No origin header
	conn, resp2, err2 := dialer2.Dial(wsURL, nil)
	if err2 != nil {
		t.Fatalf("Expected connection without Origin to succeed, got error: %v", err2)
	}
	if resp2 != nil && resp2.StatusCode != http.StatusSwitchingProtocols {
		t.Errorf("Expected 101 Switching Protocols, got %d", resp2.StatusCode)
	}
	conn.Close()
	
	// Test 3: Matching Origin header should succeed
	dialer3 := websocket.Dialer{}
	headers3 := http.Header{}
	headers3.Add("Origin", server.URL)
	conn3, resp3, err3 := dialer3.Dial(wsURL, headers3)
	if err3 != nil {
		t.Fatalf("Expected connection with matching Origin to succeed, got error: %v", err3)
	}
	if resp3 != nil && resp3.StatusCode != http.StatusSwitchingProtocols {
		t.Errorf("Expected 101 Switching Protocols, got %d", resp3.StatusCode)
	}
	conn3.Close()
}
