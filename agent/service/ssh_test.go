package service

import (
	"testing"
	"time"
)

func TestParseSSHLogDate(t *testing.T) {
	now := time.Date(2026, time.August, 24, 12, 0, 0, 0, time.FixedZone("CST", 8*60*60))
	tests := []struct {
		name string
		line string
		want string
	}{
		{
			name: "syslog",
			line: "Aug 24 11:22:33 host sshd[123]: Accepted password for user from 192.0.2.1 port 22 ssh2",
			want: "2026-08-24 11:22:33",
		},
		{
			name: "syslog with leading whitespace",
			line: "  Aug  4 02:03:04 host sshd[123]: Accepted publickey for user from 192.0.2.1 port 22 ssh2",
			want: "2026-08-04 02:03:04",
		},
		{
			name: "journalctl",
			line: "2026-08-24 11:22:33 host sshd[123]: Accepted password for user from 192.0.2.1 port 22 ssh2",
			want: "2026-08-24 11:22:33",
		},
		{
			name: "rfc3339",
			line: "2026-08-24T11:22:33.123456Z host sshd[123]: Accepted password for user from 192.0.2.1 port 22 ssh2",
			want: "2026-08-24 11:22:33",
		},
		{
			name: "prefixed rfc3339 offset",
			line: "node journal: 2026-08-24T11:22:33+08:00 sshd: Accepted password for user from 192.0.2.1 port 22 ssh2",
			want: "2026-08-24 11:22:33",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseSSHLogDate(tt.line, now); got != tt.want {
				t.Fatalf("parseSSHLogDate() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseSSHLogDateYearBoundary(t *testing.T) {
	now := time.Date(2026, time.January, 2, 12, 0, 0, 0, time.UTC)
	got := parseSSHLogDate("Dec 31 23:59:59 host sshd: Accepted password for user from 192.0.2.1 port 22 ssh2", now)
	if got != "2025-12-31 23:59:59" {
		t.Fatalf("parseSSHLogDate() = %q, want previous year", got)
	}
}

func TestParseSSHLogDateInvalid(t *testing.T) {
	if got := parseSSHLogDate("host sshd: Accepted password for user from 192.0.2.1 port 22 ssh2", time.Now()); got != "" {
		t.Fatalf("parseSSHLogDate() = %q, want empty result", got)
	}
}
