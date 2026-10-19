---
title: "Steampipe Table: scaleway_iam_policy - Query Scaleway IAM Policies using SQL"
description: "Allows users to query Scaleway IAM policies, providing details such as the principal the policy is attached to, rule counts and the rules themselves."
---

# Table: scaleway_iam_policy - Query Scaleway IAM Policies using SQL

Scaleway IAM policies grant permissions to a principal (user, group or application) through a set of rules. Each rule lists permission sets and the scope they apply to.

## Table Usage Guide

The `scaleway_iam_policy` table lists the IAM policies of your organization. A policy is attached to exactly one of `user_id`, `group_id` or `application_id`, or to none when `no_principal` is true. The `rules` column is fetched with one additional API call per policy, so select it only when needed.

**Important Notes**
- The `organization_id` config argument is optional; when set, results are limited to that organization.
- `rules` is a JSON array of objects with the keys `id`, `permission_set_names`, `permission_sets_scope_type`, `condition`, `project_ids` and `organization_id`.

## Examples

### Basic info
List the IAM policies and how many rules each contains.

```sql+postgres
select
  name,
  id,
  nb_rules,
  nb_permission_sets,
  editable,
  managed
from
  scaleway_iam_policy;
```

```sql+sqlite
select
  name,
  id,
  nb_rules,
  nb_permission_sets,
  editable,
  managed
from
  scaleway_iam_policy;
```

### List policies attached to an application
Show which policies are attached to applications.

```sql+postgres
select
  name,
  id,
  application_id
from
  scaleway_iam_policy
where
  application_id is not null;
```

```sql+sqlite
select
  name,
  id,
  application_id
from
  scaleway_iam_policy
where
  application_id is not null;
```

### List policies that grant a full-access permission set
Expand the rules and find permission sets whose name ends with FullAccess.

```sql+postgres
select
  p.name,
  p.id,
  r ->> 'id' as rule_id,
  r -> 'permission_set_names' as permission_set_names
from
  scaleway_iam_policy as p,
  jsonb_array_elements(p.rules) as r,
  jsonb_array_elements_text(r -> 'permission_set_names') as ps
where
  ps like '%FullAccess';
```

```sql+sqlite
select
  p.name,
  p.id,
  json_extract(r.value, '$.id') as rule_id,
  json_extract(r.value, '$.permission_set_names') as permission_set_names
from
  scaleway_iam_policy as p,
  json_each(p.rules) as r,
  json_each(json_extract(r.value, '$.permission_set_names')) as ps
where
  ps.value like '%FullAccess';
```

### List rules scoped to the whole organization
Find rules whose scope is the organization rather than specific projects.

```sql+postgres
select
  p.name,
  p.id,
  r ->> 'permission_sets_scope_type' as scope_type
from
  scaleway_iam_policy as p,
  jsonb_array_elements(p.rules) as r
where
  r ->> 'permission_sets_scope_type' = 'organization';
```

```sql+sqlite
select
  p.name,
  p.id,
  json_extract(r.value, '$.permission_sets_scope_type') as scope_type
from
  scaleway_iam_policy as p,
  json_each(p.rules) as r
where
  json_extract(r.value, '$.permission_sets_scope_type') = 'organization';
```
