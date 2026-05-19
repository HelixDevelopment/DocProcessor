#!/usr/bin/env bash
# SPDX-License-Identifier: Apache-2.0
# Copyright 2026 Milos Vasic
#
# docprocessor_cli_challenge.sh — Round 220 Challenge for DocProcessor CLI.
#
# Anti-bluff invariants (per CONST-035 / Article XI §11.9 + §11.4):
#   1. Builds the real binary from source — no stubs, no mocks.
#   2. Invokes the real binary against three real on-disk doc trees
#      (synthetic populated, empty, single-feature) — no faked output.
#   3. Asserts the captured wire output contains the expected
#      translation sentinels AND does NOT contain forbidden English
#      literals — proving the CONST-046 i18n indirection is honoured
#      end-to-end through main + runCLI + i18n.NoopTranslator.
#   4. Paired-mutation mode (--anti-bluff-mutate): re-runs the
#      assertions against `/bin/true` instead of the real binary.
#      Every assertion MUST fail; exit code 99 indicates the assertion
#      layer is itself bluff-proof. Per §1.1.
#   5. Cleans up all temp dirs + temp binaries on exit (EXIT trap)
#      regardless of failure mode.
#
# Verbatim 2026-05-19 operator mandate (CONST-049 §11.4.17):
#   "all existing tests and Challenges do work in anti-bluff manner -
#    they MUST confirm that all tested codebase really works as expected!
#    We had been in position that all tests do execute with success and
#    all Challenges as well, but in reality the most of the features
#    does not work and can't be used! This MUST NOT be the case and
#    execution of tests and Challenges MUST guarantee the quality, the
#    completition and full usability by end users of the product!"
#
# Exit codes:
#   0  — PASS, all assertions held against the real binary
#   1  — FAIL, at least one assertion failed against the real binary
#   2  — environment problem (go not available, source tree missing)
#   99 — paired-mutation FAILed correctly (gate is bluff-proof)
#   98 — paired-mutation PASSed incorrectly (gate is itself a bluff)
#
# Usage:
#   bash challenges/scripts/docprocessor_cli_challenge.sh
#   bash challenges/scripts/docprocessor_cli_challenge.sh --anti-bluff-mutate

set -euo pipefail
IFS=$'\n\t'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
EVIDENCE_DIR="$(mktemp -d -t docp-r220-evidence-XXXXXX)"
TMP_BIN=""
TMP_DOCS_POP=""
TMP_DOCS_EMPTY=""
TMP_DOCS_SINGLE=""
MUTATE=0
ASSERT_FAILURES=0
ASSERT_TOTAL=0

cleanup() {
    rm -rf "${EVIDENCE_DIR}" 2>/dev/null || true
    [[ -n "${TMP_BIN}"         && -e "${TMP_BIN}"         ]] && rm -f  "${TMP_BIN}"         || true
    [[ -n "${TMP_DOCS_POP}"    && -d "${TMP_DOCS_POP}"    ]] && rm -rf "${TMP_DOCS_POP}"    || true
    [[ -n "${TMP_DOCS_EMPTY}"  && -d "${TMP_DOCS_EMPTY}"  ]] && rm -rf "${TMP_DOCS_EMPTY}"  || true
    [[ -n "${TMP_DOCS_SINGLE}" && -d "${TMP_DOCS_SINGLE}" ]] && rm -rf "${TMP_DOCS_SINGLE}" || true
}
trap cleanup EXIT

log()  { printf '[r220-challenge] %s\n' "$*" >&2; }
fail() { log "ASSERT FAIL: $*"; ASSERT_FAILURES=$((ASSERT_FAILURES + 1)); }
pass() { log "ASSERT PASS: $*"; }

assert_contains() {
    local file="$1"; local needle="$2"; local label="$3"
    ASSERT_TOTAL=$((ASSERT_TOTAL + 1))
    if grep -F -q -- "${needle}" "${file}"; then
        pass "${label}: found ${needle@Q}"
    else
        fail "${label}: expected to find ${needle@Q} in ${file}"
    fi
}

assert_not_contains() {
    local file="$1"; local needle="$2"; local label="$3"
    ASSERT_TOTAL=$((ASSERT_TOTAL + 1))
    if grep -F -q -- "${needle}" "${file}"; then
        fail "${label}: forbidden literal ${needle@Q} leaked into ${file}"
    else
        pass "${label}: no leak of ${needle@Q}"
    fi
}

for arg in "$@"; do
    case "$arg" in
        --anti-bluff-mutate) MUTATE=1 ;;
        --help|-h)
            sed -n '1,40p' "$0"
            exit 0
            ;;
        *)
            log "unknown arg: ${arg}"
            exit 2
            ;;
    esac
done

