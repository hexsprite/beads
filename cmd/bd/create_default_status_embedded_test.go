//go:build cgo

package main

import (
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/types"
)

// A configured status.default is the status of a create that names none.
// Symptom: `bd config get status.default` returned "draft" while
// `bd create --dry-run` and the created issue both reported "open".
func TestEmbeddedCreateHonorsStatusDefault(t *testing.T) {
	bd := buildEmbeddedBD(t)
	dir, _, _ := bdInit(t, bd, "--prefix", "sdf")

	for _, kv := range [][]string{
		{"status.custom", "draft"},
		{"status.default", "draft"},
	} {
		if out, err := bdRunWithFlockRetry(t, bd, dir, "config", "set", kv[0], kv[1]); err != nil {
			t.Fatalf("bd config set %s: %v\n%s", kv[0], err, out)
		}
	}

	t.Run("no flag uses the configured default", func(t *testing.T) {
		if got := bdCreate(t, bd, dir, "Defaulted").Status; got != types.Status("draft") {
			t.Fatalf("status = %q, want draft", got)
		}
	})

	t.Run("dry-run reports the configured default", func(t *testing.T) {
		if got := bdCreate(t, bd, dir, "Preview", "--dry-run").Status; got != types.Status("draft") {
			t.Fatalf("dry-run status = %q, want draft", got)
		}
	})

	t.Run("explicit status flag wins", func(t *testing.T) {
		if got := bdCreate(t, bd, dir, "Explicit", "--status", "open").Status; got != types.StatusOpen {
			t.Fatalf("status = %q, want open", got)
		}
		if got := bdCreate(t, bd, dir, "Explicit preview", "--status", "open", "--dry-run").Status; got != types.StatusOpen {
			t.Fatalf("dry-run status = %q, want open", got)
		}
	})
}

func TestEmbeddedCreateWithoutStatusDefaultIsOpen(t *testing.T) {
	bd := buildEmbeddedBD(t)
	dir, _, _ := bdInit(t, bd, "--prefix", "sdn")

	if got := bdCreate(t, bd, dir, "Plain").Status; got != types.StatusOpen {
		t.Fatalf("status = %q, want open", got)
	}
}

func TestEmbeddedCreateRejectsUnknownStatusDefault(t *testing.T) {
	bd := buildEmbeddedBD(t)
	dir, _, _ := bdInit(t, bd, "--prefix", "sdu")

	if out, err := bdRunWithFlockRetry(t, bd, dir, "config", "set", "status.default", "nonesuch"); err != nil {
		t.Fatalf("bd config set status.default: %v\n%s", err, out)
	}

	out := bdCreateFail(t, bd, dir, "Doomed")
	if !strings.Contains(out, "status.default") {
		t.Fatalf("error does not name status.default: %s", out)
	}
}
