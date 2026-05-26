// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Milos Vasic

<<<<<<< HEAD
package i18n_test

import (
	"context"
	"testing"

	"digital.vasic.docprocessor/pkg/i18n"
)

// TestNoopTranslator_T_ReturnsMsgIDVerbatim asserts that the
// stripped-down fallback Translator emits the message ID unchanged.
// Per CONST-035 / Article XI §11.9 this verbatim-fallback is itself
// positive runtime evidence — operators see exactly which key was
// resolved without a bundle.
func TestNoopTranslator_T_ReturnsMsgIDVerbatim(t *testing.T) {
	tr := i18n.NoopTranslator{}
	got := tr.T(context.Background(), "docprocessor_cli_loaded_documents", map[string]any{
		"count": 42,
	})
	const want = "docprocessor_cli_loaded_documents"
	if got != want {
		t.Fatalf("NoopTranslator.T mismatch:\n got = %q\nwant = %q", got, want)
	}
}

// TestNoopTranslator_TPlural_ReturnsMsgIDVerbatim mirrors the T
// assertion for plural-form lookups.
func TestNoopTranslator_TPlural_ReturnsMsgIDVerbatim(t *testing.T) {
	tr := i18n.NoopTranslator{}
	got := tr.TPlural(context.Background(), "docprocessor_cli_category_line", 3, nil)
	const want = "docprocessor_cli_category_line"
	if got != want {
		t.Fatalf("NoopTranslator.TPlural mismatch:\n got = %q\nwant = %q", got, want)
=======
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
>>>>>>> 9f5637d2d695cd5fcf8349d1f1b8bf780fa5d865
	}
}
