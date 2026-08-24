# NeoNephos Account Administration

Tooling and process for administering organizations in the
[NeoNephos GitHub Enterprise](https://github.com/enterprises/neonephos).

## Contents

- **Member, team & permission management** — managed as code with
  [Peribolos](https://docs.prow.k8s.io/docs/components/cli-tools/peribolos/).
  Each org's membership lives in `orgs/<org>/org.yaml`; changes go through a PR
  that dry-runs the plan, and apply happens on merge.
  See [docs/peribolos-setup.md](docs/peribolos-setup.md).
- **Org migration** — guideline and issue template for transferring an
  organization into the enterprise. See
  [docs/migration-guideline.md](docs/migration-guideline.md).

## Quick start

```bash
# Dry-run what would change for an org (needs GITHUB_TOKEN with org admin scope)
make plan ORG=neonephos

# Validate all org configs (no network)
make validate
```

Managing an org's members or teams is a pull request against
`orgs/<org>/org.yaml`. The full workflow, GitHub App setup, safety rails, and
schema are documented in [docs/peribolos-setup.md](docs/peribolos-setup.md).

## Layout

```
orgs/<org>/org.yaml         # source of truth: members, admins, teams, permissions
Makefile                    # validate / dump / plan / apply
admin/update.sh             # local wrapper (dry-run by default)
.github/workflows/          # peribolos-plan (PR dry-run) + peribolos-apply (on merge)
CODEOWNERS                  # approval gates
docs/                       # setup + migration guides
```
