// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Milos Vasic

package i18n

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestNoopTranslator_EchoesMessageID confirms the loud-fallback
// contract: an unwired Translator returns the message ID verbatim so
// a missing bundle is visible (anti-bluff per Article XI §11.9).
func TestNoopTranslator_EchoesMessageID(t *testing.T) {
	var tr Translator = NoopTranslator{}
	require.Equal(t, "cli_usage", tr.T(context.Background(), "cli_usage", nil))
	require.Equal(t, "cli_docs_loaded",
		tr.TPlural(context.Background(), "cli_docs_loaded", 3, nil))
}

// TestBundleTranslator_ResolvesEnglish proves the English bundle
// loads and every migrated CLI key renders real text.
func TestBundleTranslator_ResolvesEnglish(t *testing.T) {
	tr, err := NewBundleTranslator("en")
	require.NoError(t, err)

	ctx := context.Background()
	got := tr.T(ctx, "cli_usage", nil)
	require.Equal(t, "Usage: docprocessor <docs-directory>", got)
	require.NotEqual(t, "cli_usage", got, "key must not echo verbatim")

	// Placeholder interpolation.
	got = tr.T(ctx, "cli_error_loading_docs", map[string]any{"err": "boom"})
	require.Equal(t, "Error loading docs: boom", got)

	got = tr.T(ctx, "cli_feature_map_summary", map[string]any{
		"features": 5, "screens": 2, "workflows": 1,
	})
	require.Equal(t, "Feature map: 5 features, 2 screens, 1 workflows", got)
}

// TestBundleTranslator_ResolvesFlatCLIBundle is the W6C RED→GREEN
// proof. The docprocessor CLI (cmd/docprocessor) emits message IDs in
// the `docprocessor_cli_*` namespace, which live ONLY in the FLAT
// `active.en.yaml` surface (plain-string values, no {one,other}
// nesting). Before W6C, BundleTranslator skipped every `active.*` file
// outright, so these IDs never loaded and T() echoed the ID verbatim —
// newTranslator therefore had to return NoopTranslator. This test
// FAILS on that pre-fix code (T returns the ID) and PASSES once the
// flat surface is merged into the locale map.
//
// RED proof (pre-fix): got "docprocessor_cli_usage" want the sentence.
func TestBundleTranslator_ResolvesFlatCLIBundle(t *testing.T) {
	tr, err := NewBundleTranslator("en")
	require.NoError(t, err)
	ctx := context.Background()

	// Plain (no-placeholder) flat key resolves to its real string.
	got := tr.T(ctx, "docprocessor_cli_usage", nil)
	require.Equal(t, "Usage: docprocessor [--verbose] <docs-directory>", got)
	require.NotEqual(t, "docprocessor_cli_usage", got,
		"flat CLI key must resolve, not echo verbatim (Noop behaviour)")

	// Flat key with {{.error}} placeholder interpolates.
	got = tr.T(ctx, "docprocessor_cli_error_loading_docs",
		map[string]any{"error": "boom"})
	require.Equal(t, "Error loading docs: boom", got)

	// Flat key with multiple placeholders interpolates all of them.
	got = tr.T(ctx, "docprocessor_cli_feature_map_summary", map[string]any{
		"features": 5, "screens": 2, "workflows": 1,
	})
	require.Equal(t, "Feature map: 5 features, 2 screens, 1 workflows", got)

	// A flat-namespace key the CLI uses but with NO locale-specific
	// override still resolves via the default-locale (en) bundle.
	got = tr.T(WithLocale(ctx, "sr"), "docprocessor_cli_done",
		map[string]any{"elapsed_ms": 12})
	require.Equal(t, "Done in 12 ms.", got,
		"unknown-locale flat key must fall back to default, not echo")
}

