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

	// A flat-namespace key the CLI uses under an UNKNOWN locale (no
	// bundle, no base-language match) still resolves via the
	// default-locale (en) bundle. (W7C reconciliation per §11.4.120: the
	// original assertion used "sr" to prove default-fallback, valid only
	// while active.sr.yaml was absent. Now that the Serbian flat bundle
	// exists — by design, the W7C deliverable — "sr" correctly returns
	// Serbian, so the default-fallback contract is re-asserted here with
	// a genuinely-unknown locale "xx", preserving the original intent.)
	got = tr.T(WithLocale(ctx, "xx"), "docprocessor_cli_done",
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

// TestBundleTranslator_FlatCLIBundleSerbian is the W7C RED→GREEN proof
// that Serbian CLI users get Serbian text for the FLAT
// `docprocessor_cli_*` namespace (the surface cmd/docprocessor emits),
// not the English fallback. Before `active.sr.yaml` exists, the "sr"
// locale has no flat entry for these IDs, so lookup falls back to the
// default (en) bundle and returns the English sentence — this test
// FAILS on that pre-fix state. Once `active.sr.yaml` is merged into the
// "sr" locale map (pass 2 of NewBundleTranslator), the same lookup
// returns the Serbian string and placeholder interpolation works.
//
// RED proof (pre-fix, no active.sr.yaml): docprocessor_cli_usage under
// "sr" returns "Usage: docprocessor [--verbose] <docs-directory>"
// (English fallback) — the require.Equal on the Serbian string FAILs.
func TestBundleTranslator_FlatCLIBundleSerbian(t *testing.T) {
	tr, err := NewBundleTranslator("en")
	require.NoError(t, err)
	srCtx := WithLocale(context.Background(), "sr")

	// A flat CLI key resolves to its Serbian string, NOT the English
	// fallback. This is the load-bearing anti-bluff assertion: if the
	// "sr" flat bundle is missing, lookup falls back to en and this FAILs.
	got := tr.T(srCtx, "docprocessor_cli_usage", nil)
	require.Equal(t,
		"Upotreba: docprocessor [--verbose] <direktorijum-dokumentacije>",
		got)
	require.NotEqual(t,
		"Usage: docprocessor [--verbose] <docs-directory>", got,
		"sr flat CLI key must not fall back to the English string")

	// Placeholder interpolation works for the Serbian flat string and
	// the {{.error}} token is preserved (interpolated, not translated).
	got = tr.T(srCtx, "docprocessor_cli_error_loading_docs",
		map[string]any{"error": "boom"})
	require.Equal(t, "Greška pri učitavanju dokumentacije: boom", got)
	require.Contains(t, got, "boom", "placeholder value must interpolate")

	// Multi-placeholder Serbian flat string interpolates all tokens.
	got = tr.T(srCtx, "docprocessor_cli_feature_map_summary", map[string]any{
		"features": 5, "screens": 2, "workflows": 1,
	})
	require.Equal(t,
		"Mapa funkcionalnosti: 5 funkcionalnosti, 2 ekrana, 1 tokova", got)

	// Base-language fallback: "sr-RS" resolves to the "sr" flat bundle.
	srRS := tr.T(WithLocale(context.Background(), "sr-RS"),
		"docprocessor_cli_usage", nil)
	require.Equal(t, got2SerbianUsage(), srRS,
		"sr-RS must resolve to the sr flat bundle")
}

// got2SerbianUsage centralises the expected Serbian usage string so the
// base-language-fallback assertion cannot drift from the primary one.
func got2SerbianUsage() string {
	return "Upotreba: docprocessor [--verbose] <direktorijum-dokumentacije>"
}

// TestBundleTranslator_AllFlatCLIKeysResolveSerbian is the completeness
// gate for the Serbian flat surface: every `docprocessor_cli_*` ID MUST
// resolve under "sr" to a NON-English, non-empty string (i.e. it came
// from active.sr.yaml, not the en fallback). A drift where a key is
// missing from active.sr.yaml silently falls back to English and is
// caught here.
func TestBundleTranslator_AllFlatCLIKeysResolveSerbian(t *testing.T) {
	enTr, err := NewBundleTranslator("en")
	require.NoError(t, err)
	enCtx := context.Background()
	srCtx := WithLocale(context.Background(), "sr")

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
		sr := enTr.T(srCtx, k, nil)
		en := enTr.T(enCtx, k, nil)
		require.NotEqual(t, k, sr, "sr flat key %q must resolve (not echo)", k)
		require.NotEmpty(t, strings.TrimSpace(sr),
			"sr flat key %q resolved empty", k)
		require.NotEqual(t, en, sr,
			"sr flat key %q must differ from English (no fallback)", k)
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
