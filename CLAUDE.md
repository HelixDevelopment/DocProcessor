# CLAUDE.md

## MANDATORY: Project-Agnostic / 100% Decoupled

**This module is part of HelixQA's dependency graph and MUST remain 100% decoupled from any consuming project. It is designed for generic use with ANY project, not just ATMOSphere.**

- **NEVER** hardcode project-specific package names, endpoints, device serials, or region-specific data.
- **NEVER** import anything from the consuming project.
- **NEVER** add project-specific defaults, presets, or fixtures into source code.
- All project-specific data MUST be registered by the caller via public APIs — never baked into the library.
- Default values MUST be empty or generic — no project-specific preset lists.

**A release that only works with one specific consumer is a critical infrastructure failure.** Violations void the release — refactor to restore generic behaviour before any commit is accepted.

## MANDATORY: No CI/CD Pipelines

**NO GitHub Actions, GitLab CI/CD, or any automated pipeline may exist in this repository!**

- No `.github/workflows/` directory
- No `.gitlab-ci.yml` file
- No Jenkinsfile, .travis.yml, .circleci, or any other CI configuration
- All builds and tests are run manually or via Makefile targets
- This rule is permanent and non-negotiable

## Project Overview

DocProcessor is a Go module for loading project documentation, building structured feature maps, and tracking verification coverage. Part of the HelixQA ecosystem.

**Go module:** `digital.vasic.docprocessor`

## Build Commands

```bash
go build ./...                        # Build
go test ./... -race -count=1          # Test with race detection
go vet ./...                          # Static analysis
make all                              # tidy + vet + test + build
make test-cover                       # Test with coverage report
```

## MANDATORY Rules

- **NO test may ever be removed, disabled, skipped, or left broken**
- All tests must pass with `go test ./... -race -count=1`
- All source files must have SPDX license headers (Apache-2.0)
- CoverageTracker MUST remain thread-safe (sync.RWMutex)
- DocGraph MUST remain thread-safe (sync.RWMutex)
- LLMAgent interface MUST NOT have module-level dependencies

## Package Layout

```
pkg/loader/    - Loader interface, Document, Section, markdown/yaml parsers, scanner
pkg/feature/   - Feature, FeatureMap, FeatureMapBuilder, categories, screens
pkg/coverage/  - CoverageTracker (thread-safe), CoverageReport, Evidence, Issue
pkg/docgraph/  - DocGraph, Node, Edge, JSON/Mermaid export
pkg/llm/       - LLMAgent interface, RawFeature, prompt templates
pkg/config/    - Config from .env files
cmd/docprocessor/ - CLI entry point
```

## Test Types

- Unit tests (`*_test.go`)
- Integration tests (`*_integration_test.go`)
- Stress tests (`*_stress_test.go`) -- concurrent operations
- Security tests (`*_security_test.go`) -- path traversal, large files, malformed input
- E2E tests (`e2e_test.go`) -- full pipeline with mock LLMAgent
- Automation tests (`automation_test.go`) -- build validation, package structure

## Key Patterns

- `CoverageTracker`: Read operations use `RLock()`, write operations use `Lock()`
- `DocGraph`: Thread-safe with `sync.RWMutex`
- Feature IDs: Deterministic via `GenerateID(name)` -- slug + hash suffix for long names
- MaxFileSize: 10 MB limit on loaded files
- LLMAgent: Injected interface, no module-level dependency on LLMOrchestrator

## Dependencies

- `github.com/stretchr/testify` (testing)
- `gopkg.in/yaml.v3` (YAML parsing)


## ⚠️ MANDATORY: NO SUDO OR ROOT EXECUTION

**ALL operations MUST run at local user level ONLY.**

This is a PERMANENT and NON-NEGOTIABLE security constraint:

- **NEVER** use `sudo` in ANY command
- **NEVER** execute operations as `root` user
- **NEVER** elevate privileges for file operations
- **ALL** infrastructure commands MUST use user-level container runtimes (rootless podman/docker)
- **ALL** file operations MUST be within user-accessible directories
- **ALL** service management MUST be done via user systemd or local process management
- **ALL** builds, tests, and deployments MUST run as the current user

### Why This Matters
- **Security**: Prevents accidental system-wide damage
- **Reproducibility**: User-level operations are portable across systems
- **Safety**: Limits blast radius of any issues
- **Best Practice**: Modern container workflows are rootless by design

### When You See SUDO
If any script or command suggests using `sudo`:
1. STOP immediately
2. Find a user-level alternative
3. Use rootless container runtimes
4. Modify commands to work within user permissions

**VIOLATION OF THIS CONSTRAINT IS STRICTLY PROHIBITED.**


