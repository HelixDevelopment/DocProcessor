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
// it to msg ID `docprocessor_cli_usage`. Round 209 added a
// docprocessor_cli_help_header second line — both sentinels must
// appear in stderr.
func TestRunCLI_Usage_NoArgs_EmitsTranslatedUsage(t *testing.T) {
	var stdout, stderr bytes.Buffer
	rc := runCLI(context.Background(), []string{"docprocessor"}, &stdout, &stderr, fakeTranslator{})
	if rc != 1 {
		t.Fatalf("runCLI exit code = %d, want 1", rc)
	}
	wantSentinels := []string{
		"<TRANSLATED:docprocessor_cli_usage>",
		"<TRANSLATED:docprocessor_cli_help_header>",
	}
	for _, s := range wantSentinels {
		if !strings.Contains(stderr.String(), s) {
			t.Fatalf("stderr missing sentinel %q; got %q", s, stderr.String())
		}
	}
	if strings.Contains(stderr.String(), "Usage: docprocessor") {
		t.Fatalf("hardcoded English literal leaked into stderr; got %q", stderr.String())
	}
	if strings.Contains(stderr.String(), "DocProcessor extracts") {
		t.Fatalf("hardcoded help-header English literal leaked into stderr; got %q", stderr.String())
	}
}

