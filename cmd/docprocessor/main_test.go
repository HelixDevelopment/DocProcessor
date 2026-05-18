// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Milos Vasic

package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"digital.vasic.docprocessor/pkg/i18n"
)

// fakeTranslator returns a deterministic sentinel for each msgID so
// tests can assert that the CLI's wire output went through the
// Translator layer rather than embedding a hardcoded English literal.
// Mocks permitted in unit-test sources per CONST-050(A).
type fakeTranslator struct{}

func (fakeTranslator) T(_ context.Context, msgID string, _ map[string]any) string {
	return "<TRANSLATED:" + msgID + ">"
}

func (fakeTranslator) TPlural(_ context.Context, msgID string, _ int, _ map[string]any) string {
	return "<TRANSLATED:" + msgID + ">"
}

// TestRunCLI_Usage_NoArgs_EmitsTranslatedUsage exercises the
// arguments-missing branch. The original code shipped a hardcoded
// "Usage: docprocessor <docs-directory>\n" literal; round 97 migrated
// it to msg ID `docprocessor_cli_usage`. The fake translator's
// sentinel proves the stderr line was resolved through the Translator
// indirection — no hardcoded English literal remains.
func TestRunCLI_Usage_NoArgs_EmitsTranslatedUsage(t *testing.T) {
	var stdout, stderr bytes.Buffer
	rc := runCLI(context.Background(), []string{"docprocessor"}, &stdout, &stderr, fakeTranslator{})
	if rc != 1 {
		t.Fatalf("runCLI exit code = %d, want 1", rc)
	}
	const sentinel = "<TRANSLATED:docprocessor_cli_usage>"
	if !strings.Contains(stderr.String(), sentinel) {
		t.Fatalf("stderr missing sentinel %q; got %q", sentinel, stderr.String())
	}
	if strings.Contains(stderr.String(), "Usage: docprocessor <docs-directory>") {
		t.Fatalf("hardcoded English literal leaked into stderr; got %q", stderr.String())
	}
}

// TestRunCLI_ErrorLoadingDocs_EmitsTranslatedError exercises the
// loader-error branch by pointing at a non-existent directory. The
// original "Error loading docs: %v" literal must NOT appear; the
// sentinel for `docprocessor_cli_error_loading_docs` must.
func TestRunCLI_ErrorLoadingDocs_EmitsTranslatedError(t *testing.T) {
	var stdout, stderr bytes.Buffer
	rc := runCLI(
		context.Background(),
		[]string{"docprocessor", "/nonexistent-path-round97-i18n-sentinel"},
		&stdout, &stderr, fakeTranslator{},
	)
	if rc != 1 {
		t.Fatalf("runCLI exit code = %d, want 1", rc)
	}
	const sentinel = "<TRANSLATED:docprocessor_cli_error_loading_docs>"
	if !strings.Contains(stderr.String(), sentinel) {
		t.Fatalf("stderr missing sentinel %q; got %q", sentinel, stderr.String())
	}
	if strings.Contains(stderr.String(), "Error loading docs:") {
		t.Fatalf("hardcoded English literal leaked into stderr; got %q", stderr.String())
	}
}

// TestRunCLI_SuccessPath_EmitsAllTranslatedLines exercises the
// happy-path summary block. The original code shipped five hardcoded
// formatted strings; this test asserts the captured stdout contains
// the matching sentinels and contains NONE of the historical English
// literals.
func TestRunCLI_SuccessPath_EmitsAllTranslatedLines(t *testing.T) {
	docsDir := t.TempDir()
	// Seed a minimal markdown doc so the feature-map summary emits
	// at least the loaded-documents + feature-map + doc-graph lines.
	docPath := filepath.Join(docsDir, "round97_sentinel.md")
	if err := os.WriteFile(docPath, []byte("# Round 97 i18n Sentinel\n\nDocProcessor CLI migration evidence.\n"), 0o644); err != nil {
		t.Fatalf("write seed doc: %v", err)
	}

	var stdout, stderr bytes.Buffer
	rc := runCLI(context.Background(), []string{"docprocessor", docsDir}, &stdout, &stderr, fakeTranslator{})
	if rc != 0 {
		t.Fatalf("runCLI exit code = %d, want 0 (stderr=%q)", rc, stderr.String())
	}

	wantSentinels := []string{
		"<TRANSLATED:docprocessor_cli_loaded_documents>",
		"<TRANSLATED:docprocessor_cli_feature_map_summary>",
		"<TRANSLATED:docprocessor_cli_doc_graph_summary>",
	}
	for _, s := range wantSentinels {
		if !strings.Contains(stdout.String(), s) {
			t.Fatalf("stdout missing sentinel %q; got %q", s, stdout.String())
		}
	}

	forbiddenLiterals := []string{
		"Loaded ",
		"Feature map:",
		"Doc graph:",
	}
	for _, lit := range forbiddenLiterals {
		if strings.Contains(stdout.String(), lit) {
			t.Fatalf("hardcoded English literal %q leaked into stdout; got %q", lit, stdout.String())
		}
	}
}

// TestRunCLI_NoopTranslator_EmitsMsgIDVerbatim confirms the default
// translator wired in main() (NoopTranslator) produces visible
// captured output containing the raw message IDs. This is the
// integration of CLI + default Translator (round 97's third-layer
// sanity check) — proves the production wiring isn't silently
// substituting empty strings.
func TestRunCLI_NoopTranslator_EmitsMsgIDVerbatim(t *testing.T) {
	var stdout, stderr bytes.Buffer
	rc := runCLI(context.Background(), []string{"docprocessor"}, &stdout, &stderr, i18n.NoopTranslator{})
	if rc != 1 {
		t.Fatalf("runCLI exit code = %d, want 1", rc)
	}
	if !strings.Contains(stderr.String(), "docprocessor_cli_usage") {
		t.Fatalf("noop fallback failed: stderr=%q", stderr.String())
	}
}
