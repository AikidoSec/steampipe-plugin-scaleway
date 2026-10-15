---
title: "Steampipe Table: scaleway_lb - Query Scaleway Load Balancers using SQL"
description: "Allows users to query Scaleway Load Balancers, providing details such as status, type, IP addresses, and attached frontends, backends and private networks."
---

# Table: scaleway_lb - Query Scaleway Load Balancers using SQL

Scaleway Load Balancers distribute incoming traffic across backend servers. Each load balancer has frontends, backends, and one or more IP addresses in a zone.

## Table Usage Guide

The `scaleway_lb` table lists load balancers across all configured zones. Use it to review load balancer status, type, attached IP addresses, and how many frontends, backends and private networks each one has.

## Examples

### Basic info
List the load balancers in your Scaleway projects with their status and location.

```sql+postgres
select
  name,
  id,
  status,
  type,
  zone,
  project_id
from
  scaleway_lb;
```

```sql+sqlite
select
  name,
  id,
  status,
  type,
  zone,
  project_id
from
  scaleway_lb;
```

### List load balancers with no backends
Identify load balancers that have no backend configured.

```sql+postgres
select
  name,
  id,
  zone,
  frontend_count,
  backend_count
from
  scaleway_lb
where
  backend_count = 0;
```

```sql+sqlite
select
  name,
  id,
  zone,
  frontend_count,
  backend_count
from
  scaleway_lb
where
  backend_count = 0;
```

### List load balancers with their IP addresses
Expand the attached IP addresses of each load balancer.

```sql+postgres
select
  l.name,
  l.zone,
  i ->> 'ip_address' as ip_address,
  i ->> 'reverse' as reverse
from
  scaleway_lb as l,
  jsonb_array_elements(l.ip) as i;
```

```sql+sqlite
select
  l.name,
  l.zone,
  json_extract(i.value, '$.ip_address') as ip_address,
  json_extract(i.value, '$.reverse') as reverse
from
  scaleway_lb as l,
  json_each(l.ip) as i;
```

### List load balancers attached to private networks
Find load balancers that have at least one private network attached.

```sql+postgres
select
  name,
  id,
  zone,
  private_network_count
from
  scaleway_lb
where
  private_network_count > 0;
```

```sql+sqlite
select
  name,
  id,
  zone,
  private_network_count
from
  scaleway_lb
where
  private_network_count > 0;
```