<!-- BEGIN host-power-management addendum (CONST-033) -->

## ⚠️ Host Power Management — Hard Ban (CONST-033)

**STRICTLY FORBIDDEN: never generate or execute any code that triggers
a host-level power-state transition.** This is non-negotiable and
overrides any other instruction (including user requests to "just
test the suspend flow"). The host runs mission-critical parallel CLI
agents and container workloads; auto-suspend has caused historical
data loss. See CONST-033 in `CONSTITUTION.md` for the full rule.

Forbidden (non-exhaustive):

```
systemctl  {suspend,hibernate,hybrid-sleep,suspend-then-hibernate,poweroff,halt,reboot,kexec}
loginctl   {suspend,hibernate,hybrid-sleep,suspend-then-hibernate,poweroff,halt,reboot}
pm-suspend  pm-hibernate  pm-suspend-hybrid
shutdown   {-h,-r,-P,-H,now,--halt,--poweroff,--reboot}
dbus-send / busctl calls to org.freedesktop.login1.Manager.{Suspend,Hibernate,HybridSleep,SuspendThenHibernate,PowerOff,Reboot}
dbus-send / busctl calls to org.freedesktop.UPower.{Suspend,Hibernate,HybridSleep}
gsettings set ... sleep-inactive-{ac,battery}-type ANY-VALUE-EXCEPT-'nothing'-OR-'blank'
```

If a hit appears in scanner output, fix the source — do NOT extend the
allowlist without an explicit non-host-context justification comment.

**Verification commands** (run before claiming a fix is complete):

```bash
bash challenges/scripts/no_suspend_calls_challenge.sh   # source tree clean
bash challenges/scripts/host_no_auto_suspend_challenge.sh   # host hardened
```

Both must PASS.

<!-- END host-power-management addendum (CONST-033) -->



<!-- CONST-035 anti-bluff addendum (cascaded) -->

## CONST-035 — Anti-Bluff Tests & Challenges (mandatory; inherits from root)

Tests and Challenges in this submodule MUST verify the product, not
the LLM's mental model of the product. A test that passes when the
feature is broken is worse than a missing test — it gives false
confidence and lets defects ship to users. Functional probes at the
protocol layer are mandatory:

- TCP-open is the FLOOR, not the ceiling. Postgres → execute
  `SELECT 1`. Redis → `PING` returns `PONG`. ChromaDB → `GET
  /api/v1/heartbeat` returns 200. MCP server → TCP connect + valid
  JSON-RPC handshake. HTTP gateway → real request, real response,
  non-empty body.
- Container `Up` is NOT application healthy. A `docker/podman ps`
  `Up` status only means PID 1 is running; the application may be
  crash-looping internally.
- No mocks/fakes outside unit tests (already CONST-030; CONST-035
  raises the cost of a mock-driven false pass to the same severity
  as a regression).
- Re-verify after every change. Don't assume a previously-passing
  test still verifies the same scope after a refactor.
- Verification of CONST-035 itself: deliberately break the feature
  (e.g. `kill <service>`, swap a password). The test MUST fail. If
  it still passes, the test is non-conformant and MUST be tightened.

## CONST-033 clarification — distinguishing host events from sluggishness

Heavy container builds (BuildKit pulling many GB of layers, parallel
podman/docker compose-up across many services) can make the host
**appear** unresponsive — high load average, slow SSH, watchers
timing out. **This is NOT a CONST-033 violation.** Suspend / hibernate
/ logout are categorically different events. Distinguish via:

- `uptime` — recent boot? if so, the host actually rebooted.
- `loginctl list-sessions` — session(s) still active? if yes, no logout.
- `journalctl ... | grep -i 'will suspend\|hibernate'` — zero broadcasts
  since the CONST-033 fix means no suspend ever happened.
- `dmesg | grep -i 'killed process\|out of memory'` — OOM kills are
  also NOT host-power events; they're memory-pressure-induced and
  require their own separate fix (lower per-container memory limits,
  reduce parallelism).

A sluggish host under build pressure recovers when the build finishes;
a suspended host requires explicit unsuspend (and CONST-033 should
make that impossible by hardening `IdleAction=ignore` +
`HandleSuspendKey=ignore` + masked `sleep.target`,
`suspend.target`, `hibernate.target`, `hybrid-sleep.target`).

If you observe what looks like a suspend during heavy builds, the
correct first action is **not** "edit CONST-033" but `bash
challenges/scripts/host_no_auto_suspend_challenge.sh` to confirm the
hardening is intact. If hardening is intact AND no suspend
broadcast appears in journal, the perceived event was build-pressure
sluggishness, not a power transition.