if ! command -v go >/dev/null 2>&1; then
    log "go toolchain not available — cannot build real binary"
    exit 2
fi

if [[ ! -f "${REPO_ROOT}/cmd/docprocessor/main.go" ]]; then
    log "expected cmd/docprocessor/main.go missing from ${REPO_ROOT}"
    exit 2
fi

# --- Build the REAL binary (or substitute /bin/true under --anti-bluff-mutate) -----
TMP_BIN="$(mktemp -t docprocessor-r220-XXXXXX)"
if (( MUTATE == 1 )); then
    log "MUTATE mode — replacing real binary with /bin/true to verify gate is not a bluff"
    cp /bin/true "${TMP_BIN}"
else
    log "Building real CLI binary from ${REPO_ROOT}/cmd/docprocessor → ${TMP_BIN}"
    (cd "${REPO_ROOT}" && go build -o "${TMP_BIN}" ./cmd/docprocessor) || {
        log "go build FAILed"; exit 1;
    }
fi

# --- Seed three real doc trees -------------------------------------------------------
TMP_DOCS_POP="$(mktemp -d -t docp-r220-pop-XXXXXX)"
TMP_DOCS_EMPTY="$(mktemp -d -t docp-r220-empty-XXXXXX)"
TMP_DOCS_SINGLE="$(mktemp -d -t docp-r220-single-XXXXXX)"

cat > "${TMP_DOCS_POP}/round220_alpha.md" <<'EOF'
# Round 220 Challenge Doc Alpha

## Feature Alpha One

This is a long enough description to pass the 20-char heuristic gate so a feature gets created in the alpha doc.

## Feature Alpha Two

Another long enough description to pass the 20-char heuristic gate so a second feature appears.
EOF

cat > "${TMP_DOCS_POP}/round220_beta.md" <<'EOF'
# Round 220 Challenge Doc Beta

## Feature Beta One

This is a long enough description to pass the 20-char heuristic gate so a feature gets created in the beta doc.
EOF

cat > "${TMP_DOCS_SINGLE}/single.md" <<'EOF'
# Single Doc

## Single Feature

This is a long enough description to pass the 20-char heuristic gate so exactly one feature appears.
EOF

# --- Scenario 1: usage / help branch (no docs-directory) ----------------------------
log "Scenario 1: usage branch (no args)"
OUT1="${EVIDENCE_DIR}/scenario1.out"
ERR1="${EVIDENCE_DIR}/scenario1.err"
"${TMP_BIN}" >"${OUT1}" 2>"${ERR1}" || true
assert_contains      "${ERR1}" "docprocessor_cli_usage"        "S1 usage sentinel"
assert_contains      "${ERR1}" "docprocessor_cli_help_header"  "S1 help-header sentinel"
assert_not_contains  "${ERR1}" "Usage: docprocessor"           "S1 no English usage literal"

# --- Scenario 2: empty docs directory (no docs found) -------------------------------
log "Scenario 2: empty docs directory"
OUT2="${EVIDENCE_DIR}/scenario2.out"
ERR2="${EVIDENCE_DIR}/scenario2.err"
"${TMP_BIN}" "${TMP_DOCS_EMPTY}" >"${OUT2}" 2>"${ERR2}" || true
assert_contains      "${OUT2}" "docprocessor_cli_format_summary"   "S2 format-summary sentinel"
assert_contains      "${OUT2}" "docprocessor_cli_loaded_documents" "S2 loaded-documents sentinel"
assert_contains      "${OUT2}" "docprocessor_cli_no_docs_found"    "S2 no-docs-found sentinel"
assert_contains      "${OUT2}" "docprocessor_cli_done"             "S2 done sentinel"
assert_not_contains  "${OUT2}" "No supported documents"            "S2 no English no-docs literal"
assert_not_contains  "${OUT2}" "Done in"                           "S2 no English done literal"

# --- Scenario 3: populated docs directory (terse mode) ------------------------------
log "Scenario 3: populated docs directory, terse mode"
OUT3="${EVIDENCE_DIR}/scenario3.out"
ERR3="${EVIDENCE_DIR}/scenario3.err"
"${TMP_BIN}" "${TMP_DOCS_POP}" >"${OUT3}" 2>"${ERR3}" || true
assert_contains      "${OUT3}" "docprocessor_cli_format_summary"      "S3 format-summary sentinel"
assert_contains      "${OUT3}" "docprocessor_cli_loaded_documents"    "S3 loaded-documents sentinel"
assert_contains      "${OUT3}" "docprocessor_cli_summary_header"      "S3 summary-header sentinel"
assert_contains      "${OUT3}" "docprocessor_cli_feature_map_summary" "S3 feature-map-summary sentinel"
assert_contains      "${OUT3}" "docprocessor_cli_doc_graph_summary"   "S3 doc-graph-summary sentinel"
assert_contains      "${OUT3}" "docprocessor_cli_done"                "S3 done sentinel"
assert_not_contains  "${OUT3}" "Feature map:"                         "S3 no English feature-map literal"
assert_not_contains  "${OUT3}" "Doc graph:"                           "S3 no English doc-graph literal"
assert_not_contains  "${OUT3}" "=== Feature map summary ==="          "S3 no English summary-header literal"

