# Peribolos automation for NeoNephos account administration.
#
# Peribolos manages GitHub org settings, teams, and memberships from the YAML
# files under orgs/*/org.yaml. See docs/peribolos-setup.md.
#
# Usage:
#   make validate                 # lint + schema-check all org configs (no network)
#   make dump ORG=neonephos > orgs/neonephos/org.yaml   # seed from live org
#   make plan  ORG=neonephos   # dry-run: show what WOULD change (no mutations)
#   make apply ORG=neonephos   # APPLY changes to GitHub (requires --confirm)
#
# Auth: set GITHUB_TOKEN (a token or GitHub App installation token with org
# admin scope). In CI this is provided by the GitHub App (see workflows).

# Pin the Prow image tag; bump deliberately.
# Registry: us-docker.pkg.dev/k8s-infra-prow/images (gcr.io/k8s-prow was retired).
# Tags: https://us-docker.pkg.dev/k8s-infra-prow/images/peribolos
PERIBOLOS_IMAGE ?= us-docker.pkg.dev/k8s-infra-prow/images/peribolos:v20260821-a61940897
CONFIG_DIR      ?= orgs
ORG             ?=

# Safety rails (mirror the kubernetes/org defaults).
MIN_ADMINS          ?= 2
MAX_REMOVAL_DELTA   ?= 0.25
REQUIRE_SELF        ?= false

CONFIG_PATH := $(CONFIG_DIR)/$(ORG)/org.yaml

# Run peribolos via Docker so no local Go toolchain is needed.
# Mounts the repo read-only. peribolos reads the token from a FILE
# (--github-token-path), not an env var, so we override the entrypoint to write
# $GITHUB_TOKEN into a temp file inside the container, then exec peribolos.
define RUN_PERIBOLOS
	docker run --rm \
		-v "$(CURDIR):/workspace:ro" \
		-w /workspace \
		-e GITHUB_TOKEN \
		--entrypoint sh \
		$(PERIBOLOS_IMAGE) \
		-c 'umask 077; printf %s "$$GITHUB_TOKEN" > /tmp/token; exec /ko-app/peribolos --github-token-path /tmp/token "$$@"' sh
endef

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}'

.PHONY: check-org
check-org:
	@test -n "$(ORG)" || { echo "ERROR: set ORG=<org-name> (e.g. make plan ORG=neonephos)"; exit 1; }
	@test -f "$(CONFIG_PATH)" || { echo "ERROR: config not found: $(CONFIG_PATH)"; exit 1; }

.PHONY: check-token
check-token:
	@test -n "$(GITHUB_TOKEN)" || { echo "ERROR: GITHUB_TOKEN is not set"; exit 1; }

.PHONY: validate
validate: ## Validate all org configs are well-formed YAML (no network calls)
	@for f in $(CONFIG_DIR)/*/org.yaml; do \
		echo "checking $$f"; \
		python3 -c "import sys,yaml; yaml.safe_load(open('$$f')); print('  ok')" || exit 1; \
	done

.PHONY: dump
dump: check-org check-token ## Dump the live org config to stdout (seed a new org.yaml)
	@$(RUN_PERIBOLOS) --dump "$(ORG)"

.PHONY: plan
plan: check-org check-token ## Dry-run: print the changes peribolos WOULD make (no mutations)
	$(RUN_PERIBOLOS) \
		--config-path "$(CONFIG_PATH)" \
		--fix-org --fix-org-members --fix-teams --fix-team-members \
		--min-admins=$(MIN_ADMINS) \
		--maximum-removal-delta=$(MAX_REMOVAL_DELTA) \
		--require-self=$(REQUIRE_SELF)

.PHONY: apply
apply: check-org check-token ## APPLY the config to GitHub (mutates! adds --confirm)
	$(RUN_PERIBOLOS) \
		--config-path "$(CONFIG_PATH)" \
		--fix-org --fix-org-members --fix-teams --fix-team-members \
		--min-admins=$(MIN_ADMINS) \
		--maximum-removal-delta=$(MAX_REMOVAL_DELTA) \
		--require-self=$(REQUIRE_SELF) \
		--confirm
