# Peribolos: GitHub Org Member, Team & Permission Management

This repository manages **org membership, teams, and team permissions** for
organizations in the [NeoNephos GitHub Enterprise](https://github.com/enterprises/neonephos)
as code, using [Peribolos](https://docs.prow.k8s.io/docs/components/cli-tools/peribolos/)
— the same tool the Kubernetes project uses to manage its org at scale.

The model matches the migration guideline: **the enterprise owns the guardrails
(the automation, safety rails, and default policies), while each project
self-administers its own membership** via pull requests.

---

## How it works

```
orgs/<org>/org.yaml   ──PR──▶  review + CODEOWNERS approval  ──merge──▶  apply on GitHub
  (desired state)              (plan posted as PR comment)              (peribolos --confirm)
```

1. Each org has one file: `orgs/<org>/org.yaml` — the **single source of truth**
   for that org's settings, members, admins, teams, and team membership.
2. A PR that edits it triggers `peribolos-plan` — a **dry-run** that posts the
   exact planned changes as a PR comment. No mutations happen.
3. A CODEOWNER approves. On merge to `main`, `peribolos-apply` runs
   `peribolos --confirm` and reconciles GitHub to match the file.

> **Authoritative membership:** removing a user from `members`/`admins` in the
> YAML **removes them from the org** on apply. The `--maximum-removal-delta`
> safety rail (default 0.25) blocks any run that would delete >25% of
> memberships, guarding against typos.

---

## Repository layout

```
orgs/
  neonephos/
    org.yaml            # desired state for one org
Taskfile.yml             # validate / dump / plan / apply tasks
.taskrc.yml              # enables the env-precedence experiment
admin/
  update.sh             # local wrapper (dry-run by default)
.github/workflows/
  peribolos-plan.yml    # dry-run on PRs (pass/fail is the signal)
  peribolos-apply.yml   # apply on merge to main
CODEOWNERS              # who approves what
```

**Local prerequisites:** [go-task](https://taskfile.dev) (`task`), Docker, and a
`GITHUB_TOKEN` with org admin scope. `GITHUB_TOKEN` is a global Task variable
that defaults to your shell environment (`export GITHUB_TOKEN=...`) and can be
overridden inline (`task plan ORG=<org> GITHUB_TOKEN=...`); `.taskrc.yml` enables
Task's [env-precedence experiment](https://taskfile.dev/docs/experiments/env-precedence)
so an overridden value wins over a stale OS value. CI installs Task
automatically.

---

## One-time setup (enterprise admin)

### 1. Create the GitHub App

Peribolos authenticates as a **GitHub App** (preferred over a PAT: not tied to a
person, short-lived tokens, least privilege).

1. Create an App (at the enterprise or a central org). Permissions:
   - **Organization → Members: Read and write**
   - **Organization → Administration: Read and write**
2. Generate a private key (PEM).
3. **Install the App on each managed org.**
4. Store the credentials as repo (or org) secrets in this repository:
   - `PERIBOLOS_APP_ID`
   - `PERIBOLOS_APP_PRIVATE_KEY` (the PEM contents)

### 2. Protect `main`

- Require pull request reviews.
- Enable **Require review from Code Owners**.
- (Optional) Create a `peribolos-apply` **Environment** with the enterprise
  admins as **required reviewers** — this adds a manual approval gate before any
  mutation runs, on top of PR review.

### 3. Fix the CODEOWNERS placeholders

Replace `@neonephos/enterprise-admins` and the per-org owners in `CODEOWNERS`
with real teams/users.

---

## Onboarding a new org

1. **Seed from the live org** (captures current state so the first apply is a
   no-op):

   ```bash
   export GITHUB_TOKEN=...   # token/App token with org admin scope
   task dump ORG=<org> > orgs/<org>/org.yaml
   ```

2. **Trim** the dumped file — delete anything you don't want Peribolos to manage
   (e.g. `billing_email`), and review the member/admin/team lists.
3. Add the org to the `matrix.org` list in **both** workflow files and add a
   `CODEOWNERS` line for `orgs/<org>/`.
4. Open a PR. Confirm the `peribolos-plan` check passes and its job log shows
   **no destructive changes** (a freshly-dumped config should be a near no-op).
5. Merge. `peribolos-apply` reconciles.

---

## Everyday operations

### Add/remove a member or change a team (any contributor)

Edit `orgs/<org>/org.yaml` in a PR:

```yaml
orgs:
  my-org:
    members:
      - alice        # add a member
    teams:
      reviewers:
        members:
          - bob      # add bob to the reviewers team
        repos:
          my-repo: write
```

The plan check must pass; a CODEOWNER approves; merge applies it.

### Run a dry-run locally

```bash
export GITHUB_TOKEN=...
task plan ORG=<org>          # or: ./admin/update.sh <org>
```

### Apply locally (rarely needed — CI does this)

```bash
# Dry-run (default: CONFIRM=false — safe, mutates nothing)
task apply ORG=<org>

# Actually mutate GitHub
task apply ORG=<org> CONFIRM=true   # or: ./admin/update.sh <org> --confirm
```

---

## Safety rails (configured in the Taskfile)

| Flag | Default | Purpose |
| ------ | --------- | --------- |
| `--confirm` | `CONFIRM=false` | No mutations unless `CONFIRM=true`. Local `task apply` is a dry-run by default; the `peribolos-apply.yml` workflow passes `CONFIRM=true` on merge to `main`. |
| `--maximum-removal-delta` | `0.25` | Refuse runs deleting >25% of memberships (typo guard) |
| `--min-admins` | `2` | Refuse a config with fewer than 2 admins (lockout guard) |
| `--require-self` | `false` | If true, the bot must be an admin to apply |

Tune these via `task VAR=value ...` or environment variables (see `Taskfile.yml`).

---

## Config schema reference

`org.yaml` follows the Peribolos schema. Common fields:

```yaml
orgs:
  <org-name>:
    # org settings (omit any field to leave GitHub's current value untouched)
    default_repository_permission: read      # read|write|admin|none
    members_can_create_repositories: false

    admins:   [login1, login2]   # org owners (authoritative list)
    members:  [login3]           # org members (authoritative list)

    teams:
      <team-name>:
        description: ...
        privacy: closed          # closed (org-visible) | secret
        previously: [old-name]   # rename an existing team
        maintainers: [login1]
        members:     [login3]
        repos:                   # permission this team gets per repo
          repo-a: admin          # read|triage|write|maintain|admin
          repo-b: read
```

Full schema: <https://docs.prow.k8s.io/docs/components/cli-tools/peribolos/>

---

## Notes / limitations

- Peribolos manages **org settings, members, teams, and team membership**. It
  does **not** manage repository creation, branch protection, or rulesets. For
  centrally-enforced repo policy, pair it with
  [`github/safe-settings`](https://github.com/github/safe-settings) later.
- Usernames are GitHub login handles (lowercase, no `@`).
- The `peribolos` container image tag is pinned in the `Taskfile`
  (`PERIBOLOS_IMAGE`); bump it deliberately.
