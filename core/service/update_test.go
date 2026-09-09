package service

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUpdateServiceStatus(t *testing.T) {
	tests := []struct {
		name        string
		status      string
		wantState   string
		wantPhase   string
		wantTarget  string
		wantMessage string
		wantError   string
	}{
		{
			name:      "missing status is idle",
			wantState: "idle",
			wantPhase: "idle",
		},
		{
			name:        "completed remains historical success",
			status:      `{"phase":"completed","message":"更新并重启成功","target_version":"v1.1.2","updated_at":"2026-09-09T00:00:00Z"}`,
			wantState:   "success",
			wantPhase:   "completed",
			wantTarget:  "v1.1.2",
			wantMessage: "更新并重启成功",
		},
		{
			name:       "staged remains pending success",
			status:     `{"phase":"staged","message":"更新已准备完成，请重启应用","target_version":"v1.1.3"}`,
			wantState:  "success",
			wantPhase:  "staged",
			wantTarget: "v1.1.3",
		},
		{
			name:      "failed status is failed",
			status:    `{"phase":"failed","message":"更新失败","error":"download failed"}`,
			wantState: "failed",
			wantPhase: "failed",
			wantError: "download failed",
		},
		{
			name:      "malformed status is failed",
			status:    `{not-json`,
			wantState: "failed",
			wantPhase: "failed",
			wantError: "invalid update status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statusPath := filepath.Join(t.TempDir(), "status.json")
			t.Setenv("GPANEL_UPDATE_STATUS_FILE", statusPath)
			if tt.status != "" {
				if err := os.WriteFile(statusPath, []byte(tt.status), 0600); err != nil {
					t.Fatalf("write status: %v", err)
				}
			}

			status := (&UpdateService{}).Status()
			if status.State != tt.wantState {
				t.Errorf("State = %q, want %q", status.State, tt.wantState)
			}
			if status.Phase != tt.wantPhase {
				t.Errorf("Phase = %q, want %q", status.Phase, tt.wantPhase)
			}
			if status.TargetVersion != tt.wantTarget {
				t.Errorf("TargetVersion = %q, want %q", status.TargetVersion, tt.wantTarget)
			}
			if status.Message != tt.wantMessage {
				t.Errorf("Message = %q, want %q", status.Message, tt.wantMessage)
			}
			if status.Error != tt.wantError {
				t.Errorf("Error = %q, want %q", status.Error, tt.wantError)
			}
		})
	}
}
