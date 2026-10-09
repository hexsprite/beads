package main

import (
	"strings"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/types"
)

// TestCreateBuildersResolveInitialStatus pins the precedence both single-issue
// builders share: --status flag, then a future --defer, then status.default,
// then open. ConfigDefault marks the single-issue callers; batch callers leave
// it false and start open.
func TestCreateBuildersResolveInitialStatus(t *testing.T) {
	future := time.Now().Add(time.Hour)
	adapters := []struct {
		name  string
		build func(flag string, deferUntil *time.Time, resolved string) *types.Issue
	}{
		{"buildCreateIssue", func(flag string, deferUntil *time.Time, resolved string) *types.Issue {
			return buildCreateIssue(createIssueParams{
				Title: "t", InitialStatus: flag, DeferUntil: deferUntil, ConfigDefault: true, ResolvedDefault: resolved,
			})
		}},
		{"buildCreateIssueFromInput", func(flag string, deferUntil *time.Time, resolved string) *types.Issue {
			return buildCreateIssueFromInput(createInput{
				title: "t", status: flag, deferUntil: deferUntil, previewStatus: resolved,
			})
		}},
	}
	cases := []struct {
		name       string
		flag       string
		deferUntil *time.Time
		resolved   string
		want       types.Status
	}{
		{name: "flag beats defer and default", flag: "blocked", deferUntil: &future, resolved: "backlog", want: types.StatusBlocked},
		{name: "defer beats default", deferUntil: &future, resolved: "backlog", want: types.StatusDeferred},
		{name: "default fills the preview", resolved: "backlog", want: "backlog"},
		{name: "real create leaves the status to the role", want: ""},
	}
	for _, ad := range adapters {
		for _, tc := range cases {
			t.Run(ad.name+"/"+tc.name, func(t *testing.T) {
				if got := ad.build(tc.flag, tc.deferUntil, tc.resolved).Status; got != tc.want {
					t.Fatalf("status = %q, want %q", got, tc.want)
				}
			})
		}
	}
}

func TestBuildCreateIssueBatchCallersStartOpen(t *testing.T) {
	issue := buildCreateIssue(createIssueParams{Title: "t", ResolvedDefault: "backlog"})
	if issue.Status != types.StatusOpen {
		t.Fatalf("status = %q, want open: batch builders never consult status.default", issue.Status)
	}
}

func TestPreviewDefaultStatus(t *testing.T) {
	future := time.Now().Add(time.Hour)
	past := time.Now().Add(-time.Hour)
	custom := []string{"backlog"}
	cases := []struct {
		name       string
		flag       string
		deferUntil *time.Time
		configured string
		want       string
		wantErr    string
	}{
		{name: "custom default", configured: "backlog", want: "backlog"},
		{name: "built-in default", configured: " blocked ", want: "blocked"},
		{name: "unset is open", want: "open"},
		{name: "invalid default is refused", configured: "nonsense", wantErr: `invalid status.default "nonsense"`},
		{name: "flag skips an invalid default", flag: "open", configured: "nonsense", want: ""},
		{name: "future defer skips an invalid default", deferUntil: &future, configured: "nonsense", want: ""},
		{name: "past defer does not skip", deferUntil: &past, configured: "backlog", want: "backlog"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := previewDefaultStatus(tc.flag, tc.deferUntil, tc.configured, custom)
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("error = %v, want containing %q", err, tc.wantErr)
				}
				return
			}
			if err != nil || got != tc.want {
				t.Fatalf("previewDefaultStatus() = %q, %v; want %q", got, err, tc.want)
			}
		})
	}
}
