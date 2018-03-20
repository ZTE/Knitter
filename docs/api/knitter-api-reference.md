# Knitter NetWork API Reference

[TOC]

## Network Operations
This section shows an example for the request of each network operation and its possible response.
Following the example, the description of the operation is provided.    

#####  1. Create network
Request:
```bash
curl "http://127.0.0.1:9527/nw/v1/tenants/user_foo/networks" -H Content-Type:application/json -X POST -d '{"network": {"name": "network_bar","public":false, "gateway": "123.124.125.1", "cidr": "123.124.125.0/24"}}' | python -m json.tool
```
N.B.: `python -m json.tool` is just for formatting the response. It's not required for the command. 
Response:
```json
{
    "network": {
        "cidr": "123.124.125.0/24",
        "create_time": "2018-03-01T07:56:02Z",
        "description": "",
        "gateway": "123.124.125.1",
        "name": "network_bar",
        "public": false,
        "owner": "user_foo",
        "network_id": "98260ff6-4c47-4c2e-86a5-e18d71a09d6e",
        "state": "ACTIVE"
    }
}
```
    Description : create network for the specified tenant.
    Method      : POST
    Path        : /nw/v1/tenants/{user}/networks
    Input       :
        {user}    tenant name
        name      network name (optional)
        public    whether the network is public/shared (optional, default to false)
        gate-way  gateway (required)
        cidr      CIDR (required)
    Return code :
        Success : 200
        Failure : other code

#####  2. Show network
Request:
```bash
curl "http://127.0.0.1:9527/nw/v1/tenants/user_foo/networks/440d2691-fbc6-43c1-9a33-3d7c9cf91432" -XGET | python -m json.tool
```
Response:
```json
{
    "network": {
        "cidr": "123.124.125.0/24",
        "create_time": "2018-03-01T08:12:13Z",
        "description": "",
        "gateway": "123.124.125.1",
        "name": "network_bar",
        "public": false,
        "owner": "user_foo",
        "network_id": "440d2691-fbc6-43c1-9a33-3d7c9cf91432",
        "state": "ACTIVE"
    }
}
```
    Description : get network by UUID in the specified tenant
    Method      : GET
    Path        : nw/v1/tenants/{user}/networks/{network_uuid}
    Input       :
        {user}            tenant name
        {network_uuid}    network UUID
    Return code :
        Success : 200
        Failure : other code

#####  3. List network
Request:
```bash
curl "http://127.0.0.1:9527/nw/v1/tenants/user_foo/networks" -XGET | python -m json.tool
```
Response:
```json
{
    "networks": [
        {
            "cidr": "123.124.125.0/24",
            "create_time": "2018-03-01T07:56:02Z",
            "description": "",
            "gateway": "123.124.125.1",
            "name": "network_bar",
            "network_id": "98260ff6-4c47-4c2e-86a5-e18d71a09d6e",
            "state": "ACTIVE"
        },
        {
            "cidr": "123.124.125.0/24",
            "create_time": "2018-03-01T08:12:13Z",
            "description": "",
            "gateway": "123.124.125.1",
            "name": "provider_network_bar",
            "network_id": "440d2691-fbc6-43c1-9a33-3d7c9cf91432",
            "state": "ACTIVE"
        }
    ]
}
```
    Description : list all of the networks in specified tenant
    Method      : GET
    Path        : nw/v1/tenants/{user}}/networks
    Input       :
        {user}    tenant name
    Return code :
        Success : 200
        Failure : other code

#####  4. Get network by name
Request:
```bash
curl "http://127.0.0.1:9527/nw/v1/tenants/user_foo/networks?name=network_bar" -XGET | python -m json.tool
```
Response:
```json
{
    "networks": [
        {
            "cidr": "123.124.125.0/24",
            "create_time": "2018-03-01T07:56:02Z",
            "description": "",
            "gateway": "123.124.125.1",
            "name": "network_bar",
            "network_id": "98260ff6-4c47-4c2e-86a5-e18d71a09d6e",
            "state": "ACTIVE"
        }
    ]
}
```
    Description : get network by name in the specified tenant
    Method      : GET
    Path        : nw/v1/tenants/{user}}/networks
    Input       :
        {user}    tenant name
        name      network name
    Return code :
        Success : 200
        Failure : other code

#####  5. Delete network
Request:
```bash
curl "http://127.0.0.1:9527/nw/v1/tenants/user_foo/networks/98260ff6-4c47-4c2e-86a5-e18d71a09d6e" -XDELETE
```
Response:
```json
null
```
    Description : delete network by uuid in the specified tenant
    Method      : DELETE
    Path        : nw/v1/tenants/{user}/networks/{network_uuid}
    Input       :
        {user}            tenant name
        {network_uuid}    network UUID
    Return code :
        Success : 200
        Failure : other code

## Tenant operations
This section shows an example for the request of each tenant operation and its possible response.
Following the example, the description of the operation is provided.    

#####  1. Create tenant
Request:
```bash
curl "http://127.0.0.1:9527/nw/v1/tenants/{tenant-name}" -XPOST
```
Response:
```json
{
  "tenant": {
    "name": "string",
    "id": "string",
    "net_number": "string",
    "create_at": "string",
    "quotas": {
      "network": "string"
    },
    "status": "string"
  }
}
```
    Description : create tenant with the default quota
    Method      : POST
    Path        : nw/v1/tenants/{tenant-name}
    Input       :
        {tenant-name}    tenant name
    Return code :
        Success : 200
        Failure : other code

#####  2. Delete tenant
Request:
```bash
curl "http://127.0.0.1:9527/nw/v1/tenants/{tenant-name}" -XDELETE
```
Response:
```json
{
  "ERROR": "string",
  "message": "string",
  "code": "string"
}
```
    Description : delete tenant and its resources and information
    Method      : delete
    Path        : nw/v1/tenants/{tenant-name}
    Input       :
        {tenant-name}    tenant name
    Return code :
        Success : 200
        Failure : other code

#####  3. Get tenant
Request:
```bash
curl "http://127.0.0.1:9527/nw/v1/tenants/{tenant-name}" -XGET
```
Response:
```json
{
  "tenant": {
    "name": "string",
    "id": "string",
    "net_number": "string",
    "create_at": "string",
    "quotas": {
      "network": "string"
    },
    "status": "string"
  }
}
```
    Description : get information of a tenant
    Method      : GET
    Path        : nw/v1/tenants/{tenant-name}
    Input       :
        {tenant-name}    tenant name
    Return code :
        Success : 200
        Failure : other code
