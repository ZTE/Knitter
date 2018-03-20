## Knitter configuration files
[toc]

Knitter includes four components: `knitter-manager`, `knitter-monitor`, `knitter-agent` and `knitter`.

### 1. knitter-manager
`knitter-manager` is running on Kubernetes master.

#### 1.1 File directory structure
```
--conf
  --app.conf
--knitter-manager
--knitter.json
```

#### 1.2 knitter.json
All of the files named `knitter.json` are the configuration of Knitter components themselves. The file below is for `knitter-manager`.

N.B. the inline comments of json the json files in this documentation are just for clarifying. You should remove them if you want to use those configuration. 
```json
{
  "conf": {
    "manager": {
      "etcd": {
        "api_version": 3,						// API version of etcd
        "urls": "http://172.120.0.209:2379"	    // etcd address
      },
      "self_service": {
        "url": "http://172.120.0.209:9527/nw/v1" // serving url of knitter-manager
      },
      "interval": {
        "senconds": "15"    					// interval of hearbeat period
      },
      "net_quota": {
        "no_admin": "10",						// networks quota of common users
        "admin": "100"							// networks quota of admin
      }
    }
  }
}
```

#### 1.3 app.conf
conf/app.conf is the configuration file of [beego](https://github.com/astaxie/beego) framework.
```
appname = knitter_manager // component name
httpport = 9527			  // listening port of the component
runmode = prod			  // running as production mode
autorender = false		  // don't auto render, used by beego
copyrequestbody = true	  // copy requestbody, used by beego
EnableDocs = true		  // enable docs, used by beego
```

### 2. knitter-monitor
`knitter-monitor` is running on Kubernetes master.

#### 2.1 File directory structure
The same as `knitter-manager`.

#### 2.2 knitter.json
```json
{
  "conf": {
    "monitor": {
      "log_dir": "/root/info/logs/nwnode",			// path of logging
      "manager": {
        "url": "http://172.120.0.209:9527/api/v1"	// serving url of knitter-manager
      },
      "etcd": {
        "api_version": 3,						   // API version of etcd
        "urls": "http://172.120.0.209:2379"		   // etcd address
      }
    }
  }
}
```

#### 2.3 app.conf
It's similar to the `app.conf` configuration file of knitter-manager.
```
appname = knitter-monitor
httpport = 6001
runmode = prod
autorender = false
copyrequestbody = true
EnableDocs = true
```

### 3. knitter-agent
`knitter-agent` is running on Kubernetes nodes.

#### 3.1 File directory structure
The same as `knitter-manager`.

#### 3.2 knitter.json
```json
{
  "conf": {
    "agent": {
      "cluster_type": "k8s",                                        // cluster type
      "cluster_uuid": "2ccd499e-42b5-4075-92df-8bbc0ed52bbf",		// cluster uuid
      "etcd": {
        "api_version": 3,											// API version of etcd
        "urls": "http://172.120.0.209:2379"							// etcd address
      },
      "external": {
        "ip": "172.120.0.210"										// source address when accessing external network
      },
      "host": {
        "ip": "172.120.0.210",										// host information
        "mtu": "1400",
        "type": "virtual_machine",
        "vm_id": "f0c79d22-a79b-4be4-b369-bd93d3e82dec"
      },
      "internal": {
        "ip": "172.120.0.210"									    // vxlan channel IP
      },
      "k8s": {
        "url": "http://172.120.0.209:8080"						    // k8s api server
      },
      "manager": {
        "url": "http://172.120.0.209:9527/api/v1"				    // knitter-manager service
      },
      "monitor": {
        "url": "http://172.120.0.209:6001/api/v1"				    // knitter-monitor service
      },
      "run_mode": {
        "sync": true,
        "type": "overlay"									       // running mode
      }
    },
    "dev": {
      "recover_flag": true                                         // recovering flag, used by dev mode.
    }
  }
}
```

#### 3.3 app.conf
It's similar to the same configuration file of knitter-manager.
```
appname = knitter_agent
httpport = 6006
runmode = prod
autorender = false
copyrequestbody = true
EnableDocs = true
```

### 4. knitter
`knitter` is the CNI plugin which stays on Kubernetes nodes. Its configuration file is `10-knitter.conf` whose location is `/etc/cni/net.d`.

This file is just a placeholder for CNI. It's content is as below. Knitter doesn't consume it actually.
```
{
    "name": "knitter",
    "type": "knitter"
}
```
