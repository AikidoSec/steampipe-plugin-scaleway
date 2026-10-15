---
title: "Steampipe Table: scaleway_instance_security_group - Query Scaleway Instance Security Groups using SQL"
description: "Allows users to query Scaleway Instance Security Groups, providing insights into the configuration, state, and rules associated with each security group."
---

# Table: scaleway_instance_security_group - Query Scaleway Instance Security Groups using SQL

A Scaleway Instance Security Group is a virtual firewall that controls inbound and outbound traffic for one or more instances. It acts as a barrier between an instance and the rest of the network, allowing only traffic that matches the defined rules. These security groups are stateful, meaning that any outbound traffic that is permitted will automatically allow the corresponding inbound traffic.

## Table Usage Guide

The `scaleway_instance_security_group` table provides insights into the security groups within Scaleway Instance. As a security analyst, explore security group-specific details through this table, including their configuration, state, and associated rules. Utilize it to uncover information about security groups, such as those with overly permissive rules, the state of each security group, and the verification of inbound and outbound rules.

## Examples

### Basic info
Explore which security groups were created on a specific date within your Scaleway instance. This can help you identify instances where changes were made to the default project or specific zones, aiding in configuration review and management.

```sql+postgres
select
  name,
  id,
  creation_date,
  project_default,
  zone,
  project
from
  scaleway_instance_security_group;
```

```sql+sqlite
select
  name,
  id,
  creation_date,
  project_default,
  zone,
  project
from
  scaleway_instance_security_group;
```

### List default security groups
Explore which security groups are set as the default in your project to ensure correct configurations and prevent potential security risks. This can be particularly useful in managing access controls and maintaining secure project environments.

```sql+postgres
select
  name,
  id,
  creation_date,
  project_default,
  zone,
  project
from
  scaleway_instance_security_group
where
  project_default;
```

```sql+sqlite
select
  name,
  id,
  creation_date,
  project_default,
  zone,
  project
from
  scaleway_instance_security_group
where
  project_default = 1;
```

### List inbound accept rules open to the internet
Expand the `rules` column and find inbound accept rules that allow any source address. The column is fetched with one additional API call per security group.

```sql+postgres
select
  sg.name,
  sg.id,
  r ->> 'protocol' as protocol,
  r ->> 'dest_port_from' as dest_port_from,
  r ->> 'dest_port_to' as dest_port_to,
  r ->> 'ip_range' as ip_range
from
  scaleway_instance_security_group as sg,
  jsonb_array_elements(sg.rules) as r
where
  r ->> 'direction' = 'inbound'
  and r ->> 'action' = 'accept'
  and r ->> 'ip_range' in ('0.0.0.0/0', '::/0');
```

```sql+sqlite
select
  sg.name,
  sg.id,
  json_extract(r.value, '$.protocol') as protocol,
  json_extract(r.value, '$.dest_port_from') as dest_port_from,
  json_extract(r.value, '$.dest_port_to') as dest_port_to,
  json_extract(r.value, '$.ip_range') as ip_range
from
  scaleway_instance_security_group as sg,
  json_each(sg.rules) as r
where
  json_extract(r.value, '$.direction') = 'inbound'
  and json_extract(r.value, '$.action') = 'accept'
  and json_extract(r.value, '$.ip_range') in ('0.0.0.0/0', '::/0');
```
