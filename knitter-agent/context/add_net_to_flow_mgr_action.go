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
	"github.com/ZTE/Knitter/knitter-agent/domain/object/bridge-obj"
	"github.com/ZTE/Knitter/knitter-agent/domain/object/knitter-agent-obj"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/ZTE/Knitter/pkg/trans-dsl"
)

type AddNetToFlowMgrAction struct {
}

func (this *AddNetToFlowMgrAction) Exec(transInfo *transdsl.TransInfo) (err error) {
	klog.Infof("***AddNetToFlowMgrAction:Exec begin***")
	defer func() {
		if p := recover(); p != nil {
			RecoverErr(p, &err, "AddNetToFlowMgrAction")
		}
		AppendActionName(&err, "AddNetToFlowMgrAction")
	}()
	portObj := transInfo.AppInfo.(*KnitterInfo).podObj.PortObjs[transInfo.RepeatIdx]
	knitterAgtObj := knitteragtobj.GetKnitterAgtObjSingleton()
	vlanID := knitterAgtObj.VlanIDAllocatorRole.Alloc()
	bridgeObj := bridgeobj.GetBridgeObjSingleton()
	err = bridgeObj.BrintRole.InsertTenantNetworkTable(portObj.LazyAttr.NetAttr.ID,
		portObj.LazyAttr.Vni, vlanID)
	if err != nil {
		klog.Errorf("AddNetToFlowMgrAction:Exec:bridgeObj.BrintRole.InsertTenantNetworkTable err: %v", err)
	}

	err = bridgeObj.BrtunRole.AddNetwork(portObj.LazyAttr.NetAttr.ID, portObj.LazyAttr.Vni, vlanID)
	if err != nil {
		klog.Errorf("AddNetToFlowMgrAction:Exec bridgeObj.BrtunRole.AddNetwork error: %v", err)
		errFree := knitterAgtObj.VlanIDAllocatorRole.Free(vlanID)
		if errFree != nil {
			klog.Errorf("AddNetToFlowMgrAction:Exec:knitterAgtObj.VlanIdAllocatorRole.Free err: %v", errFree)
		}
		return err
	}
	portObj.LazyAttr.VlanID = vlanID
	klog.Infof("***AddNetToFlowMgrAction:Exec end***")
	return nil
}

func (this *AddNetToFlowMgrAction) RollBack(transInfo *transdsl.TransInfo) {
	klog.Infof("***AddNetToFlowMgrAction:RollBack begin***")
	bridgeObj := bridgeobj.GetBridgeObjSingleton()
	portObj := transInfo.AppInfo.(*KnitterInfo).podObj.PortObjs[transInfo.RepeatIdx]
	err := bridgeObj.BrtunRole.RemoveNetwork(portObj.LazyAttr.NetAttr.ID)
	if err != nil {
		klog.Errorf("AddNetToFlowMgrAction:RollBack bridgeObj.BrtunRole.RemoveNetwork error: %v", err)
	}

	err = bridgeObj.BrintRole.DelTenantNetworkTable(portObj.LazyAttr.NetAttr.ID)
	if err != nil {
		klog.Errorf("AddNetToFlowMgrAction:RollBack:bridgeObj.BrintRole.DeleteTenantNetworkTable err: %v", err)
	}
	knitterAgtObj := knitteragtobj.GetKnitterAgtObjSingleton()
	err = knitterAgtObj.VlanIDAllocatorRole.Free(portObj.LazyAttr.VlanID)
	if err != nil {
		klog.Errorf("AddNetToFlowMgrAction:RollBack:knitterAgtObj.VlanIdAllocatorRole.Free err: %v", err)
	}
	klog.Infof("***AddNetToFlowMgrAction:RollBack end***")
}
