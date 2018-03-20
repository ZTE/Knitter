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

package context

import (
	"encoding/json"
	"github.com/ZTE/Knitter/knitter-agent/domain/adapter"
	"github.com/ZTE/Knitter/knitter-agent/domain/cni"
	"github.com/ZTE/Knitter/knitter-agent/domain/object/db-obj"
	"github.com/ZTE/Knitter/knitter-agent/domain/object/port-obj"
	"github.com/ZTE/Knitter/pkg/db-accessor"
	"github.com/ZTE/Knitter/pkg/iaas-accessor"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/ZTE/Knitter/pkg/trans-dsl"
)

type DeleteLogicPortsForPodAction struct {
}

func (this *DeleteLogicPortsForPodAction) Exec(transInfo *transdsl.TransInfo) (err error) {
	klog.Infof("***DeleteLogicPortsForPodAction:Exec begin***")
	defer func() {
		if p := recover(); p != nil {
			RecoverErr(p, &err, "DeleteLogicPortsForPodAction")
		}
		AppendActionName(&err, "DeleteLogicPortsForPodAction")
	}()

	knitterInfo := transInfo.AppInfo.(*KnitterInfo)
	cniParam := knitterInfo.KnitterObj.CniParam
	agtCtx := cni.GetGlobalContext()
	key := dbaccessor.GetKeyOfLogicPortsInPod(cniParam.TenantID, cniParam.PodNs, cniParam.PodName)
	nodes, err := agtCtx.RemoteDB.ReadDir(key)
	if err != nil {
		klog.Warningf("DeleteLogicPortForPodAction ReadDir[%v] error: %v", key, err)
		return nil
	}
	klog.Infof("Logic port lenth is %v", len(nodes))
	for i, node := range nodes {
		var port portobj.LogicPortObj
		json.Unmarshal([]byte(node.Value), &port)
		klog.Infof("Logic port[%v] is %v", i, port)
		delPort := makeDelPortWithLogicPort(port)
		err = adapter.DestroyPort(agtCtx, delPort)
		if err != nil {
			klog.Warningf("DestroyPort error: %v", err)
			continue
		}
		key := dbaccessor.GetKeyOfLogicPort(cniParam.TenantID, cniParam.PodNs, cniParam.PodName, port.ID)
		err := agtCtx.RemoteDB.DeleteLeaf(key)
		if err != nil {
			klog.Warningf("DeleteLeaf[%v] error: %v", port.ID, err)
		}
	}
	key = dbaccessor.GetKeyOfLogicPortsInPod(cniParam.TenantID, cniParam.PodNs, cniParam.PodName)
	nodes, err = agtCtx.RemoteDB.ReadDir(key)
	if err != nil {
		klog.Warningf("DeleteLogicPortForPodAction ReadDir[%v] error: %v", key, err)
		return nil
	}
	if len(nodes) == 0 {
		dbobj.GetDbObjSingleton().PodRole.DeleteLogicPodDir(cniParam.TenantID, cniParam.PodNs, cniParam.PodName)
	}

	klog.Infof("***DeleteLogicPortsForPodAction:Exec end***")
	return nil
}

func (this *DeleteLogicPortsForPodAction) RollBack(transInfo *transdsl.TransInfo) {
	klog.Infof("***DeleteLogicPortsForPodAction:RollBack begin***")
	klog.Infof("***DeleteLogicPortsForPodAction:RollBack end***")
}

func makeDelPortWithLogicPort(port portobj.LogicPortObj) iaasaccessor.Interface {
	return iaasaccessor.Interface{
		Id:       port.ID,
		TenantID: port.TenantID,
	}
}
