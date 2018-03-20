/*
Copyright 2018 ZTE Corporation. All rights reserved.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package errobj

import (
	"encoding/json"
	"errors"
	"strings"
)

func IsEqual(errLeft, errRight error) bool {
	return errLeft.Error() == errRight.Error()
}

var (
	ErrNetExist              = errors.New("tenant network exist")
	ErrRecordNtExist         = errors.New("record is not existed")
	ErrTooManyRecordsExist   = errors.New("too many records existed")
	ErrTblNtExist            = errors.New("table is not existed")
	ErrTransEnd              = errors.New("trans end")
	ErrNetFieldNtFound       = errors.New("networks field not found")
	ErrJasonNewObjectFailed  = errors.New("jason new object failded")
	ErrJasonGetStringFailed  = errors.New("jason get string failded")
	ErrGetVxlanIDFailed      = errors.New("get vxlan id failded")
	ErrMapNtFound            = errors.New("map not found")
	ErrInvalidStateCode      = errors.New("invalide state code")
	ErrAttachVethToPodFailed = errors.New("attach veth to pod failed")
	ErrBuildNormalNicFailed  = errors.New("build normal nic failed")
	ErrIncRefcountFailed     = errors.New("inc refcount failed")
	ErrContinue              = errors.New("continue")
	ErrHTTPPostFailed        = errors.New("http post failed")
)

/* ERROR CODE for AttachPortsToPod */
var (
	ErrUnmarshalFailed           = errors.New("json unmarshal failed")
	ErrAnalyzeCniFailed          = errors.New("analyze cni param failed")
	ErrNopciFailed               = errors.New("no pci to allocate failed")
	ErrGetpodFailed              = errors.New("get podinfo failed")
	ErrAnalyzePodFailed          = errors.New("analyze pod json failed")
	ErrAddPort2podFailed         = errors.New("add port to pod failed")
	ErrStorePod2etcdFailed       = errors.New("store pod to etcd failed")
	ErrStorePod2DBFailed         = errors.New("store pod to db failed")
	ErrC0PodAlreadyExist         = errors.New("c0 pod already exist")
	ErrInvalidC0ImageName        = errors.New("invalid c0 image name")
	ErrGetC0ContainerIDFailed    = errors.New("c0 container id not found in etcd")
	ErrUsingC0ShareOnNondpdkNode = errors.New("using c0 share on non-dpdk node")
	ErrGetOseNodeFailed          = errors.New("get ose node failed")
	ErrGetK8sNodeFailed          = errors.New("get k8s node failed")
	ErrInvalidClusterType        = errors.New("invalid cluster type error")
	ErrInvalidHostType           = errors.New("invalid host type")
	ErrNodeDpdkLabelNtFound      = errors.New("node dpdk label not found error")
	ErrInvalidNetworkAttrs       = errors.New("invalid network attrs")
)

/* ERROR CODE for AttachPortsToPod */
var (
	ErrBmGetIaasPortFailed = errors.New("bm get iaas port failed")
)

var (
	ErrAddPortToOvsBriFailed       = errors.New("add port to ovs bri failed")
	ErrPortNtEnteredPromiscuousMod = errors.New("port not entered promiscuous mod")
	ErrPortNtFound                 = errors.New("port not found")
	ErrNetPlaneNotSupport          = errors.New("netplane not support")
)

var (
	ErrWaitOvsUsableFailed = errors.New("wait ovs usable failed")
)

var (
	ErrArgTypeMismatch = errors.New("argument type mismatch")
)

var (
	ErrCheckPhysnet = errors.New("check physnet failed")
)

var (
	ErrFixIpsIsNil             = errors.New("fix ips is nil ")
	ErrTenantsIDOrPodNameIsNil = errors.New("tenantId or podName is nil")
	ErrTenantsIDIsNil          = errors.New("tenantId is nil")
)

/* ERROR CODE for portRecycle */
var (
	ErrDbKeyNotFound                = errors.New("key not found")
	ErrDbConnRefused                = errors.New("connect refused by remote")
	ErrNoPodCurrentNodeDbReliable   = errors.New("no running pod in node which db service reliable")
	ErrNoPodCurrentNodeDbNtReliable = errors.New("no running pod in node which db service not reliable")
	ErrK8sGetPodByNodeID            = errors.New("get pod by nodeId failed")
	ErrK8sPodInfoAnalyze            = errors.New("analyze pod info failed")
	ErrOseGetPodByNodeID            = errors.New("get pod of ose by nodeId failed")
	ErrOsePodInfoAnalyze            = errors.New("analyze pod of ose info failed")
	ErrNwClusterType                = errors.New("cluster type error in knitter.json")
	ErrNwPortInfo                   = errors.New("get port info failed")
	ErrNwRecyclePort                = errors.New("recycle iaas port failed")
)

var (
	ErrAny = errors.New("any error")
)

func GetErrMsg(respData []byte) string {
	var respMap map[string]string

	if respData == nil {
		return ""
	}
	err := json.Unmarshal(respData, &respMap)
	if err != nil {
		return ""
	}
	return respMap["message"]
}

var (
	EtcdKeyNotFound    = "Key not found"
	LevelDBKeyNotFound = "leveldb: not found"
)

func IsKeyNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	if strings.Contains(errStr, EtcdKeyNotFound) || strings.Contains(errStr, LevelDBKeyNotFound) {
		return true
	}
	return false
}
