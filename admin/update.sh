#!/usr/bin/env bash
# Apply (or dry-run) a NeoNephos org's Peribolos configuration.
#
# This mirrors the kubernetes/org admin/update.sh pattern. It defaults to
# DRY-RUN; pass --confirm to actually mutate GitHub.
#
#   ./admin/update.sh neonephos-example                 # dry-run
#   ./admin/update.sh neonephos-example --confirm       # apply
#
# Requires: docker, and GITHUB_TOKEN in the environment (a token or GitHub App
# installation token with org admin scope).

set -o errexit
set -o nounset
set -o pipefail

ORG="${1:-}"
shift || true

if [[ -z "${ORG}" ]]; then
  echo "usage: $0 <org-name> [--confirm]" >&2
  exit 1
fi

if [[ -z "${GITHUB_TOKEN:-}" ]]; then
  echo "ERROR: GITHUB_TOKEN is not set" >&2
  exit 1
fi

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CONFIG_PATH="orgs/${ORG}/org.yaml"

if [[ ! -f "${REPO_ROOT}/${CONFIG_PATH}" ]]; then
  echo "ERROR: config not found: ${CONFIG_PATH}" >&2
  exit 1
fi

CONFIRM=""
for arg in "$@"; do
  if [[ "${arg}" == "--confirm" ]]; then
    CONFIRM="--confirm"
  fi
done

if [[ -z "${CONFIRM}" ]]; then
  echo ">>> DRY-RUN for org '${ORG}'. No changes will be made. Pass --confirm to apply." >&2
else
  echo ">>> APPLYING changes to org '${ORG}'." >&2
fi

PERIBOLOS_IMAGE="${PERIBOLOS_IMAGE:-gcr.io/k8s-prow/peribolos:v20250710-e2a6a9a3e}"
MIN_ADMINS="${MIN_ADMINS:-2}"
MAX_REMOVAL_DELTA="${MAX_REMOVAL_DELTA:-0.25}"
REQUIRE_SELF="${REQUIRE_SELF:-false}"

docker run --rm \
  -v "${REPO_ROOT}:/workspace:ro" \
  -w /workspace \
  -e GITHUB_TOKEN \
  "${PERIBOLOS_IMAGE}" \
  --config-path "${CONFIG_PATH}" \
  --fix-org --fix-org-members --fix-teams --fix-team-members \
  --min-admins="${MIN_ADMINS}" \
  --maximum-removal-delta="${MAX_REMOVAL_DELTA}" \
  --require-self="${REQUIRE_SELF}" \
  ${CONFIRM}
