# Organization Migration to NeoNephos Enterprise — Guideline

This document provides a detailed guide for migrating a GitHub organization into the [NeoNephos GitHub Enterprise](https://github.com/enterprises/neonephos). It is intended for enterprise administrators conducting the migration and for project maintainers receiving their organization under the enterprise umbrella.

---

## Overview

When an organization is transferred to the NeoNephos Enterprise, it becomes subject to enterprise-level infrastructure, billing, and governance. However, the project retains full operational independence and assumes ownership of day-to-day administration.

---

## Before You Start

Ensure the following prerequisites are met before initiating a migration:

- The project is onboarded to LFX and has reached **"Live" stage**.
- At least two enterprise contacts with GitHub accounts are identified to retain emergency owner access.
- The project maintainers (TSC or equivalent) have reviewed and acknowledged this guideline.
- A target migration date has been agreed upon.
- Verify that sufficient **enterprise seat capacity** exists. The seat count must be sufficient. If capacity is insufficient, request a seat increase from enterprise administrators before proceeding.


---

## What Changes After Migration

### Ownership

- Ownership of the organization moves to the NeoNephos Enterprise.
- The project assumes **full ownership** of the organization within the enterprise.
- The originating enterprise will no longer enforce policies on the organization.

### User Management


- Projects must establish their own user and group management processes.
- There is no automated onboarding mechanism provided by NeoNephos at this time.

### Enterprise Admin Access

- The enterprise contacts listed in the migration issue retain the right to **grant themselves owner access** on demand to address urgent policy violations or security incidents.
- Enterprise contacts will not interfere with routine organization setup or day-to-day administration.
- This access is reserved for emergency use only.

---

## GitHub Advanced Security (GHAS)

**GHAS is not available for private repositories** under the NeoNephos Enterprise after migration.

- This restriction exists to control enterprise billing costs.
- Code scanning, secret scanning (advanced), and dependency review for private repositories are not covered.
- Projects relying on GHAS for private repositories must make alternative arrangements (e.g., using public repositories where feasible, or procuring a separate GHAS license).
- GHAS features remain available for **public repositories** as part of GitHub's standard offering for open source projects.

### Why This Matters — Billing Impact

GitHub **Secret Protection** (part of Advanced Security) is billed **per active committer per month** on private repositories. This cost is incurred automatically when Advanced Security is enabled and can escalate rapidly — for an organization with a large number of members.

### How to Disable GHAS for Private Repos (Enterprise Admin)

After migration, enterprise administrators must immediately verify and configure GHAS:

1. Navigate to **Enterprise Settings → Billing and Licensing → Advanced Security**.
2. Select **"Manage and disable advanced security"**.
3. Disable Advanced Security for **private and internal repositories**.
4. Confirm that public repositories remain protected (this is free and automatic).

> **Important:** Monitor the enterprise billing dashboard for several days after migration to catch any unexpected charges. The billing for Secret Protection can start accruing immediately upon transfer.

---

## Billing and CI

### Monthly Spending Limit

Organizations migrated to NeoNephos Enterprise are subject to a **$500/month billing cap** for:

- Private repository CI (GitHub Actions minutes)
- Custom/self-hosted runner usage billed through GitHub

This limit is in place due to current constraints on the enterprise budget and is subject to revision as additional budget becomes available.

### Usage Budget

Each organization is assigned a **usage budget** at the enterprise level, which governs the total GitHub Actions and related compute allocation. This is configured during migration by the enterprise administrator. If your organization needs a higher usage budget, request it via the enterprise administrators.

### What This Means in Practice

- Actions minutes on public repositories are **free and unlimited** and do not count toward this cap.
- Only compute costs incurred by **private repository workflows** or **custom runners** are subject to the cap.
- Once the $500 cap is reached in a billing cycle, private repository CI jobs will be blocked until the next cycle or until the cap is raised.

### Recommendations for Managing CI Costs

- Audit all workflows in private repositories and remove or disable unused ones.
- Prefer **public repository workflows** for open source components wherever possible.
- Avoid non-standard (large or GPU) runners for routine CI tasks.
- Use caching aggressively (e.g., `actions/cache`) to reduce redundant compute.
- Review and right-size matrix builds — limit unnecessary combinations.
- Consider triggering expensive workflows only on release branches or manually, not on every push.

### Requesting a Budget Increase

If your project has legitimate need for more than $500/month in private CI spending, submit a request to the NeoNephos enterprise administrators with:

- Current and projected monthly spend breakdown.
- Justification for private repository usage (vs. public).
- Steps already taken to reduce costs.

Budget increases are reviewed on a case-by-case basis and are not guaranteed.

---

## Policies

At the time of migration, **no enterprise-level policies are enforced** on the organization beyond the billing cap and GHAS restriction described above.

Projects are encouraged to monitor the [NeoNephos Project Guidelines](https://github.com/neonephos/guidelines-development/blob/main/project-guidelines/project-guidelines.md) which are being developed for eventual TAC approval. Until formally approved by the TAC, these guidelines are not binding.

Areas expected to be covered by future guidelines include:

- Outbound contribution policies
- New repository creation processes
- DCO/CLA enforcement
- Code of Conduct requirements
- SPDX-compliant license identifiers

Until the project adopts its own outbound and DCO/CLA processes (in any form), the processes previously enforced by the originating enterprise **remain in place and are the project's responsibility to maintain**. The originating enterprise has committed to keeping their processes running until the foundation explicitly replaces or removes them.

### CLA and DCO Compliance

Projects migrating with an existing CLA or DCO setup (e.g., CLA Assistant with Developer Certificate of Origin) should **keep the existing setup in place during migration**. The CLA Assistant is open source and not tied to any specific enterprise, so it continues to function after transfer.

The foundation is evaluating replacement tooling, which may include:

- **Linux Foundation's EasyCLA**
- **DCO sign-off trailers** (as used by projects like OpenSearch)

Each project may decide on its own CLA/DCO approach in the interim. A foundation-wide decision is pending TAC approval.

---

## Post-Migration Verification (with Project owners)

After migration is complete, the following should be verified:

1. **Billing monitoring** — Monitor the enterprise billing dashboard daily for at least one week after migration. Watch for unexpected charges from Advanced Security, Actions minutes, or storage.
2. **GHAS configuration** — Confirm that Advanced Security is disabled for private repos and enabled for public repos.
3. **Public/private repo transitions** — Test what happens when a private repository is made public: verify that GHAS features (secret scanning, code scanning) auto-apply as expected. (One off task)
4. **User access** — Confirm that all members have access and that no automated sync from the previous enterprise is still running.
5. **CI workflows** — Verify that GitHub Actions workflows are running correctly and that spending is within the allocated budget.
6. **CLA/DCO tooling** — Confirm that any existing CLA or DCO enforcement (e.g., CLA Assistant) is still functioning on pull requests.

Report any post-migration issues to the originating enterprise contact or to the NeoNephos enterprise administrators.

---

## Lessons Learned from Past Migrations

The following observations come from early migrations into the NeoNephos Enterprise:

- **GHAS billing is immediate.** Secret Protection billing started accruing the moment organizations were transferred. The disable setting is not in the obvious location — it is under `Billing and Licensing → Advanced Security → Manage and disable advanced security`, not under the organization's security settings.
- **Seat count must be pre-configured.** Ensure the enterprise seat count accommodates all contributors across all organizations before migration. Organizations can have a lot of members each, with overlap between them.
- **Migrate one org at a time.** Transfer the first organization as a test case before proceeding with additional ones. This sequential approach allows issues (like the billing surprise) to be caught and resolved before the next transfer.
- **Keep the originating enterprise in the loop.** Having a representative from the originating enterprise on the migration call proved essential for troubleshooting enterprise-level settings.

---

## Migration Checklist

There is an issue template in this repository containing a check list.
---

## Summary of Restrictions

| Area                          | Restriction                         | Reason                          |
|-------------------------------|-------------------------------------|---------------------------------|
| GHAS (private repos)          | Not available                       | Cost control                    |
| Monthly CI spend              | $500/month cap                      | Enterprise budget constraint    |
| Custom runners                | Counted toward $500 cap             | Cost control                    |
| Enterprise policy enforcement | None currently                      | Guidelines pending TAC approval |
| User management automation    | Not provided                        | Org takes full ownership        |
| CLA/DCO tooling               | Existing setup retained temporarily | Foundation tooling pending TAC  |

---

## Contact

For questions or to request a budget review, open an issue in this repository or contact the NeoNephos enterprise administrators directly.
