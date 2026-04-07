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

---

## What Changes After Migration

### Ownership

- Ownership of the organization moves to the NeoNephos Enterprise.
- The project assumes **full ownership** of the organization within the enterprise.
- The originating OSPO will no longer enforce policies on the organization.

### User Management

- Any automated user management (e.g., automated onboarding, group sync) linked to the previous enterprise or OSPO **will be unlinked**.
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

---

## Billing and CI

### Monthly Spending Limit

Organizations migrated to NeoNephos Enterprise are subject to a **$500/month billing cap** for:

- Private repository CI (GitHub Actions minutes)
- Custom/self-hosted runner usage billed through GitHub

This limit is in place due to current constraints on the enterprise budget and is subject to revision as additional budget becomes available.

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

Until the project adopts its own outbound and DCO/CLA processes (in any form), the processes previously enforced by the originating OSPO **remain in place and are the project's responsibility to maintain**.

---

## Migration Checklist

There is an issue template in this repository containing a check list.
---

## Summary of Restrictions

| Area                          | Restriction             | Reason                          |
|-------------------------------|-------------------------|---------------------------------|
| GHAS (private repos)          | Not available           | Cost control                    |
| Monthly CI spend              | $500/month cap          | Enterprise budget constraint    |
| Custom runners                | Counted toward $500 cap | Cost control                    |
| Enterprise policy enforcement | None currently          | Guidelines pending TAC approval |
| User management automation    | Not provided            | Org takes full ownership        |

---

## Contact

For questions or to request a budget review, open an issue in this repository or contact the NeoNephos enterprise administrators directly.
