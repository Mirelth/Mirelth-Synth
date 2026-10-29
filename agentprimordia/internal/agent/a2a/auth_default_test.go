package a2a

import (
	"net/http"
	"testing"
)

func TestLoopbackOnlyAuthenticator_DefaultPolicy(t *testing.T) {
	a := NewLoopbackOnlyAuthenticator()

	loop := &http.Request{RemoteAddr: "127.0.0.1:1234"}
	if _, err := a.Authenticate(loop); err != nil {
		t.Fatalf("loopback should pass: %v", err)
	}
	v6 := &http.Request{RemoteAddr: "[::1]:1234"}
	if _, err := a.Authenticate(v6); err != nil {
		t.Fatalf("IPv6 loopback should pass: %v", err)
	}
	remote := &http.Request{RemoteAddr: "203.0.113.10:1234"}
	if _, err := a.Authenticate(remote); err == nil {
		t.Fatal("non-loopback must be rejected by default")
	}
}

func TestNoopAuthenticator_ExplicitOptOutAllowsRemote(t *testing.T) {
	a := NewNoopAuthenticator()
	if _, err := a.Authenticate(&http.Request{RemoteAddr: "203.0.113.10:1234"}); err != nil {
		t.Fatalf("explicit Noop should allow remote: %v", err)
	}
}
