---
page_title: "Scoped API Tokens and Service Accounts"
subcategory: ""
description: |-
  Use the provider with a scoped API token or a service account, and the token scopes it needs.
---

# Scoped API Tokens and Service Accounts

The provider accepts three kinds of credentials:

| Credentials | `email_address` | `token` | `is_scoped_token` |
|---|---|---|---|
| User with a classic API token | the user's email | classic (unscoped) API token | `false` (default) |
| User with a scoped API token | the user's email | scoped API token | `true` |
| Service account | `...@serviceaccount.atlassian.com` | the service account's scoped API token | `true` |

Classic API tokens keep working exactly as before; nothing in existing configurations needs to change.

## Enabling a scoped API token

Set `is_scoped_token = true` in the provider block, or set the `ATLASSIAN_OPS_IS_SCOPED_TOKEN=true` environment variable.

With a user's scoped API token:

```terraform
# A user's scoped API token
provider "atlassian-operations" {
  cloud_id        = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  domain_name     = "xxxx.atlassian.net"
  email_address   = "email@example.com"
  token           = "<YOUR_SCOPED_API_TOKEN>"
  is_scoped_token = true
}
```

With a service account:

```terraform
# A service account's scoped API token
provider "atlassian-operations" {
  cloud_id        = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  domain_name     = "xxxx.atlassian.net"
  email_address   = "my-bot@serviceaccount.atlassian.com"
  token           = "<SERVICE_ACCOUNT_SCOPED_API_TOKEN>"
  is_scoped_token = true
}
```

The rest of the configuration stays the same. Resources keep the same IDs and import formats, so an existing state also
keeps working after switching an existing configuration from a classic to a scoped API token.

## What changes with a scoped API token

The Teams REST API does not accept scoped API tokens. With `is_scoped_token = true` the provider therefore:

* manages teams (the `atlassian-operations_team` resource and data source) through the Teams GraphQL API
  (Atlassian GraphQL Gateway), and
* looks up users (the `atlassian-operations_user` data source) through `api.atlassian.com/ex/jira/{cloud_id}`.

All other resources use the JSM Operations REST API, which accepts both kinds of token.

~> **Note:** The Teams GraphQL mutations are in beta, so Atlassian can still change them without a deprecation notice.

## Required token scopes

Create the scoped API token with these scopes:

| Scope | Needed for |
|---|---|
| `read:jira-user` | Looking up users by email (`atlassian-operations_user` data source) |
| `read:account` | Account details of the token's owner and of team members |
| `view:team:teams` | Reading teams (`atlassian-operations_team` resource and data source) |
| `view:membership:teams` | Reading team members |
| `write:team:teams` | Creating, updating and deleting teams |
| `write:membership:teams` | Adding and removing team members |
| `read:ops-config:jira-service-management` | Reading JSM Operations configuration: schedules, rotations, escalations, routing rules, integrations, policies, heartbeats, maintenance windows |
| `write:ops-config:jira-service-management` | Creating and updating that configuration, and enabling Operations for a new team |
| `delete:ops-config:jira-service-management` | Deleting that configuration, including the defaults removed by `delete_default_resources` |
| `read:ops-alert:jira-service-management` | Alert-related JSM Operations APIs |
| `write:ops-alert:jira-service-management` | Alert-related JSM Operations APIs |
| `delete:ops-alert:jira-service-management` | Alert-related JSM Operations APIs |

`read:jira-user` is enough for user lookups; no extra Jira permission is needed.

## Service accounts

A service account authenticates with its email address (`...@serviceaccount.atlassian.com`) and a scoped API token
created for it, with the scopes above.

* **Team permissions.** To update or delete teams, add the service account to a group and give that group the
  **Manage teams** and **Delete teams** permissions on the team permissions page of the Teams product.
* **A human Operations admin.** The first entry in a team's `member` list becomes the team's Operations admin, so it
  must be a human JSM user, not a service account.
* **No on-call duty.** A service account cannot be a schedule rotation participant. Keep service accounts out of
  `participants`, otherwise the rotation fails with "A participant added to this rotation no longer exists".
* **Team types.** Service accounts can create `OPEN` and `MEMBER_INVITE` teams.

Not supported with service accounts:

* `atlassian-operations_user_contact`
* `atlassian-operations_custom_role`
* `atlassian-operations_maintenance` with `team_id` set (team maintenance); global maintenance windows work

## Differences in team attributes

* **`user_permissions`** is filled in from the Teams GraphQL `permissions` of the account the token belongs to. The
  GraphQL API has no separate permission for adding or removing members, so those follow `CAN_MANAGE_TEAM`:

  | `user_permissions` attribute | Is `true` when the account has |
  |---|---|
  | `update_team` | `CAN_MANAGE_TEAM` |
  | `add_members` | `CAN_MANAGE_TEAM` |
  | `remove_members` | `CAN_MANAGE_TEAM` |
  | `delete_team` | `CAN_DELETE_TEAM` |

  A service account gets `CAN_MANAGE_TEAM` and `CAN_DELETE_TEAM` from the Manage teams and Delete teams group
  permissions described above.

* **`organization_id`** is still required. The Teams GraphQL API does not use it; the provider keeps the configured
  value in state.
* **IDs** stay raw IDs (team IDs, account IDs, site IDs), exactly as with a classic API token. The provider converts
  them to and from the ARIs the Teams GraphQL API uses.

## Troubleshooting

| Error | Likely cause |
|---|---|
| `401 Unauthorized; scope does not match` on a team | `is_scoped_token` is not set, so the scoped token was sent to the Teams REST API; or the token is missing one of the scopes above |
| `401` when looking up a user | The token is missing `read:jira-user` |
| `A participant added to this rotation no longer exists` | A service account, or an ID that is not a raw account ID, is in the rotation's `participants` |
