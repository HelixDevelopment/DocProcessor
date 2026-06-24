# CONSTITUTION.md — Doc-Processor

## INHERITED FROM constitution/Constitution.md

All rules in `constitution/Constitution.md` (and the `constitution/Constitution.md` it references) apply unconditionally. This file's rules below extend them — they MUST NOT weaken any inherited rule. See parent root `CLAUDE.md` §6.AD for the Lava-specific incorporation context (29th §6.L cycle, 2026-05-14) and §6.AD-debt for the implementation-gap inventory. Use `constitution/find_constitution.sh` from the parent project root to resolve the absolute path of the submodule from any nested location.

## INHERITED FROM the Helix Constitution

This module is governed by the Helix Constitution. All rules in the
constitution's `Constitution.md` and the `Constitution.md` it references apply
unconditionally. Locate the constitution from any nested depth via its
`find_constitution.sh` helper — do NOT hardcode a path (this module stays
fully decoupled and project-agnostic per §11.4.28).

Canonical reference: https://github.com/HelixDevelopment/HelixConstitution

## Mission

DocProcessor turns project documentation into structured, queryable feature
maps and verification-coverage data for QA automation, while remaining a
standalone, project-not-aware, fully decoupled Go module
(`digital.vasic.docprocessor`).

## Module-Specific Mandatory Standards

This module extends the Helix Constitution; it adds no overrides that relax any
canonical rule. The following are concrete, code-enforced restatements scoped to
this module:

- Decoupling: no code under `pkg/**` or `cmd/**` may import a consuming
  project's namespace. Project-specific behaviour is injected only via the
  `Translator`, `LLMAgent`, and `Config` contracts.
- No hardcoded user-facing content: every CLI line is emitted through the
  injected `i18n.Translator`; message IDs live in
  `pkg/i18n/bundles/active.en.yaml`. `NoopTranslator` is a loud fallback that
  returns the message ID verbatim.
- Secrets hygiene: `.env` is git-ignored and must be `chmod 600`; only
  `.env.example` (placeholder values) is committed.
- Quality gates: `go build ./...` and `go vet ./...` must pass with zero
  warnings, and the full suite (`go test ./... -race`) — unit, integration,
  stress, security, E2E, automation — must pass against real infrastructure
  (filesystem and the real CLI binary), never via stubbed or skipped tests.
