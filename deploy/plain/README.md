# How to setup Knitter

## Get all Knitter components
Download tarball from release page for the version you want to try. the download link **here**

- knitter-plugin、knitter-agent、knitter-morinitor and knitter-manager are all of binaries for knitter networking.

```
Note：there is no any command line tool for knitter networking solution now, but we will provide this in future relase.
```

## Install Dependencies
- etcd
specific version:
Download etcd release version from [here](https://github.com/coreos/etcd/releases).

- ovs
specific version:
Download ovs release version from [here](https://github.com/openvswitch/ovs/releases).

## Start knitter-manager
```
knitter-manager -help
 -cfg string
        config file path
  -klog_alsologtostderr
        log to standard error as well as files
  -klog_log_backtrace_at value
        when logging hits line file:N, emit a stack trace
  -klog_log_dir string
        If non-empty, write log files in this directory
  -klog_log_level value
        logs at or above this level will output to log file
  -klog_logtostderr
        log to standard error instead of files
  -klog_stderrthreshold value
        logs at or above this threshold go to stderr
  -version
        print version infomation
```


### Step 1. Create the config file

```
knitter.json:
{
  "conf": {
    "manager": {
      "etcd": {
        "api_version": 3,
        "urls": "http://172.120.0.209:2379"
      },
      "self_service": {
        "url": "http://172.120.0.209:9527/nw/v1"
      },
      "interval": {
        "senconds": "15"
      },
      "net_quota": {
        "no_admin": "10",
        "admin": "100"
      }
    }
  }
}
```
### Step 2. Run knitter manager
Start knitter-manager as root:
```
# knitter-manger -cfg /opt/cni/manager/knitter.json
```

## Start knitter-mornitor
Start knitter-mornitor as root:

## Start knitter-agent
knitter agent component has the same paras with knitter-manger

Start knitter-agent as root:
```
knitter.json:
{
  "conf": {
    "agent": {
      "cluster_type": "k8s",
      "cluster_uuid": "d796cb8a-edff-4066-9467-b8a2cc89517e",
      "etcd": {
        "api_version": 3,
        "url": "http://192.169.1.75:10080",
        "etcd_service_query_url": "http://192.169.1.75:10081/api/microservices/v1/services/etcd/version/v2"
      },
      "event_url": {
        "application": "http://192.169.1.75:10080/harvestor/v1/tenants/admin/events?type=app",
        "wiki": "http://wiki.zte.com.cn/pages/viewpage.action?pageId=17786624"
      },
      "external": {
        "ip": "192.168.1.122"
      },

      "host": {
        "ip": "192.169.1.75",
        "mtu": "1500",
        "type": "virtual_machine",
        "vm_id": "f2a566aa-4bc8-409c-837e-66adc5bc0de8"
      },

      "internal": {
        "ip": "192.168.1.122"
      },

      "k8s": {
        "url": "http://192.169.1.75:8080"
      },

      "manager": {
        "url": "http://192.169.1.75:10080/nwapi/v1"
      },

      "phy": {
        "net_cfg": "/etc/paasnw/paasnw_drivers.conf"
      },

      "region_id": "null",
      "run_mode": {
        "sync": true,
        "type": "underlay"
      }

    },
    "dev": {
      "recover_flag": true
    }
  }
}
```
## Start knitter-plugin
### step 1. install knitter-plugin
knitter-plugin is a CNI plugin, so just put it on default directory that kubernetes can find it by default configuration. Currently, the default directory is `/opt/cni/bin`


### step 2. create cni plugin configuration
default directory is `/etc/cni/net.d`


```
10-knitter.conf:
{
    "name": "knitter",
    "type": "knitter"
}
```