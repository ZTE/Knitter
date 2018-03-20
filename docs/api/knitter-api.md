
# PaaS NetWork API 手册

[TOC]

## 网络操作说明
-------------------

#####  1. 创建网络
请求示例：
```bash
curl "http://127.0.0.1:9527/nw/v1/tenants/user_m11/networks" -H Content-Type:application/json -X POST -d '{"network": {"name": "network_show","public":false, "gateway": "123.124.125.1", "cidr": "123.124.125.0/24"}}' | python -m json.tool
```
返回示例：
```json
{
    "network": {
        "cidr": "123.124.125.0/24",
        "create_time": "2016-10-03T07:56:02Z",
        "description": "",
        "gateway": "123.124.125.1",
        "name": "network_show",
        "public": false,
        "name": "user_m11",
        "network_id": "98260ff6-4c47-4c2e-86a5-e18d71a09d6e",
        "state": "ACTIVE"
    }
}
```
    说明：在对应用户名下创建网络
    方法：POST
    路径：/nw/v1/tenants/{user}/networks
    输入信息：
        {user}               租户名称
        name                 网络名称(可选)
        public               是否共享（可选，默认不共享）
        gate-way             默认网关(必须)
        cidr                 网段信息(必须)
    返回码：
        成功：200
        失败：其他
#####  2. Show网络
请求示例：
```bash
curl "http://127.0.0.1:9527/nw/v1/tenants/user_m11/networks/440d2691-fbc6-43c1-9a33-3d7c9cf91432" -XGET | python -m json.tool
```
返回示例：
```json
{
    "network": {
        "cidr": "123.124.125.0/24",
        "create_time": "2016-10-03T08:12:13Z",
        "description": "",
        "gateway": "123.124.125.1",
        "name": "network_show",
        "public": false,
        "name": "user_m11",
        "network_id": "440d2691-fbc6-43c1-9a33-3d7c9cf91432",
        "state": "ACTIVE"
    }
}
```
    说明：根据用户提供的网络UUID查询指定网络信息
    方法：GET
    路径：nw/v1/tenants/{user}/networks/{network_uuid}
    输入信息：
        {user}                租户名称
        {network_uuid}        网络UUID
    返回码：
        成功：200
        失败：其他


#####  3. List网络
请求示例：
```bash
curl "http://127.0.0.1:9527/nw/v1/tenants/user_m11/networks" -XGET | python -m json.tool
```
返回示例：
```json
{
    "networks": [
        {
            "cidr": "123.124.125.0/24",
            "create_time": "2016-10-03T07:56:02Z",
            "description": "",
            "gateway": "123.124.125.1",
            "name": "network_show",
            "network_id": "98260ff6-4c47-4c2e-86a5-e18d71a09d6e",
            "state": "ACTIVE"
        },
        {
            "cidr": "123.124.125.0/24",
            "create_time": "2016-10-03T08:12:13Z",
            "description": "",
            "gateway": "123.124.125.1",
            "name": "provider_network_show",
            "network_id": "440d2691-fbc6-43c1-9a33-3d7c9cf91432",
            "state": "ACTIVE"
        }
    ]
}
```
    说明：查询租户可见的所有网络
    方法：GET
    路径：nw/v1/tenants/{user}}/networks
    输入信息：
        {user}                租户名称
    返回码：
        成功：200
        失败：其他

#####  4. List网络通过网络名称做过滤
请求示例：
```bash
curl "http://127.0.0.1:9527/nw/v1/tenants/user_m11/networks?name=network_show" -XGET | python -m json.tool
```
返回示例：
```json
{
    "networks": [
        {
            "cidr": "123.124.125.0/24",
            "create_time": "2016-10-03T07:56:02Z",
            "description": "",
            "gateway": "123.124.125.1",
            "name": "network_show",
            "network_id": "98260ff6-4c47-4c2e-86a5-e18d71a09d6e",
            "state": "ACTIVE"
        }
    ]
}
```
    说明：查询租户可见的所有网络
    方法：GET
    路径：nw/v1/tenants/{user}}/networks
    输入信息：
        {user}              租户名称
        {name}              网络名称
    返回码：
        成功：200
        失败：其他

#####  5. 删除网络
请求示例：
```bash
curl "http://127.0.0.1:9527/nw/v1/tenants/user_m11/networks/98260ff6-4c47-4c2e-86a5-e18d71a09d6e" -XDELETE
```
返回示例：
```json
无返回值
```
    说明：根据用户提供的网络UUID删除指定网络信息
    方法：DELETE
    路径：nw/v1/tenants/{user}/networks/{network_uuid}
    输入信息：
        {user}                租户名称
        {network_uuid}        网络UUID
    返回码：
        成功：200
        失败：其他



## 租户操作说明
-------------------

#####  1. 创建租户
请求示例：
```bash
POST /opapi/nw/v1/tenants/{tenant-id}
curl "http://127.0.0.1:9527/nw/v1/tenants/{tenant-id} -XPOST
```
返回示例：
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
    说明：使用默认配额数据创建PaaS网络租户
    方法：post
    路径：nw/v1/tenants/{tenant-id}
    输入信息：
        {tenant-id}             租户名称
    返回码：
        成功：200
        失败：其他

#####  2. 删除租户
请求示例：
```bash
DELETE /opapi/nw/v1/tenants/{tenant-id}
curl "http://127.0.0.1:9527/nw/v1/tenants/{tenant-id} -XDELETE
```
返回示例：
```json
{
  "ERROR": "string",
  "message": "string",
  "code": "string"
}
```
    说明：删除PaaS网络租户的所有资源和信息
    方法：delete
    路径：nw/v1/tenants/{tenant-id}
    输入信息：
        {tenant-id}             租户名称
    返回码：
        成功：200
        失败：其他




#####  3. 获取指定租户信息
请求示例：
```bash
GET /opapi/nw/v1/tenants/{tenant-id}
curl "http://127.0.0.1:9527/nw/v1/tenants/{tenant-id} -XGET
```
返回示例：
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
    说明：删除PaaS网络租户的所有资源和信息
    方法：GET
    路径：nw/v1/tenants/{tenant-id}
    输入信息：
        {tenant-id}             租户名称
    返回码：
        成功：200
        失败：其他







