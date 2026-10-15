---
title: "Steampipe Table: scaleway_iam_organization_security_settings - Query Scaleway IAM Organization Security Settings using SQL"
description: "Allows users to query the Scaleway IAM security settings of an organization."
---

# Table: scaleway_iam_organization_security_settings - Query Scaleway IAM Organization Security Settings using SQL

Scaleway IAM organization security settings define organization-wide policies such as API key lifetime, password renewal, login lockout and session duration.

## Table Usage Guide

The `scaleway_iam_organization_security_settings` table returns a single row with the security settings of the configured organization. Duration columns are expressed in seconds.

**Important Notes**
- This table requires the `organization_id` config argument to be set.
- `max_api_key_expiration_duration` is 0 when no maximum is enforced.

## Examples

### Basic info
Show the security settings of the organization.

```sql+postgres
select
  organization_id,
  max_api_key_expiration_duration,
  enforce_password_renewal,
  login_attempts_before_locked,
  max_login_session_duration,
  grace_period_duration
from
  scaleway_iam_organization_security_settings;
```

```sql+sqlite
select
  organization_id,
  max_api_key_expiration_duration,
  enforce_password_renewal,
  login_attempts_before_locked,
  max_login_session_duration,
  grace_period_duration
from
  scaleway_iam_organization_security_settings;
```

### Check whether API keys can be non-expiring
API keys have no enforced maximum lifetime when the duration is 0.

```sql+postgres
select
  organization_id,
  max_api_key_expiration_duration = 0 as api_keys_can_be_non_expiring
from
  scaleway_iam_organization_security_settings;
```

```sql+sqlite
select
  organization_id,
  max_api_key_expiration_duration = 0 as api_keys_can_be_non_expiring
from
  scaleway_iam_organization_security_settings;
```

### Check whether password renewal is enforced
Show whether password renewal is enforced and its grace period in days.

```sql+postgres
select
  organization_id,
  enforce_password_renewal,
  grace_period_duration / 86400 as grace_period_days
from
  scaleway_iam_organization_security_settings;
```

```sql+sqlite
select
  organization_id,
  enforce_password_renewal,
  grace_period_duration / 86400 as grace_period_days
from
  scaleway_iam_organization_security_settings;
```
