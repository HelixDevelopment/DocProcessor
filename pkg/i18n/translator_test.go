// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Milos Vasic

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
	}
}