// TestBundleTranslator_AllFlatCLIKeysResolve is the completeness gate
// for the flat surface: every `docprocessor_cli_*` ID the CLI emits
// MUST resolve (not echo) and render non-empty in the default locale.
func TestBundleTranslator_AllFlatCLIKeysResolve(t *testing.T) {
	tr, err := NewBundleTranslator("en")
	require.NoError(t, err)
	ctx := context.Background()

	flatKeys := []string{
		"docprocessor_cli_usage",
		"docprocessor_cli_error_loading_docs",
		"docprocessor_cli_loaded_documents",
		"docprocessor_cli_error_building_feature_map",
		"docprocessor_cli_feature_map_summary",
		"docprocessor_cli_doc_graph_summary",
		"docprocessor_cli_category_line",
		"docprocessor_cli_platform_line",
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
	for _, k := range flatKeys {
		got := tr.T(ctx, k, nil)
		require.NotEqual(t, k, got, "flat key %q must resolve (not echo)", k)
		require.NotEmpty(t, strings.TrimSpace(got), "flat key %q resolved empty", k)
	}
}

// TestBundleTranslator_PluralSelection proves count==1 selects the
// "one" form and other counts select "other".
func TestBundleTranslator_PluralSelection(t *testing.T) {
	tr, err := NewBundleTranslator("en")
	require.NoError(t, err)
	ctx := context.Background()

	require.Equal(t, "Loaded 1 document", tr.TPlural(ctx, "cli_docs_loaded", 1, nil))
	require.Equal(t, "Loaded 7 documents", tr.TPlural(ctx, "cli_docs_loaded", 7, nil))
	require.Equal(t, "  Category auth: 1 feature",
		tr.TPlural(ctx, "cli_category_line", 1, map[string]any{"category": "auth"}))
	require.Equal(t, "  Category auth: 3 features",
		tr.TPlural(ctx, "cli_category_line", 3, map[string]any{"category": "auth"}))
}

// TestBundleTranslator_LocaleSwitch is the anti-bluff proof that the
// seam GENUINELY localises: requesting "sr" yields Serbian text, not
// the English literal. A no-op translator would fail this.
func TestBundleTranslator_LocaleSwitch(t *testing.T) {
	tr, err := NewBundleTranslator("en")
	require.NoError(t, err)

	en := tr.T(context.Background(), "cli_usage", nil)
	sr := tr.T(WithLocale(context.Background(), "sr"), "cli_usage", nil)

	require.Equal(t, "Usage: docprocessor <docs-directory>", en)
	require.Equal(t, "Upotreba: docprocessor <direktorijum-dokumentacije>", sr)
	require.NotEqual(t, en, sr, "Serbian locale must differ from English")

	// Base-language fallback: "sr-RS" resolves to the "sr" bundle.
	srRS := tr.T(WithLocale(context.Background(), "sr-RS"), "cli_usage", nil)
	require.Equal(t, sr, srRS)
}

// TestBundleTranslator_UnknownLocaleFallsBack confirms an unknown
// locale falls back to the default bundle rather than echoing.
func TestBundleTranslator_UnknownLocaleFallsBack(t *testing.T) {
	tr, err := NewBundleTranslator("en")
	require.NoError(t, err)

	got := tr.T(WithLocale(context.Background(), "xx"), "cli_usage", nil)
	require.Equal(t, "Usage: docprocessor <docs-directory>", got)
}

// TestBundleTranslator_UnknownKeyEchoes confirms an absent key is
// echoed verbatim — loud, never silently swallowed.
func TestBundleTranslator_UnknownKeyEchoes(t *testing.T) {
	tr, err := NewBundleTranslator("en")
	require.NoError(t, err)
	require.Equal(t, "no_such_key",
		tr.T(context.Background(), "no_such_key", nil))
}

// TestBundleTranslator_MissingDefaultBundleErrors is the paired
// mutation: asking for a default locale with no embedded bundle MUST
// fail construction. If this passed, a typo'd default could silently
// echo every key.
func TestBundleTranslator_MissingDefaultBundleErrors(t *testing.T) {
	_, err := NewBundleTranslator("zz")
	require.Error(t, err)
	require.Contains(t, err.Error(), "zz")
}

// TestBundleTranslator_AllCLIKeysPresent is the completeness gate:
// every message ID the docprocessor CLI uses MUST resolve in BOTH
// shipped bundles. A new CLI string with no bundle entry — or a
// bundle that drifts from the other — fails here loudly.
func TestBundleTranslator_AllCLIKeysPresent(t *testing.T) {
	tr, err := NewBundleTranslator("en")
	require.NoError(t, err)

	cliKeys := []string{
		"cli_usage",
		"cli_error_loading_docs",
		"cli_error_building_feature_map",
		"cli_docs_loaded",
		"cli_feature_map_summary",
		"cli_doc_graph_summary",
		"cli_category_line",
		"cli_platform_line",
	}
	for _, locale := range []string{"en", "sr"} {
		for _, k := range cliKeys {
			got := tr.T(WithLocale(context.Background(), locale), k, nil)
			require.NotEqual(t, k, got,
				"key %q must resolve (not echo) in locale %q", k, locale)
			require.NotEmpty(t, strings.TrimSpace(got),
				"key %q resolved empty in locale %q", k, locale)
		}
	}
}
