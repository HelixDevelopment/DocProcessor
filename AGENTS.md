# AGENTS.md

## MANDATORY: No CI/CD Pipelines

**NO GitHub Actions, GitLab CI/CD, or any automated pipeline may exist in this repository!**

- No `.github/workflows/` directory
- No `.gitlab-ci.yml` file
- No Jenkinsfile, .travis.yml, .circleci, or any other CI configuration
- All builds and tests are run manually or via Makefile targets
- This rule is permanent and non-negotiable

Instructions for AI coding agents working on DocProcessor.

## Project Context

DocProcessor is a Go module (`digital.vasic.docprocessor`) that processes project documentation into structured feature maps for QA automation. It is part of the HelixQA ecosystem alongside LLMOrchestrator, VisionEngine, and LLMsVerifier.

## Key Constraints

1. **Thread safety** -- CoverageTracker and DocGraph use sync.RWMutex. Never bypass locking.
2. **No test deletion** -- Tests must never be removed or disabled. Fix root causes.
3. **Race-safe** -- All tests must pass with `go test ./... -race -count=1`.
4. **Interface stability** -- LLMAgent, Loader, CoverageTracker, FeatureMapBuilder interfaces are public API.
5. **No circular imports** -- pkg/llm has no internal dependencies. pkg/feature depends on pkg/llm and pkg/docgraph.

## Adding New Functionality

1. Write tests first (TDD)
2. Implement the minimum code to pass tests
3. Run `go test ./... -race -count=1`
4. Run `go vet ./...`
5. Ensure SPDX headers on all new files

## Common Tasks

- **Add new document format**: Add parser in `pkg/loader/`, update `DefaultLoader.LoadFile()` switch
- **Add new feature category**: Add constant in `pkg/feature/feature.go`, update `AllCategories()`
- **Add new prompt template**: Add function in `pkg/llm/prompts.go`
- **Add coverage metric**: Add method to `CoverageTracker` interface and `tracker` implementation

<!-- BEGIN host-power-management addendum (CONST-033) -->

## Host Power Management — Hard Ban (CONST-033)

**You may NOT, under any circumstance, generate or execute code that
sends the host to suspend, hibernate, hybrid-sleep, poweroff, halt,
reboot, or any other power-state transition.** This rule applies to:

- Every shell command you run via the Bash tool.
- Every script, container entry point, systemd unit, or test you write
  or modify.
- Every CLI suggestion, snippet, or example you emit.

**Forbidden invocations** (non-exhaustive — see CONST-033 in
`CONSTITUTION.md` for the full list):

- `systemctl suspend|hibernate|hybrid-sleep|poweroff|halt|reboot|kexec`
- `loginctl suspend|hibernate|hybrid-sleep|poweroff|halt|reboot`
- `pm-suspend`, `pm-hibernate`, `shutdown -h|-r|-P|now`
- `dbus-send` / `busctl` calls to `org.freedesktop.login1.Manager.Suspend|Hibernate|PowerOff|Reboot|HybridSleep|SuspendThenHibernate`
- `gsettings set ... sleep-inactive-{ac,battery}-type` to anything but `'nothing'` or `'blank'`

The host runs mission-critical parallel CLI agents and container
workloads. Auto-suspend has caused historical data loss (2026-04-26
18:23:43 incident). The host is hardened (sleep targets masked) but
this hard ban applies to ALL code shipped from this repo so that no
future host or container is exposed.

**Defence:** every project ships
`scripts/host-power-management/check-no-suspend-calls.sh` (static
scanner) and
`challenges/scripts/no_suspend_calls_challenge.sh` (challenge wrapper).
Both MUST be wired into the project's CI / `run_all_challenges.sh`.

**Full background:** `docs/HOST_POWER_MANAGEMENT.md` and `CONSTITUTION.md` (CONST-033).

<!-- END host-power-management addendum (CONST-033) -->

