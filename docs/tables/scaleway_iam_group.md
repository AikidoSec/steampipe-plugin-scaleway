---
title: "Steampipe Table: scaleway_iam_group - Query Scaleway IAM Groups using SQL"
description: "Allows users to query Scaleway IAM groups, providing details such as name, description, member users and applications, and tags."
---

# Table: scaleway_iam_group - Query Scaleway IAM Groups using SQL

Scaleway IAM groups are collections of users and applications that can be attached to IAM policies.

## Table Usage Guide

The `scaleway_iam_group` table lists the IAM groups of your organization. Use it to review which groups exist and which users and applications are members of each.

**Important Notes**
- The `organization_id` config argument is optional; when set, results are limited to that organization.

## Examples

### Basic info
List the IAM groups of the organization.

```sql+postgres
select
  name,
  id,
  description,
  created_at
from
  scaleway_iam_group;
```

```sql+sqlite
select
  name,
  id,
  description,
  created_at
from
  scaleway_iam_group;
```

### List groups without members
Find groups that have no user or application as member.

```sql+postgres
select
  name,
  id,
  created_at
from
  scaleway_iam_group
where
  jsonb_array_length(coalesce(user_ids, '[]'::jsonb)) = 0
  and jsonb_array_length(coalesce(application_ids, '[]'::jsonb)) = 0;
```

```sql+sqlite
select
  name,
  id,
  created_at
from
  scaleway_iam_group
where
  json_array_length(coalesce(user_ids, '[]')) = 0
  and json_array_length(coalesce(application_ids, '[]')) = 0;
```

### List the members of each group
Expand the user members of every group.

```sql+postgres
select
  g.name,
  u.value as user_id
from
  scaleway_iam_group as g,
  jsonb_array_elements_text(g.user_ids) as u(value);
```

```sql+sqlite
select
  g.name,
  u.value as user_id
from
  scaleway_iam_group as g,
  json_each(g.user_ids) as u;
```
