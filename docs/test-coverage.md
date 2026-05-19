# DocProcessor — Test Coverage Matrix

**Round 220 (2026-05-19)** — per-test-file coverage matrix enumerating the
seven test types CONST-050(B) demands for this module's domain (unit,
integration, E2E, automation, security, stress, Challenge), the file-system
location of each, the package or surface they cover, the kind of evidence
they capture, and the paired-mutation guard (per §1.1) that protects against
the test itself silently degrading into a bluff.

> Verbatim 2026-05-19 operator mandate (CONST-049 §11.4.17):
> *"all existing tests and Challenges do work in anti-bluff manner - they
> MUST confirm that all tested codebase really works as expected! ...
> execution of tests and Challenges MUST guarantee the quality, the
> completition and full usability by end users of the product!"*

## 1. Coverage matrix (real surfaces actually exercised)

| Test file                                            | Type        | Real surface exercised                                             | Captured evidence                                    | Paired mutation                                                              |
|------------------------------------------------------|-------------|--------------------------------------------------------------------|------------------------------------------------------|------------------------------------------------------------------------------|
| `pkg/loader/loader_test.go`                          | Unit        | `loader.NewDefaultLoader`, format detection, single-file load     | Loaded `Document` struct contents asserted          | Rename a supported format ext in `formats` slice → assertion fails           |
| `pkg/loader/loader_integration_test.go`              | Integration | `loader.LoadDir` against real temp dir tree                       | Real fs file count + per-doc body                   | Skip a file format in registry → loaded count drops                          |
| `pkg/loader/loader_security_test.go`                 | Security    | Path-traversal / symlink-escape / oversize-input rejection         | Rejected-path error message + no read attempted     | Remove path-canonicalisation step → traversal succeeds → test fails          |
| `pkg/loader/loader_stress_test.go`                   | Stress      | Concurrent `LoadDir` invocations, large fanout                    | No data race (`-race`) + per-iter timing            | Strip `sync.RWMutex` in loader → race detector trips                         |
| `pkg/loader/markdown.go` + parser tests              | Unit        | Markdown section parsing                                          | Parsed AST sections                                  | Truncate parser output → section count assertion fails                       |
| `pkg/feature/builder_test.go`                        | Unit        | `BuildFromDocs` heuristic extraction                              | Feature/Screen/Workflow counts + category map       | Remove 20-char heuristic gate → spurious features → count check fails        |
| `pkg/feature/feature_test.go`                        | Unit        | `Feature` struct invariants + equality                            | Equality + JSON round-trip                          | Drop a struct field → JSON round-trip diff fails                             |
| `pkg/feature/convert.go` tests                       | Unit        | Doc → feature conversion                                          | Converted feature contents                          | Skip category inference → category-map missing                               |
| `pkg/coverage/tracker_test.go`                       | Unit        | `CoverageTracker` set/get per-platform                            | Per-platform status map                              | Replace RWMutex with no-op → concurrent get returns stale state               |
| `pkg/coverage/tracker_stress_test.go`                | Stress      | 1000-goroutine concurrent set+get                                 | No data race + final state consistency              | Remove lock → race detector trips                                            |
| `pkg/docgraph/graph_test.go`                         | Unit        | `Graph.AddNode/AddEdge`, JSON + Mermaid export                    | Exported JSON + Mermaid strings                     | Skip Mermaid escaping → invalid syntax → diff fails                          |
| `pkg/docgraph/graph_stress_test.go`                  | Stress      | Concurrent node/edge insertion                                    | No data race + edge count                            | Strip lock → race detector trips                                             |
| `pkg/llm/prompts_test.go`                            | Unit        | Prompt template formatting                                        | Formatted prompt string                              | Drop placeholder → assertion fails                                           |
| `pkg/config/config_test.go`                          | Unit        | `.env` parsing into `Config`                                      | Parsed `Config` struct                               | Skip default-value injection → defaults missing                              |
| `pkg/config/config_security_test.go`                 | Security    | Reject path-traversal in `HELIX_DOCS_ROOT`                         | Rejection error message                              | Remove validation → traversal succeeds → test fails                          |
| `pkg/i18n/translator_test.go`                        | Unit        | `NoopTranslator.T` + `TPlural` verbatim-ID semantics              | Returned-string equality with msg ID                | Make Noop return `""` → equality fails                                       |
| `cmd/docprocessor/main_test.go`                      | Unit + E2E  | `runCLI` direct invocation, 8 branch tests, bundle integrity     | Sentinel translations + forbidden-literal absence   | Drop verbose gate / remove a sentinel → bundle test or branch test fails    |
| `e2e_test.go`                                        | E2E         | Built binary executed against real temp tree                      | Exit code + stdout/stderr captures                  | Replace binary with `/bin/true` → sentinel check fails                       |
| `automation_test.go`                                 | Automation  | Full doc-load → feature-build → coverage-track pipeline (real)    | Pipeline-output equality                            | Skip enrichment step → output diff fails                                     |
| `security_test.go`                                   | Security    | Reject symlink-escape, oversize args, malformed env                | Rejection error + no execution                      | Remove validation guard → test fails                                         |
| `challenges/scripts/docprocessor_cli_challenge.sh`   | Challenge   | Built binary against synthetic + empty docs dirs, both modes      | Captured wire output + exit code + paired mutation   | `--anti-bluff-mutate` strips assertions → script returns 99                  |

