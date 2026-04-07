# NeoNephos Account Administration

Declarative GitHub organization management using [github/safe-settings](https://github.com/github/safe-settings).

## How it works

YAML configuration in `.github/` defines the desired state. A GitHub Actions workflow runs safe-settings to reconcile the org on every push to `main` and on a 6-hour schedule.

| File | Purpose |
|------|---------|
| `.github/settings.yml` | Org-wide defaults (repo settings, labels, teams, branch protections, org rulesets) |
| `.github/repos/<name>.yml` | Per-repository overrides |
| `.github/suborgs/<name>.yml` | Group-level overrides (by repo list) |
| `.github/deployment-settings.yml` | Controls which repos are managed |

## Configuration hierarchy

Precedence (highest wins):

1. `repos/<name>.yml` — per-repo overrides
2. `suborgs/<name>.yml` — group overrides
3. `settings.yml` — org-wide defaults

## Making changes

1. Create a branch
2. Edit the YAML configuration
3. Open a pull request
4. Merge to `main` — sync runs automatically

Manual sync: Actions tab → "Safe Settings Sync" → "Run workflow"

## Setup

Requires a GitHub App installed on the `neonephos` org with the following repo secrets/variables:

| Type | Name | Description |
|------|------|-------------|
| Variable | `APP_ID` | GitHub App ID |
| Variable | `GH_ORG` | `neonephos` |
| Secret | `PRIVATE_KEY` | GitHub App private key (.pem contents) |

## What safe-settings manages

**Per-repo (applied to all repos):** repository settings, labels, milestones, teams, collaborators, branch protections, environments, autolinks, custom properties, variables

**Org-level:** rulesets only (branch/tag rules enforced across all repos)

**Not managed:** org profile, member privileges, default repo permissions, 2FA policy, Actions/Pages/Packages policies, enterprise settings. These must be configured manually or via Terraform.
