---
title: "Steampipe Table: scaleway_iam_application - Query Scaleway IAM Applications using SQL"
description: "Allows users to query Scaleway IAM applications, providing details such as name, description, API key count and tags."
---

# Table: scaleway_iam_application - Query Scaleway IAM Applications using SQL

Scaleway IAM applications are non-human principals that own API keys and can be attached to IAM policies.

## Table Usage Guide

The `scaleway_iam_application` table lists the IAM applications of your organization. Use it to review which applications exist, how many API keys each has, and which are editable or managed by Scaleway.

**Important Notes**
- The `organization_id` config argument is optional; when set, results are limited to that organization.

## Examples

### Basic info
List the IAM applications of the organization.

```sql+postgres
select
  name,
  id,
  description,
  nb_api_keys,
  created_at
from
  scaleway_iam_application;
```

```sql+sqlite
select
  name,
  id,
  description,
  nb_api_keys,
  created_at
from
  scaleway_iam_application;
```

### List applications that have API keys
Find applications that currently own at least one API key.

```sql+postgres
select
  name,
  id,
  nb_api_keys
from
  scaleway_iam_application
where
  nb_api_keys > 0;
```

```sql+sqlite
select
  name,
  id,
  nb_api_keys
from
  scaleway_iam_application
where
  nb_api_keys > 0;
```

### List applications without API keys
Find applications that own no API key.

```sql+postgres
select
  name,
  id,
  created_at
from
  scaleway_iam_application
where
  nb_api_keys = 0;
```

```sql+sqlite
select
  name,
  id,
  created_at
from
  scaleway_iam_application
where
  nb_api_keys = 0;
```

### List user-managed applications
Exclude applications managed by Scaleway.

```sql+postgres
select
  name,
  id,
  editable
from
  scaleway_iam_application
where
  not managed;
```

```sql+sqlite
select
  name,
  id,
  editable
from
  scaleway_iam_application
where
  managed = 0;
```
