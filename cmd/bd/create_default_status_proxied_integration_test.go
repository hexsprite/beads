//go:build cgo

package main

import (
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/types"
)

func TestProxiedServerCreateDefaultStatus(t *testing.T) {
	requireSharedProxiedServer(t)
	t.Parallel()

	bd := buildEmbeddedBD(t)
	p := newSharedProxiedProject(t, bd, "pds")
	if out, err := bdProxiedRun(t, bd, p.dir, "config", "set", "status.custom", "backlog"); err != nil {
		t.Fatalf("config set status.custom: %v\n%s", err, out)
	}
	setDefault := func(t *testing.T, value string) {
		t.Helper()
		if out, err := bdProxiedRun(t, bd, p.dir, "config", "set", "status.default", value); err != nil {
			t.Fatalf("config set status.default %s: %v\n%s", value, err, out)
		}
	}

	if got := bdProxiedCreate(t, bd, p.dir, "unset").Status; got != types.StatusOpen {
		t.Errorf("unset default: status = %q, want open", got)
	}

	setDefault(t, "backlog")
	if got := bdProxiedCreate(t, bd, p.dir, "defaulted").Status; got != "backlog" {
		t.Errorf("default applies: status = %q, want backlog", got)
	}
	if got := bdProxiedCreate(t, bd, p.dir, "flagged", "--status", "in_progress").Status; got != types.StatusInProgress {
		t.Errorf("--status beats the default: status = %q, want in_progress", got)
	}
	if got := bdProxiedCreate(t, bd, p.dir, "preview", "--dry-run").Status; got != "backlog" {
		t.Errorf("--dry-run previews the default: status = %q, want backlog", got)
	}

	setDefault(t, "nonsense")
	for _, args := range [][]string{{"bad default"}, {"bad default preview", "--dry-run"}} {
		if out := bdProxiedCreateFail(t, bd, p.dir, args...); !strings.Contains(out, "status.default") {
			t.Errorf("bd create %v error = %q, want it to name status.default", args, out)
		}
	}
	if got := bdProxiedCreate(t, bd, p.dir, "flagged", "--status", "blocked").Status; got != types.StatusBlocked {
		t.Errorf("--status skips an invalid default: status = %q, want blocked", got)
	}
	if got := bdProxiedCreate(t, bd, p.dir, "deferred", "--defer", "+2h").Status; got != types.StatusDeferred {
		t.Errorf("--defer skips an invalid default: status = %q, want deferred", got)
	}
}