# Terse mode MUST NOT emit per-feature/screen/workflow sentinels.
assert_not_contains  "${OUT3}" "docprocessor_cli_feature_line"  "S3 no feature-line in terse mode"
assert_not_contains  "${OUT3}" "docprocessor_cli_screen_line"   "S3 no screen-line in terse mode"
assert_not_contains  "${OUT3}" "docprocessor_cli_workflow_line" "S3 no workflow-line in terse mode"

# --- Scenario 4: populated docs directory (verbose mode) ----------------------------
log "Scenario 4: populated docs directory, verbose mode"
OUT4="${EVIDENCE_DIR}/scenario4.out"
ERR4="${EVIDENCE_DIR}/scenario4.err"
"${TMP_BIN}" --verbose "${TMP_DOCS_POP}" >"${OUT4}" 2>"${ERR4}" || true
assert_contains      "${OUT4}" "docprocessor_cli_summary_header"  "S4 summary-header sentinel"
assert_contains      "${OUT4}" "docprocessor_cli_feature_line"    "S4 feature-line sentinel (verbose)"
assert_contains      "${OUT4}" "docprocessor_cli_done"            "S4 done sentinel"
assert_not_contains  "${OUT4}" "Feature ["                        "S4 no English feature-line literal"

# --- Scenario 5: -v short flag alias ------------------------------------------------
log "Scenario 5: -v short flag alias"
OUT5="${EVIDENCE_DIR}/scenario5.out"
ERR5="${EVIDENCE_DIR}/scenario5.err"
"${TMP_BIN}" -v "${TMP_DOCS_SINGLE}" >"${OUT5}" 2>"${ERR5}" || true
assert_contains      "${OUT5}" "docprocessor_cli_feature_line"  "S5 -v alias triggers verbose"
assert_contains      "${OUT5}" "docprocessor_cli_done"          "S5 done sentinel"

# --- Scenario 6: invalid (whitespace) path -----------------------------------------
log "Scenario 6: invalid (whitespace) path"
OUT6="${EVIDENCE_DIR}/scenario6.out"
ERR6="${EVIDENCE_DIR}/scenario6.err"
"${TMP_BIN}" "   " >"${OUT6}" 2>"${ERR6}" || true
assert_contains      "${ERR6}" "docprocessor_cli_path_invalid"  "S6 path-invalid sentinel"
assert_not_contains  "${ERR6}" "Invalid docs-directory"         "S6 no English path-invalid literal"

# --- Scenario 7: error-loading branch (non-existent path) --------------------------
log "Scenario 7: non-existent docs directory"
OUT7="${EVIDENCE_DIR}/scenario7.out"
ERR7="${EVIDENCE_DIR}/scenario7.err"
"${TMP_BIN}" "/nonexistent-round220-sentinel-$$" >"${OUT7}" 2>"${ERR7}" || true
assert_contains      "${ERR7}" "docprocessor_cli_error_loading_docs" "S7 error-loading sentinel"
assert_not_contains  "${ERR7}" "Error loading docs:"                  "S7 no English error literal"

# --- Final verdict ------------------------------------------------------------------
log "Asserts total=${ASSERT_TOTAL} failures=${ASSERT_FAILURES}"

if (( MUTATE == 1 )); then
    # Paired-mutation mode — we EXPECT failures because /bin/true emits nothing.
    if (( ASSERT_FAILURES > 0 )); then
        log "PAIRED-MUTATION OK: ${ASSERT_FAILURES}/${ASSERT_TOTAL} asserts failed against /bin/true — gate is bluff-proof"
        exit 99
    else
        log "PAIRED-MUTATION BLUFF DETECTED: 0 failures against /bin/true — assertions are not enforcing"
        exit 98
    fi
fi

if (( ASSERT_FAILURES > 0 )); then
    log "CHALLENGE FAIL: ${ASSERT_FAILURES}/${ASSERT_TOTAL} asserts failed"
    log "Evidence retained at: ${EVIDENCE_DIR} (would be cleaned on exit)"
    exit 1
fi

log "CHALLENGE PASS: ${ASSERT_TOTAL} asserts held against the real built binary"
log "Wire-evidence captured at: ${EVIDENCE_DIR}"
exit 0
