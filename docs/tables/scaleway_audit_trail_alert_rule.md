---
title: "Steampipe Table: scaleway_audit_trail_alert_rule - Query Scaleway Audit Trail Alert Rules using SQL"
description: "Allows users to query Scaleway Audit Trail alert rules, providing details such as name, description and status."
---

# Table: scaleway_audit_trail_alert_rule - Query Scaleway Audit Trail Alert Rules using SQL

Scaleway Audit Trail records actions performed on an organization's resources. Alert rules define which audit events raise an alert.

## Table Usage Guide

The `scaleway_audit_trail_alert_rule` table lists the Audit Trail alert rules across all configured regions.

## Examples

### Basic info
List the alert rules and their status.

```sql+postgres
select
  name,
  id,
  status,
  region
from
  scaleway_audit_trail_alert_rule;
```

```sql+sqlite
select
  name,
  id,
  status,
  region
from
  scaleway_audit_trail_alert_rule;
```

### List alert rules that are not enabled
Find alert rules whose status is anything other than enabled.

```sql+postgres
select
  name,
  id,
  status,
  region
from
  scaleway_audit_trail_alert_rule
where
  status <> 'enabled';
```

```sql+sqlite
select
  name,
  id,
  status,
  region
from
  scaleway_audit_trail_alert_rule
where
  status <> 'enabled';
```

### Count alert rules by region
Show how many alert rules exist per region.

```sql+postgres
select
  region,
  count(*) as rule_count
from
  scaleway_audit_trail_alert_rule
group by
  region;
```

```sql+sqlite
select
  region,
  count(*) as rule_count
from
  scaleway_audit_trail_alert_rule
group by
  region;
```
