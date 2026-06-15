// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Milos Vasic

package i18n

import (
	"context"
	"testing"
)

// These tests exercise the genuine resolution behaviour of
// BundleTranslator that the existing translator_test.go does not yet
// cover: the nil-context path of localeFromContext, the base-language
// fallback in lookup ("sr-RS" -> "sr"), the count==1-but-no-One-form
// fall-through in TPlural, and the unknown-key echo path of TPlural.
//
// Each test asserts a concrete, end-user-visible rendering against the
// REAL embedded bundles (en.yaml / sr.yaml). Per §11.4 / CONST-035,
// every assertion catches its own negation: replace the resolution
// logic with a trivial stub and these tests fail.

func newTestTranslator(t *testing.T) *BundleTranslator {
	t.Helper()
	bt, err := NewBundleTranslator("en")
	if err != nil {
		t.Fatalf("NewBundleTranslator(en): %v", err)
	}
	return bt
}

// TestLocaleFromContext drives the unexported helper directly through
// its three branches, including the nil-context branch (bundle.go:57-59)
// that no public-API test reaches.
func TestLocaleFromContext(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		ctx  context.Context
		want string
	}{
		{
			name: "nil context yields empty locale",
			ctx:  nil,
			want: "",
		},
		{
			name: "context without locale yields empty",
			ctx:  context.Background(),
			want: "",
		},
		{
			name: "context with locale yields it",
			ctx:  WithLocale(context.Background(), "sr"),
			want: "sr",
		},
		{
			name: "later WithLocale overrides earlier",
			ctx:  WithLocale(WithLocale(context.Background(), "en"), "sr"),
			want: "sr",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := localeFromContext(tt.ctx); got != tt.want {
				t.Fatalf("localeFromContext = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestT_NilContextUsesDefaultLocale proves that calling T with a nil
// context (a real caller scenario) resolves against the default locale
// rather than panicking — the public-API path through the nil branch of
// localeFromContext.
func TestT_NilContextUsesDefaultLocale(t *testing.T) {
	t.Parallel()
	bt := newTestTranslator(t)

	//nolint:staticcheck // SA1012: passing nil context is the behaviour under test.
	got := bt.T(nil, "cli_usage", nil)
	want := "Usage: docprocessor <docs-directory>"
	if got != want {
		t.Fatalf("T(nil, cli_usage) = %q, want %q", got, want)
	}
}

// TestT_BaseLanguageFallback proves the "sr-RS" -> "sr" base-language
// fallback in lookup: a region-tagged locale with no exact bundle
// resolves to its base-language bundle, NOT to the default English.
func TestT_BaseLanguageFallback(t *testing.T) {
	t.Parallel()
	bt := newTestTranslator(t)

	tests := []struct {
		name   string
		locale string
		want   string
	}{
		{
			name:   "dash region tag falls back to base language",
			locale: "sr-RS",
			want:   "Upotreba: docprocessor <direktorijum-dokumentacije>",
		},
		{
			name:   "underscore region tag falls back to base language",
			locale: "sr_RS",
			want:   "Upotreba: docprocessor <direktorijum-dokumentacije>",
		},
		{
			name:   "exact base locale resolves directly",
			locale: "sr",
			want:   "Upotreba: docprocessor <direktorijum-dokumentacije>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := WithLocale(context.Background(), tt.locale)
			if got := bt.T(ctx, "cli_usage", nil); got != tt.want {
				t.Fatalf("T(%s, cli_usage) = %q, want %q", tt.locale, got, tt.want)
			}
		})
	}
}

// TestT_UnknownRegionFallsThroughToDefault proves that a region tag
// whose base language ALSO has no bundle falls through to the default
// locale (English) — distinct from the base-language fallback above.
func TestT_UnknownRegionFallsThroughToDefault(t *testing.T) {
	t.Parallel()
	bt := newTestTranslator(t)

	ctx := WithLocale(context.Background(), "ja-JP") // no ja bundle, no ja-JP bundle
	got := bt.T(ctx, "cli_usage", nil)
	want := "Usage: docprocessor <docs-directory>" // default-locale (en) rendering
	if got != want {
		t.Fatalf("T(ja-JP, cli_usage) = %q, want default-en %q", got, want)
	}
}

// TestTPlural_FormSelection drives the plural selector across its real
// branches: count==1 with a One form, count!=1 with the Other form,
// and the {{.count}} placeholder being injected.
func TestTPlural_FormSelection(t *testing.T) {
	t.Parallel()
	bt := newTestTranslator(t)

	tests := []struct {
		name  string
		count int
		want  string
	}{
		{name: "singular uses one form", count: 1, want: "Loaded 1 document"},
		{name: "zero uses other form", count: 0, want: "Loaded 0 documents"},
		{name: "plural uses other form", count: 5, want: "Loaded 5 documents"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := bt.TPlural(context.Background(), "cli_docs_loaded", tt.count, nil)
			if got != tt.want {
				t.Fatalf("TPlural(cli_docs_loaded, %d) = %q, want %q", tt.count, got, tt.want)
			}
		})
	}
}

// TestTPlural_NoOneFormFallsBackToOther proves the count==1-but-empty-One
// fall-through (bundle.go:224-226): cli_usage has only an Other form, so
// TPlural with count==1 MUST render the Other form rather than an empty
// string. The {{.count}} placeholder is absent from cli_usage, so the
// rendered text is the plain Other string.
func TestTPlural_NoOneFormFallsBackToOther(t *testing.T) {
	t.Parallel()
	bt := newTestTranslator(t)

	got := bt.TPlural(context.Background(), "cli_usage", 1, nil)
	want := "Usage: docprocessor <docs-directory>"
	if got != want {
		t.Fatalf("TPlural(cli_usage, 1) = %q, want Other-form %q", got, want)
	}
}

// TestTPlural_UnknownKeyEchoes drives the unknown-messageID echo path of
// TPlural (bundle.go:221-223): an absent key is returned verbatim (loud
// fallback per the package contract / §11.4 no-silent-swallow), and the
// supplied count/args are NOT interpolated because no bundle entry was
// found.
func TestTPlural_UnknownKeyEchoes(t *testing.T) {
	t.Parallel()
	bt := newTestTranslator(t)

	const unknown = "cli_no_such_plural_key"
	got := bt.TPlural(context.Background(), unknown, 3, map[string]any{"x": "y"})
	if got != unknown {
		t.Fatalf("TPlural(unknown) = %q, want verbatim echo %q", got, unknown)
	}
}

// TestTPlural_MergesArgsWithCount proves TPlural injects the count under
// the {{.count}} key alongside caller-supplied args, using the
// category-line message which carries both {{.category}} and {{.count}}.
func TestTPlural_MergesArgsWithCount(t *testing.T) {
	t.Parallel()
	bt := newTestTranslator(t)

	gotSingular := bt.TPlural(context.Background(), "cli_category_line", 1,
		map[string]any{"category": "UI"})
	wantSingular := "  Category UI: 1 feature"
	if gotSingular != wantSingular {
		t.Fatalf("TPlural(cli_category_line, 1) = %q, want %q", gotSingular, wantSingular)
	}

	gotPlural := bt.TPlural(context.Background(), "cli_category_line", 4,
		map[string]any{"category": "UI"})
	wantPlural := "  Category UI: 4 features"
	if gotPlural != wantPlural {
		t.Fatalf("TPlural(cli_category_line, 4) = %q, want %q", gotPlural, wantPlural)
	}
}

// TestInterpolate_EdgeCases exercises the single-scan interpolator's
// branches directly: empty args short-circuit, no-placeholder short-
// circuit, an unterminated "{{." (no closing "}}"), and an unknown key
// left literal. These guard the order-independent contract documented on
// interpolate().
func TestInterpolate_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		tmpl string
		args map[string]any
		want string
	}{
		{
			name: "nil args returns template verbatim",
			tmpl: "hello {{.name}}",
			args: nil,
			want: "hello {{.name}}",
		},
		{
			name: "no placeholder returns template verbatim",
			tmpl: "plain text",
			args: map[string]any{"name": "x"},
			want: "plain text",
		},
		{
			name: "known key substituted",
			tmpl: "hi {{.name}}!",
			args: map[string]any{"name": "Mila"},
			want: "hi Mila!",
		},
		{
			name: "unknown key left literal",
			tmpl: "hi {{.missing}}!",
			args: map[string]any{"name": "Mila"},
			want: "hi {{.missing}}!",
		},
		{
			name: "unterminated placeholder left literal",
			tmpl: "hi {{.name",
			args: map[string]any{"name": "Mila"},
			want: "hi {{.name",
		},
		{
			name: "value containing placeholder syntax is not re-expanded",
			tmpl: "path={{.p}}",
			args: map[string]any{"p": "{{.name}}", "name": "leak"},
			want: "path={{.name}}",
		},
		{
			name: "non-string arg rendered via fmt.Sprint",
			tmpl: "n={{.n}}",
			args: map[string]any{"n": 42},
			want: "n=42",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := interpolate(tt.tmpl, tt.args); got != tt.want {
				t.Fatalf("interpolate(%q) = %q, want %q", tt.tmpl, got, tt.want)
			}
		})
	}
}