## 2. Anti-bluff floor (per CONST-035 / Article XI §11.9)

Every row above satisfies these four invariants:

1. **Positive captured evidence.** The test/Challenge writes the actual
   wire output, file contents, or struct state it asserts on (no
   "no error returned therefore PASS" patterns).
2. **Real-data exercise.** Non-unit rows touch real filesystems, real
   binaries, real parsers — no mock loaders, no mock binaries.
3. **Paired-mutation guard.** Every gate has a paired mutation (last
   column) that, when applied, makes the test FAIL — proving the
   assertion catches the negation per §1.1.
4. **CONST-046 wire-content check.** Every CLI test asserts the
   translated sentinel appears AND the forbidden English literal does
   NOT (preventing accidental hardcoded-string regression).

## 3. Coverage of CONST-050(A) — no fakes beyond unit tests

| Layer            | Permitted fakes        | DocProcessor compliance                                                |
|------------------|------------------------|-------------------------------------------------------------------------|
| Unit             | mocks, stubs allowed   | `fakeTranslator` lives in `cmd/docprocessor/main_test.go` (unit-only)  |
| Integration      | none                   | `loader_integration_test.go` writes real fs trees, no mocks            |
| E2E              | none                   | `e2e_test.go` invokes built binary, no mocks                           |
| Automation       | none                   | `automation_test.go` runs real pipeline end-to-end                     |
| Security         | none                   | `loader_security_test.go` / `config_security_test.go` use real inputs  |
| Stress           | none                   | `*_stress_test.go` use real concurrent goroutines + `-race`            |
| Challenge        | none                   | `challenges/scripts/*.sh` invoke the built binary                      |

Production code (`pkg/**/*.go` not ending `_test.go`, `cmd/**/main.go`) is
periodically swept for forbidden patterns (`grep -rn "simulated\|for now\|TODO
implement\|placeholder" pkg cmd`); the sweep returns empty as of round 220.

## 4. Coverage of CONST-046 — no hardcoded content

The CLI is the only user-facing surface in this module; its eighteen wire
strings are exhaustively gated by `cmd/docprocessor/main_test.go`:

- Eight branch tests assert each sentinel appears in the appropriate stream.
- Each branch test asserts the corresponding English literal does NOT appear
  (paired forbidden-literal check).
- `TestRunCLI_BundleContainsAllRound209MsgIDs` asserts every msg ID exists
  in `pkg/i18n/bundles/active.en.yaml` — planting a typo in the bundle key
  makes the test fail (paired mutation: rename `docprocessor_cli_done` to
  `docprocessor_cli_done2` → test fails).
- `TestRunCLI_NoopTranslator_EmitsMsgIDVerbatim` integrates the production
  `i18n.NoopTranslator` default to prove the verbatim-ID fallback path is
  itself captured (so absence-of-bundle is loudly visible per CONST-035).

## 5. How to add a new test that satisfies the matrix

1. Pick the layer (unit / integration / E2E / automation / security / stress / Challenge).
2. Place the file in the canonical location (see §1).
3. Capture positive runtime evidence (assert on actual struct contents,
   wire output, file system state — NOT on `err == nil`).
4. Add a paired-mutation note to the §1 row.
5. If the new test touches the CLI, add the new msg ID to both
   `pkg/i18n/bundles/active.en.yaml` AND the §1.4 list.
6. Update `cmd/docprocessor/main.go` package doc with the new round number
   and msg-ID rationale.
7. Run `go test ./... -race -count=1` AND the relevant Challenge script;
   confirm both green; capture the run output in your commit message.

## 6. Round-history of test-matrix growth

| Round | Date        | Net change                                                                |
|-------|-------------|---------------------------------------------------------------------------|
| 97    | 2026-04     | CONST-046 first migration — 8 CLI msg IDs externalised + tests           |
| 209   | 2026-05-19  | CONST-046 second migration — 10 additional CLI msg IDs + 7 new tests     |
| 220   | 2026-05-19  | docs/test-coverage.md + Challenge script (this document)                  |

## 7. References

- Root: [`CONSTITUTION.md`](../CONSTITUTION.md), [`CLAUDE.md`](../CLAUDE.md), [`AGENTS.md`](../AGENTS.md).
- Anti-bluff covenant: §11.4 of the canonical constitution submodule.
- Test-type coverage mandate: CONST-050(B).
- No-hardcoded-content mandate: CONST-046.
- Decoupling mandate: CONST-051(B).
