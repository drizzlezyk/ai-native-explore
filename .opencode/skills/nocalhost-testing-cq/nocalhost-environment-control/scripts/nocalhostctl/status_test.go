//go:build debug

package main

import "testing"

func TestResolveStatusSnapshot(t *testing.T) {
	tests := []struct {
		name          string
		hasConfig     bool
		hasState      bool
		podRunning    bool
		serverRunning bool
		wantState     string
		wantNextHint  string
	}{
		{
			name:         "missing config",
			hasConfig:    false,
			wantState:    "not_prepared",
			wantNextHint: "prepare",
		},
		{
			name:         "config only",
			hasConfig:    true,
			hasState:     false,
			wantState:    "uninstalled",
			wantNextHint: "oneclickstart",
		},
		{
			name:         "pod missing",
			hasConfig:    true,
			hasState:     true,
			podRunning:   false,
			wantState:    "uninstalled",
			wantNextHint: "oneclickstart",
		},
		{
			name:          "server running",
			hasConfig:     true,
			hasState:      true,
			podRunning:    true,
			serverRunning: true,
			wantState:     "server_running",
			wantNextHint:  "rebuild",
		},
		{
			name:          "pod running only",
			hasConfig:     true,
			hasState:      true,
			podRunning:    true,
			serverRunning: false,
			wantState:     "pod_running",
			wantNextHint:  "rebuild --sync-vendor",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state, nextHint := resolveStatusSnapshot(tt.hasConfig, tt.hasState, tt.podRunning, tt.serverRunning)
			if state != tt.wantState {
				t.Fatalf("state: got %q, want %q", state, tt.wantState)
			}
			if nextHint != tt.wantNextHint {
				t.Fatalf("nextHint: got %q, want %q", nextHint, tt.wantNextHint)
			}
		})
	}
}
