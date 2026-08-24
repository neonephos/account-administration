# Peribolos GitHub App

Peribolos authenticates to GitHub as a **GitHub App** (preferred over a PAT:
not tied to a person, short-lived tokens, least privilege). This directory
helps you create it.

## Create the App (one-time, ~5 min, browser required)

App creation and the private key can only be done in the browser — the private
key is shown exactly once and cannot be retrieved via API.

1. Open this URL (creates the App **owned by the `neonephos` org**):

   <https://github.com/organizations/neonephos/settings/apps/new>

2. Fill in the form with the settings below.

### Required settings

| Field | Value |
| ------- | ------- |
| **GitHub App name** | `neonephos-peribolos` |
| **Homepage URL** | `https://github.com/neonephos/account-administration` |
| **Webhook** | **Uncheck "Active"** (Peribolos doesn't need webhooks) |
| **Organization permissions → Members** | **Read and write** |
| **Organization permissions → Administration** | **Read and write** |
| **Where can this app be installed?** | Only on this account |

Leave all other permissions at "No access".

### After creation

1. Note the **Client ID** (shown on the App's settings page, looks like
   `Iv23li...`). GitHub now recommends the Client ID over the numeric App ID.
2. Click **Generate a private key** → downloads a `.pem` file (store securely,
   it is shown only once).
3. **Install the App**: App settings → *Install App* → install on the
   `neonephos` org (and any other org you onboard).
4. Register the credentials as repo secrets (see below).

## Register the secrets

Once you have the App ID and the `.pem` file, run from the repo root:

```bash
# Client ID (recommended by GitHub; looks like Iv23li...).
# The secret keeps the name PERIBOLOS_APP_ID for continuity; the workflows
# pass it to the action's `client-id` input.
gh secret set PERIBOLOS_APP_ID \
  --repo neonephos/account-administration \
  --body "Iv23li..."

# Private key (PEM). Point at the downloaded file.
gh secret set PERIBOLOS_APP_PRIVATE_KEY \
  --repo neonephos/account-administration \
  < ~/Downloads/neonephos-peribolos.YYYY-MM-DD.private-key.pem
```

The `peribolos` workflow reads these two
secrets to mint an installation token scoped to the target org.
