#!/bin/bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
# shellcheck source=scripts/install.sh
source "${ROOT_DIR}/scripts/install.sh"

TESTS_RUN=0
TMP_ROOT="$(mktemp -d)"
trap 'rm -rf "${TMP_ROOT}"' EXIT

fail() {
    echo "FAIL: $1" >&2
    exit 1
}

assert_success() {
    local name="$1"
    shift
    TESTS_RUN=$((TESTS_RUN + 1))
    if ! ("$@") >/dev/null 2>&1; then
        fail "${name}"
    fi
}

assert_failure() {
    local name="$1"
    shift
    TESTS_RUN=$((TESTS_RUN + 1))
    if ("$@") >/dev/null 2>&1; then
        fail "${name}"
    fi
}

assert_failure_contains() {
    local name="$1"
    local expected="$2"
    shift 2
    local output
    TESTS_RUN=$((TESTS_RUN + 1))
    if output=$( ("$@") 2>&1 ); then
        fail "${name}"
    fi
    if [[ "${output}" != *"${expected}"* ]]; then
        fail "${name}: expected output to contain '${expected}', got '${output}'"
    fi
}

with_fixture() {
    local checksum_line="${1-}"
    local with_archive="${2:-yes}"
    local with_checksums="${3:-yes}"
    local dir
    dir="$(mktemp -d "${TMP_ROOT}/case.XXXXXX")"
    if [ "${with_archive}" = "yes" ]; then
        printf 'release archive' > "${dir}/kkcli.tar.gz"
    fi
    if [ "${with_checksums}" = "yes" ]; then
        printf '%s\n' "${checksum_line}" > "${dir}/checksums.txt"
    fi
    printf '%s' "${dir}"
}

run_verify_checksum() {
    local dir="$1"
    TMP_DIR="${dir}"
    LATEST="v1.2.3"
    OS="linux"
    ARCH="amd64"
    verify_checksum
}

expected_checksum() {
    if command -v sha256sum >/dev/null 2>&1; then
        printf 'release archive' | sha256sum | awk '{print $1}'
    elif command -v shasum >/dev/null 2>&1; then
        printf 'release archive' | shasum -a 256 | awk '{print $1}'
    else
        return 1
    fi
}

test_matching_checksum_succeeds() {
    local checksum dir
    checksum="$(expected_checksum)"
    dir="$(with_fixture "${checksum}  kkcli_1.2.3_linux_amd64.tar.gz")"
    run_verify_checksum "${dir}"
}

test_missing_checksums_fails() {
    local dir
    dir="$(with_fixture '' yes no)"
    run_verify_checksum "${dir}"
}

test_missing_artifact_entry_fails() {
    local checksum dir
    checksum="$(expected_checksum)"
    dir="$(with_fixture "${checksum}  other.tar.gz")"
    run_verify_checksum "${dir}"
}

test_malformed_hash_fails() {
    local dir
    dir="$(with_fixture 'not-a-sha256  kkcli_1.2.3_linux_amd64.tar.gz')"
    run_verify_checksum "${dir}"
}

test_checksum_mismatch_fails() {
    local dir
    dir="$(with_fixture "$(printf '0%.0s' {1..64})  kkcli_1.2.3_linux_amd64.tar.gz")"
    run_verify_checksum "${dir}"
}

test_no_checksum_tool_fails() {
    local checksum dir tool_path real_path
    checksum="$(expected_checksum)"
    dir="$(with_fixture "${checksum}  kkcli_1.2.3_linux_amd64.tar.gz")"
    tool_path="$(mktemp -d "${TMP_ROOT}/tool-path.XXXXXX")"
    real_path="$(command -v awk)"
    ln -s "${real_path}" "${tool_path}/awk"
    real_path="$(command -v tr)"
    ln -s "${real_path}" "${tool_path}/tr"
    PATH="${tool_path}" run_verify_checksum "${dir}"
}

test_piped_installer_runs_main() {
    local output
    output="$({ printf 'main() { echo piped-main-ran; }\n'; printf '%s\n' 'if [[ "${BASH_SOURCE[0]:-}" == "$0" || -z "${BASH_SOURCE[0]:-}" ]]; then main "$@"; fi'; } | bash)"
    [[ "${output}" == "piped-main-ran" ]]
}

assert_success "matching checksum succeeds" test_matching_checksum_succeeds
assert_failure "missing checksums.txt fails" test_missing_checksums_fails
assert_failure "missing artifact entry fails" test_missing_artifact_entry_fails
assert_failure "malformed hash fails" test_malformed_hash_fails
assert_failure "checksum mismatch fails" test_checksum_mismatch_fails
assert_failure_contains "no checksum tool fails" "No checksum tool available" test_no_checksum_tool_fails
assert_success "piped installer runs main" test_piped_installer_runs_main

echo "PASS: ${TESTS_RUN} installer checksum tests"
