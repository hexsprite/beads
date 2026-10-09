//go:build cgo

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/types"
)

func TestEmbeddedCreateDefaultStatus(t *testing.T) {
	if os.Getenv("BEADS_TEST_EMBEDDED_DOLT") != "1" {
		t.Skip("set BEADS_TEST_EMBEDDED_DOLT=1 to run embedded dolt create tests")
	}
	t.Parallel()

	bd := buildEmbeddedBD(t)
	dir, _, _ := bdInit(t, bd, "--prefix", "ds")
	bdRunOK(t, bd, dir, "config", "set", "status.custom", "backlog")

	t.Run("unset_default_is_open", func(t *testing.T) {
		if got := bdCreate(t, bd, dir, "unset").Status; got != types.StatusOpen {
			t.Errorf("status = %q, want open", got)
		}
	})

	bdRunOK(t, bd, dir, "config", "set", "status.default", "backlog")

	t.Run("default_applies", func(t *testing.T) {
		if got := bdCreate(t, bd, dir, "defaulted").Status; got != "backlog" {
			t.Errorf("status = %q, want backlog", got)
		}
	})

	t.Run("status_flag_beats_default", func(t *testing.T) {
		if got := bdCreate(t, bd, dir, "flagged", "--status", "in_progress").Status; got != types.StatusInProgress {
			t.Errorf("status = %q, want in_progress", got)
		}
	})

	t.Run("dry_run_previews_the_default", func(t *testing.T) {
		if got := bdCreate(t, bd, dir, "preview", "--dry-run").Status; got != "backlog" {
			t.Errorf("dry-run status = %q, want backlog", got)
		}
	})

	t.Run("cross_repo_ignores_the_source_default", func(t *testing.T) {
		targetDir := filepath.Join(dir, "target-repo")
		if err := os.MkdirAll(targetDir, 0o750); err != nil {
			t.Fatal(err)
		}
		initGitRepoAt(t, targetDir)
		runBDInit(t, bd, targetDir, "--prefix", "tgt")

		if got := bdCreate(t, bd, dir, "routed", "--repo", targetDir).Status; got != types.StatusOpen {
			t.Errorf("cross-repo status = %q, want open: the source workspace's default must not follow the issue", got)
		}
		if got := bdCreate(t, bd, dir, "routed preview", "--repo", targetDir, "--dry-run").Status; got != types.StatusOpen {
			t.Errorf("cross-repo dry-run status = %q, want open", got)
		}
	})

	bdRunOK(t, bd, dir, "config", "set", "status.default", "nonsense")

	t.Run("invalid_default_is_rejected", func(t *testing.T) {
		if out := bdCreateFail(t, bd, dir, "bad default"); !strings.Contains(out, "status.default") {
			t.Errorf("error = %q, want it to name status.default", out)
		}
		if out := bdCreateFail(t, bd, dir, "bad default preview", "--dry-run"); !strings.Contains(out, "status.default") {
			t.Errorf("dry-run error = %q, want it to name status.default", out)
		}
	})

	t.Run("invalid_default_does_not_block_a_higher_precedence_status", func(t *testing.T) {
		if got := bdCreate(t, bd, dir, "flagged", "--status", "blocked").Status; got != types.StatusBlocked {
			t.Errorf("flag status = %q, want blocked", got)
		}
		if got := bdCreate(t, bd, dir, "deferred", "--defer", "+2h").Status; got != types.StatusDeferred {
			t.Errorf("defer status = %q, want deferred", got)
		}
	})

	t.Run("invalid_default_does_not_block_a_cross_repo_create", func(t *testing.T) {
		targetDir := filepath.Join(dir, "target-repo")
		if got := bdCreate(t, bd, dir, "routed again", "--repo", targetDir).Status; got != types.StatusOpen {
			t.Errorf("cross-repo status = %q, want open", got)
		}
	})
}
