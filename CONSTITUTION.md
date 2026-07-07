# CONSTITUTION.md — Doc-Processor

## INHERITED FROM constitution/Constitution.md

All rules in `constitution/Constitution.md` (and the `constitution/Constitution.md` it references) apply unconditionally. This file's rules below extend them — they MUST NOT weaken any inherited rule. Use `constitution/find_constitution.sh` from the parent project root to resolve the absolute path of the submodule from any nested location.

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


---

## Constitutional Anti-Bluff Forensic Anchor (CONST-035 / §11.9, inherited)

> Verbatim user mandate: *"We had been in position that all tests do execute with success and all Challenges as well, but in reality the most of the features does not work and can't be used! This MUST NOT be the case and execution of tests and Challenges MUST guarantee the quality, the completion and full usability by end users of the product!"*
>
> Operative rule: **The bar for shipping is not "tests pass" but "users can use the feature."** Every PASS in this codebase MUST carry positive runtime evidence captured during execution. Metadata-only / configuration-only / absence-of-error / grep-based PASS without runtime evidence are critical defects regardless of how green the summary line looks. No false-success results are tolerable.

This anchor is inherited from the Helix Constitution (`constitution/Constitution.md` §11.9 / CONST-035); resolve it via `constitution/find_constitution.sh` from the parent project root. This submodule stays fully decoupled and project-not-aware (§11.4.28) — this is generic governance inheritance only, never project-specific context.
