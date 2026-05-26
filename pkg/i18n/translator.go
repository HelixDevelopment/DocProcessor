// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Milos Vasic

<<<<<<< HEAD
// Package i18n defines the Translator contract DocProcessor's CLI uses
// to externalise user-facing strings per CONST-046 (no-hardcoded-content
// mandate cascaded via constitution submodule §11.4.36).
//
// The package intentionally avoids any import of consumer-project paths
// (CONST-051(B) decoupling mandate) — DocProcessor stays standalone and
// reusable; any consuming project may supply its own Translator
// implementation that loads bundles, calls an LLM, or composes from
// verifier metadata at runtime.
=======
// Package i18n declares DocProcessor's hardcoded-content abstraction
// per CONST-046 (round-400 §11.4 anti-bluff sweep, 2026-05-19). It
// externalises every user-facing CLI string the docprocessor binary
// prints so a Serbian, Japanese, or Spanish operator receives a
// localised rendering instead of a verbatim English literal.
//
// The package intentionally avoids any import of consumer-project
// paths (CONST-051(B) decoupling mandate) — DocProcessor stays a
// standalone, project-not-aware, reusable Go module. A consuming
// project (or DocProcessor's own CLI) supplies a Translator: the
// bundle-backed BundleTranslator shipped here, an LLM-backed
// implementation, or anything satisfying the contract.
//
// The package-level NoopTranslator is the loud fallback — it returns
// the message ID verbatim so an absent / mis-keyed bundle is visible
// in captured CLI output rather than silently swallowed (which would
// be a §11.4 PASS-bluff at the i18n layer per Article XI §11.9).
>>>>>>> 9f5637d2d695cd5fcf8349d1f1b8bf780fa5d865
package i18n

import "context"

<<<<<<< HEAD
// Translator is the contract every i18n implementation must satisfy.
//
// T returns the localised rendering of msgID with named arguments
// substituted (`{{.key}}` style at the implementation's discretion).
//
// TPlural returns the localised rendering of msgID using plural-form
// resolution against count (CLDR Cardinal rules at the implementation's
// discretion).
type Translator interface {
	T(ctx context.Context, msgID string, args map[string]any) string
	TPlural(ctx context.Context, msgID string, count int, args map[string]any) string
}

// NoopTranslator is the default Translator returned when no other
// implementation is wired. It returns the message ID verbatim so the
// CLI remains functional in stripped-down environments (CI smoke
// builds, integration harnesses that exercise wire format only) and
// so absence-of-bundle is loudly visible in captured output.
//
// Per CONST-035 / Article XI §11.9 the verbatim-ID fallback is itself
=======
// Translator is the contract DocProcessor uses for every
// CONST-046-migrated user-facing string.
type Translator interface {
	// T resolves messageID against the active locale. args supplies
	// named placeholders for {{.key}}-style interpolation; pass nil
	// when the message has no placeholders.
	T(ctx context.Context, messageID string, args map[string]any) string

	// TPlural resolves messageID with plural-form selection driven
	// by count. args carries any non-count placeholders.
	TPlural(ctx context.Context, messageID string, count int, args map[string]any) string
}

// NoopTranslator returns the messageID verbatim. SAFETY default for
// unit tests and for callers that have not wired a real Translator.
// Per CONST-035 / Article XI §11.9 the verbatim-ID echo is itself
>>>>>>> 9f5637d2d695cd5fcf8349d1f1b8bf780fa5d865
// positive evidence — operators see exactly which key failed to
// resolve rather than an opaque empty string.
type NoopTranslator struct{}

<<<<<<< HEAD
// T returns msgID unchanged.
func (NoopTranslator) T(_ context.Context, msgID string, _ map[string]any) string {
	return msgID
}

// TPlural returns msgID unchanged.
func (NoopTranslator) TPlural(_ context.Context, msgID string, _ int, _ map[string]any) string {
	return msgID
=======
// T returns id unchanged (loud echo).
func (NoopTranslator) T(_ context.Context, id string, _ map[string]any) string {
	return id
}

// TPlural returns id unchanged (loud echo).
func (NoopTranslator) TPlural(_ context.Context, id string, _ int, _ map[string]any) string {
	return id
>>>>>>> 9f5637d2d695cd5fcf8349d1f1b8bf780fa5d865
}
