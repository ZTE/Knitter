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
	"github.com/ZTE/Knitter/knitter-agent/domain/cni"
	"github.com/ZTE/Knitter/knitter-agent/domain/const-value"
	"github.com/ZTE/Knitter/knitter-agent/domain/object/bridge-obj"
	"github.com/ZTE/Knitter/knitter-agent/domain/object/physical-resource-obj"
	"github.com/ZTE/Knitter/knitter-agent/domain/role/physical-resource-role"
	"github.com/ZTE/Knitter/knitter-agent/domain/role/pod-role"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/ZTE/Knitter/pkg/trans-dsl"
)

type DestroyNeutronPortAction struct {
}

func (this *DestroyNeutronPortAction) Exec(transInfo *transdsl.TransInfo) (err error) {
	klog.Infof("***DestroyNeutronPortAction:Exec begin***")
	defer func() {
		if p := recover(); p != nil {
			RecoverErr(p, &err, "DestroyNeutronPortAction")
		}
		AppendActionName(&err, "DestroyNeutronPortAction")
	}()
	knitterInfo := transInfo.AppInfo.(*KnitterInfo)
	portObj := knitterInfo.portObj

	bridgeObj := bridgeobj.GetBridgeObjSingleton()
	err = bridgeObj.BrintRole.DelPortTable(portObj.LazyAttr.ID)
	if err != nil {
		klog.Errorf("DestroyNeutronPortAction:bridgeObj.BrintRole.DelPortTable portId: %v, err: %v", portObj.LazyAttr.ID, err)
	}

	networkID := portObj.LazyAttr.NetAttr.ID
	err = bridgeObj.BrintRole.DecRefCount(networkID, portObj.EagerAttr.PodNs, portObj.EagerAttr.PodName)
	if err != nil {
		klog.Errorf("DestroyNeutronPortAction:bridgeObj.BrintRole.DecRefCount networkId: %v, err: %v", networkID, err)
	}

	err = podrole.DeleteFromDB(cni.GetGlobalContext().DB, portObj)
	if err != nil {
		klog.Errorf("DestroyNeutronPortAction:Exec:transInfo.podObj.PortRole.DeleteFromDB err: %v", err)
	}

	err = podrole.DeleteFromDB(cni.GetGlobalContext().RemoteDB, portObj)
	if err != nil {
		klog.Errorf("DestroyNeutronPortAction:Exec:transInfo.podObj.PortRole.DeleteFromDB(remote) err: %v", err)
	}

	//delete veth record in local db
	phyNouth := &physicalresourceobj.NouthInterfaceObj{
		ContainerID: knitterInfo.KnitterObj.CniParam.ContainerID,
		Driver:      constvalue.VethType,
		InterfaceID: knitterInfo.vethPair.VethNameOfBridge,
	}
	phyNouth.DeleteInterface()
	phySelf := &physicalresourcerole.VethRole{
		ContainerID:     phyNouth.ContainerID,
		NameByBridge:    phyNouth.InterfaceID,
		NameByContainer: knitterInfo.vethPair.VethNameOfPod,
		MacByContainer:  knitterInfo.vethPair.VethMacOfPod,
		MacByBridge:     knitterInfo.vethPair.VethMacOfBridge,
	}
	phySelf.DeleteResourceFromLocalDB()

	klog.Infof("***DestroyNeutronPortAction:Exec end***")
	return nil
}

func (this *DestroyNeutronPortAction) RollBack(transInfo *transdsl.TransInfo) {
	klog.Infof("***DestroyNeutronPortAction:RollBack begin***")
	klog.Infof("***DestroyNeutronPortAction:RollBack end***")
}
