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

package adapter

import (
	"encoding/json"
	"fmt"
	"github.com/ZTE/Knitter/knitter-agent/domain/cni"
	"github.com/ZTE/Knitter/knitter-agent/domain/manager"
	"github.com/ZTE/Knitter/knitter-agent/err-obj"
	"github.com/ZTE/Knitter/knitter-agent/infra"
	"github.com/ZTE/Knitter/pkg/iaas-accessor"
	"github.com/ZTE/Knitter/pkg/inter-cmpt/agt-mgr"
	"github.com/ZTE/Knitter/pkg/inter-cmpt/mgr-agt"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/antonholmquist/jason"
	"github.com/coreos/etcd/client"
)

var CreateSouthVnicInterface = func(tenantID, portName, networkName, vnicType, containerID string) (*mgragt.CreatePortInfo, error) {
	defer func() {
		if err := recover(); err != nil {
			klog.Infof("bridge:adapter_func: CreateSouthVnicInterface panic recover!")
		}
	}()
	agtCtx := cni.GetGlobalContext()
	var req = agtmgr.AttachPortReq{
		TenantID:    tenantID,
		NetworkName: networkName,
		PortName:    portName,
		NodeID:      agtCtx.VMID,
		VnicType:    vnicType,
		FixIP:       "",
		ClusterID:   agtCtx.ClusterID,
	}
	reqJSON, err := json.Marshal(&req)
	postURL := agtCtx.Mc.GetVnicInterfaceCreateURL(tenantID) + manager.MakeURLReqIDSuffix(manager.NewGUID(containerID))
	klog.Infof("CreateSouthVnicInterface:reqJSON:[%v], postUrl:[%v]", string(reqJSON), postURL)
	postStatusCode, portByte, err := agtCtx.Mc.PostBytes(postURL, reqJSON)
	if err != nil {
		klog.Errorf("masterClient.PostBytes error : %v, %v", err, string(portByte))
		return nil, fmt.Errorf("%v:CreateSouthVnicInterface-ERROR", err)
	}
	if postStatusCode != 200 {
		klog.Errorf("masterClient.PostBytes error : the status code is: -%v, portByte:[%v]", postStatusCode, string(portByte))
		return nil, fmt.Errorf("CreateSouthVnicInterface-ERROR, StatusCode:%v, ErrMsg:%v", postStatusCode, errobj.GetErrMsg(portByte))
	}
	var portObj mgragt.CreatePortResp
	err = json.Unmarshal(portByte, &portObj)
	if err != nil {
		klog.Errorf("CreateSouthVnicInterface: Unmarshal[%s] FAIL, error: %v", portByte, err)
		return nil, err
	}
	klog.Infof("CreateSouthVnicInterface: execute SUCC portObj[%+v]", portObj)

	return &(portObj.Port), nil
}

var DestroySouthVnicInterface = func(tenantId, portID, containerID string) error {
	defer func() {
		if err := recover(); err != nil {
			klog.Infof("bridge:adapter_func: DestroySouthVnicInterface panic recover!")
		}
	}()
	agtCtx := cni.GetGlobalContext()
	delURL := agtCtx.Mc.GetVnicInterfaceDeleteURL(tenantId, agtCtx.VMID, portID) + manager.MakeURLReqIDSuffix(manager.NewGUID(containerID))
	_, statusCode, err := agtCtx.Mc.Delete(delURL)
	if err != nil {
		klog.Errorf("masterClient.Delete error : %v", err)
		return fmt.Errorf("%v:DestroySouthVnicInterface-ERROR", err)
	}
	if statusCode < 200 || statusCode >= 300 {
		klog.Errorf("masterClient.Delete error : the status code is: -%v", statusCode)
		return fmt.Errorf("DestroySouthVnicInterface-ERROR, StatusCode:%v", statusCode)
	}
	return nil
}

// br0
var DestroyPort = func(agtObj *cni.AgentContext, port iaasaccessor.Interface) error {
	return agtObj.Mc.DeleteNeutronPort(port.Id, port.TenantID)
}

var ReadDirFromDb = func(url string) ([]*client.Node, error) {
	agtCtx := cni.GetGlobalContext()
	return agtCtx.DB.ReadDir(url)
}

var ReadLeafFromDb = func(url string) (string, error) {
	agtCtx := cni.GetGlobalContext()
	return agtCtx.DB.ReadLeaf(url)
}

var ClearDirFromDb = func(url string) error {
	agtCtx := cni.GetGlobalContext()
	return agtCtx.DB.DeleteDir(url)
}

var ClearLeafFromDb = func(url string) error {
	agtCtx := cni.GetGlobalContext()
	return agtCtx.DB.DeleteLeaf(url)
}

var ClearLeafFromRemoteDB = func(url string) error {
	agtCtx := cni.GetGlobalContext()
	return agtCtx.RemoteDB.DeleteLeaf(url)
}

//k8s
var GetPodsByNodeID = func(nodeId string) ([]*jason.Object, error) {
	agtCtx := cni.GetGlobalContext()
	return agtCtx.K8s.GetPodsOfNode(nodeId)
}

var JasonObjectGetString = func(object *jason.Object, keys ...string) (string, error) {
	return object.GetString(keys...)
}

var JSONMarshal = func(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

var DataPersisterIsExist = func(dp *infra.DataPersister) bool {
	return dp.IsExist()
}

var DataPersisterLoadFromMemFile = func(dp *infra.DataPersister, i interface{}) error {
	return dp.LoadFromMemFile(i)
}

var DataPersisterSaveToMemFile = func(dp *infra.DataPersister, i interface{}) error {
	return dp.SaveToMemFile(i)
}