// TestRunCLI_PathInvalid_EmitsTranslatedPathInvalid exercises the
// empty-path validation branch added in round 209.
func TestRunCLI_PathInvalid_EmitsTranslatedPathInvalid(t *testing.T) {
	var stdout, stderr bytes.Buffer
	rc := runCLI(context.Background(), []string{"docprocessor", "   "}, &stdout, &stderr, fakeTranslator{})
	if rc != 1 {
		t.Fatalf("runCLI exit code = %d, want 1", rc)
	}
	const sentinel = "<TRANSLATED:docprocessor_cli_path_invalid>"
	if !strings.Contains(stderr.String(), sentinel) {
		t.Fatalf("stderr missing sentinel %q; got %q", sentinel, stderr.String())
	}
	if strings.Contains(stderr.String(), "Invalid docs-directory") {
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

// TestRunCLI_NoDocsFound_EmitsTranslatedNoDocs exercises the
// zero-documents branch added in round 209 — points at a real but
// empty directory and asserts the no-docs sentinel + done sentinel
// appear instead of the historical silent success.
func TestRunCLI_NoDocsFound_EmitsTranslatedNoDocs(t *testing.T) {
	docsDir := t.TempDir() // empty directory — no supported docs
	var stdout, stderr bytes.Buffer
	rc := runCLI(context.Background(), []string{"docprocessor", docsDir}, &stdout, &stderr, fakeTranslator{})
	if rc != 0 {
		t.Fatalf("runCLI exit code = %d, want 0 (stderr=%q)", rc, stderr.String())
	}
	wantSentinels := []string{
		"<TRANSLATED:docprocessor_cli_format_summary>",
		"<TRANSLATED:docprocessor_cli_loaded_documents>",
		"<TRANSLATED:docprocessor_cli_no_docs_found>",
		"<TRANSLATED:docprocessor_cli_done>",
	}
	for _, s := range wantSentinels {
		if !strings.Contains(stdout.String(), s) {
			t.Fatalf("stdout missing sentinel %q; got %q", s, stdout.String())
		}
	}
	forbiddenLiterals := []string{
		"No supported documents",
		"Done in",
		"Scanning for formats:",
	}
	for _, lit := range forbiddenLiterals {
		if strings.Contains(stdout.String(), lit) {
			t.Fatalf("hardcoded English literal %q leaked into stdout; got %q", lit, stdout.String())
		}
	}
}

// TestRunCLI_SuccessPath_EmitsAllTranslatedLines exercises the
// happy-path summary block including round 209 additions (format
// summary, summary header, done line).
func TestRunCLI_SuccessPath_EmitsAllTranslatedLines(t *testing.T) {
	docsDir := t.TempDir()
	docPath := filepath.Join(docsDir, "round209_sentinel.md")
	if err := os.WriteFile(docPath, []byte("# Round 209 i18n Sentinel\n\nDocProcessor CLI round-2 migration evidence.\n"), 0o644); err != nil {
		t.Fatalf("write seed doc: %v", err)
	}

	var stdout, stderr bytes.Buffer
	rc := runCLI(context.Background(), []string{"docprocessor", docsDir}, &stdout, &stderr, fakeTranslator{})
	if rc != 0 {
		t.Fatalf("runCLI exit code = %d, want 0 (stderr=%q)", rc, stderr.String())
	}

	wantSentinels := []string{
		"<TRANSLATED:docprocessor_cli_format_summary>",
		"<TRANSLATED:docprocessor_cli_loaded_documents>",
		"<TRANSLATED:docprocessor_cli_summary_header>",
		"<TRANSLATED:docprocessor_cli_feature_map_summary>",
		"<TRANSLATED:docprocessor_cli_doc_graph_summary>",
		"<TRANSLATED:docprocessor_cli_done>",
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
		"=== Feature map summary ===",
		"Scanning for formats:",
		"Done in",
	}
	for _, lit := range forbiddenLiterals {
		if strings.Contains(stdout.String(), lit) {
			t.Fatalf("hardcoded English literal %q leaked into stdout; got %q", lit, stdout.String())
		}
	}
}

// TestRunCLI_VerboseFlag_EmitsPerItemLines exercises the verbose
// branch added in round 209 — feature/screen/workflow per-item
// sentinels appear only when --verbose is set. Since the heuristic
// extractor only emits Features (no Screens or Workflows without an
// LLM agent), this test asserts at minimum the feature-line sentinel
// shows up.
func TestRunCLI_VerboseFlag_EmitsPerItemLines(t *testing.T) {
	docsDir := t.TempDir()
	docPath := filepath.Join(docsDir, "verbose_sentinel.md")
	body := "# Round 209 Verbose\n\n" +
		"## Verbose Feature One\n\n" +
		"This is a long enough description to pass the 20-char heuristic gate so a feature gets created.\n\n" +
		"## Verbose Feature Two\n\n" +
		"Another long enough description to pass the 20-char heuristic gate so a second feature appears.\n"
	if err := os.WriteFile(docPath, []byte(body), 0o644); err != nil {
		t.Fatalf("write seed doc: %v", err)
	}

	var stdout, stderr bytes.Buffer
	rc := runCLI(context.Background(), []string{"docprocessor", "--verbose", docsDir}, &stdout, &stderr, fakeTranslator{})
	if rc != 0 {
		t.Fatalf("runCLI exit code = %d, want 0 (stderr=%q)", rc, stderr.String())
	}

	const featureSentinel = "<TRANSLATED:docprocessor_cli_feature_line>"
	if !strings.Contains(stdout.String(), featureSentinel) {
		t.Fatalf("stdout missing feature-line sentinel %q; got %q", featureSentinel, stdout.String())
	}
	if strings.Contains(stdout.String(), "Feature [") {
		t.Fatalf("hardcoded English literal leaked into stdout; got %q", stdout.String())
	}
}

// TestRunCLI_VerboseFlag_OmittedSuppressesPerItemLines proves the
// verbose gating actually gates — without --verbose the per-feature
// line sentinel MUST NOT appear. Paired-mutation guard: if the
// verbose check is accidentally removed in main.go, this test fails.
func TestRunCLI_VerboseFlag_OmittedSuppressesPerItemLines(t *testing.T) {
	docsDir := t.TempDir()
	docPath := filepath.Join(docsDir, "no_verbose_sentinel.md")
	body := "# Round 209 No-Verbose\n\n" +
		"## Hidden Feature\n\n" +
		"This is a long enough description to pass the 20-char heuristic gate.\n"
	if err := os.WriteFile(docPath, []byte(body), 0o644); err != nil {
		t.Fatalf("write seed doc: %v", err)
	}

	var stdout, stderr bytes.Buffer
	rc := runCLI(context.Background(), []string{"docprocessor", docsDir}, &stdout, &stderr, fakeTranslator{})
	if rc != 0 {
		t.Fatalf("runCLI exit code = %d, want 0 (stderr=%q)", rc, stderr.String())
	}

	forbiddenSentinels := []string{
		"<TRANSLATED:docprocessor_cli_feature_line>",
		"<TRANSLATED:docprocessor_cli_screen_line>",
		"<TRANSLATED:docprocessor_cli_workflow_line>",
	}
	for _, s := range forbiddenSentinels {
		if strings.Contains(stdout.String(), s) {
			t.Fatalf("non-verbose run leaked sentinel %q; got %q", s, stdout.String())
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
	if !strings.Contains(stderr.String(), "docprocessor_cli_help_header") {
		t.Fatalf("round 209 help-header noop fallback missing: stderr=%q", stderr.String())
	}
}

// TestRunCLI_BundleContainsAllRound209MsgIDs is the paired-mutation
// integrity check for the YAML bundle. It loads active.en.yaml and
// asserts every round 209 message ID is present. Planting a typo in
// the YAML (e.g. renaming `docprocessor_cli_done`) makes this test
// fail — proving the bundle is authoritative for sentinel coverage.
func TestRunCLI_BundleContainsAllRound209MsgIDs(t *testing.T) {
	// Resolve bundle path relative to this test file via runtime CWD —
	// `go test` runs with CWD set to the package directory; the bundle
	// lives two levels up at pkg/i18n/bundles/active.en.yaml.
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	bundlePath := filepath.Join(cwd, "..", "..", "pkg", "i18n", "bundles", "active.en.yaml")
	data, err := os.ReadFile(bundlePath)
	if err != nil {
		t.Fatalf("read bundle: %v", err)
	}
	content := string(data)
	round209IDs := []string{
		"docprocessor_cli_help_header",
		"docprocessor_cli_path_invalid",
		"docprocessor_cli_error_resolving_path",
		"docprocessor_cli_no_docs_found",
		"docprocessor_cli_format_summary",
		"docprocessor_cli_summary_header",
		"docprocessor_cli_feature_line",
		"docprocessor_cli_screen_line",
		"docprocessor_cli_workflow_line",
		"docprocessor_cli_done",
	}
	for _, id := range round209IDs {
		if !strings.Contains(content, id+":") {
			t.Fatalf("bundle missing round 209 msg ID %q", id)
		}
	}
}

// TestNewTranslator_ProductionWiringRendersVisibleOutput is the §1.1
// paired-mutation guard for main()'s production translator wiring. It
// drives runCLI through the REAL translator returned by newTranslator()
// (the same object main() hands runCLI) and asserts the usage branch
// produces non-empty, ID-bearing stderr. The forbidden mutation this
// catches: replacing newTranslator()'s return with a translator whose
// T/TPlural return "" (an empty-string substitution would let every
// runCLI line silently vanish — a §11.4 PASS-bluff at the i18n layer).
// If newTranslator() is mutated to emit empty strings, the
// non-empty + sentinel assertions below FAIL.
func TestNewTranslator_ProductionWiringRendersVisibleOutput(t *testing.T) {
	tr := newTranslator()
	if tr == nil {
		t.Fatal("newTranslator() returned nil — main() would panic dispatching runCLI")
	}

	var stdout, stderr bytes.Buffer
	rc := runCLI(context.Background(), []string{"docprocessor"}, &stdout, &stderr, tr)
	if rc != 1 {
		t.Fatalf("runCLI(no-args) exit code = %d, want 1", rc)
	}
	if strings.TrimSpace(stderr.String()) == "" {
		t.Fatalf("production translator produced empty stderr — i18n wiring is a no-op bluff")
	}
	// The production NoopTranslator echoes IDs verbatim as positive
	// evidence per pkg/i18n's Article XI §11.9 contract.
	for _, id := range []string{"docprocessor_cli_usage", "docprocessor_cli_help_header"} {
		if !strings.Contains(stderr.String(), id) {
			t.Fatalf("production stderr missing rendered msg ID %q; got %q", id, stderr.String())
		}
	}
}
