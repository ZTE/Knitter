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
	"github.com/ZTE/Knitter/knitter-agent/domain/object/db-obj"
	"github.com/ZTE/Knitter/knitter-agent/domain/object/knitter-mgr-obj"
	"github.com/ZTE/Knitter/knitter-agent/domain/object/port-obj"
	"github.com/ZTE/Knitter/knitter-agent/trans/enhanced-mode/bridge"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/ZTE/Knitter/pkg/trans-dsl"
	"github.com/rackspace/gophercloud/openstack/networking/v2/ports"
)

type GeneralModeCreateNeutronBulkPortsAction struct {
}

func (c *GeneralModeCreateNeutronBulkPortsAction) Exec(transInfo *transdsl.TransInfo) (err error) {
	klog.Infof("***CreateNeutronBulkPortsAction:Exec begin***")

	defer func() {
		if p := recover(); p != nil {
			RecoverErr(p, &err, "CreateNeutronBulkPortsAction")
		}
		AppendActionName(&err, "CreateNeutronBulkPortsAction")
	}()

	knitterInfo := transInfo.AppInfo.(*KnitterInfo)
	podObj := knitterInfo.podObj
	knitterMgrObj := knittermgrobj.GetKnitterMgrObjSingleton()

	for i, port := range podObj.PortObjs {
		var tenantID = bridge.GetSouthBoundNetTenantID(port.LazyAttr.NetAttr.Public,
			knitterInfo.KnitterObj.CniParam.TenantID)
		mgrPort, err := knitterMgrObj.NeutronPortRole.Create(
			tenantID,
			knitterInfo.podObj.PodName,
			&port.EagerAttr,
			false)
		if err != nil {
			klog.Errorf("CreateNeutronBulkPortsAction:Exec:knitterMgrObj.NeutronPortRole.Create[%v] err: %v", port.EagerAttr.NetworkName, err)
			return err
		}
		knitterInfo.podObj.PortObjs[i].LazyAttr.ID = mgrPort.ID
		knitterInfo.podObj.PortObjs[i].LazyAttr.Name = mgrPort.Name
		knitterInfo.podObj.PortObjs[i].LazyAttr.MacAddress = mgrPort.MACAddress
		knitterInfo.podObj.PortObjs[i].LazyAttr.FixedIps = []ports.IP{{SubnetID: mgrPort.FixedIPs[0].SubnetID, IPAddress: mgrPort.FixedIPs[0].Address}}
		knitterInfo.podObj.PortObjs[i].LazyAttr.TenantID = mgrPort.TenantID
		knitterInfo.podObj.PortObjs[i].LazyAttr.Cidr = mgrPort.CIDR
		knitterInfo.podObj.PortObjs[i].LazyAttr.GatewayIP = mgrPort.GatewayIP
	}

	//Save logic ports info to etcd
	logicPorts := portobj.CreateLogicPortsFromTransInfoPortObjs(knitterInfo.podObj.PortObjs)
	if len(logicPorts) != 0 {
		dbObj := dbobj.GetDbObjSingleton()
		err = dbObj.PodRole.SaveLogicPortInfoForPod(
			knitterInfo.KnitterObj.CniParam.TenantID,
			knitterInfo.KnitterObj.CniParam.PodNs,
			knitterInfo.KnitterObj.CniParam.PodName, logicPorts)
		if err != nil {
			klog.Errorf("SaveLogicPortInfoForPod error: %v", err)
			return err
		}
	}

	klog.Infof("***CreateNeutronBulkPortsAction:Exec end***")
	return nil
}

func (c *GeneralModeCreateNeutronBulkPortsAction) RollBack(transInfo *transdsl.TransInfo) {
	klog.Infof("***CreateNeutronBulkPortsAction:RollBack begin***")
	knitterMgrObj := knittermgrobj.GetKnitterMgrObjSingleton()
	knitterInfo := transInfo.AppInfo.(*KnitterInfo)
	for _, port := range knitterInfo.podObj.PortObjs {
		if port.LazyAttr.ID != "" {
			err := knitterMgrObj.NeutronPortRole.Destroy(
				&knitterInfo.KnitterObj.CniParam.Manager,
				port.LazyAttr.TenantID,
				port.LazyAttr.ID)
			if err != nil {
				klog.Errorf("CreateNeutronBulkPortsAction.RollBack:knitterMgrObj.NeutronPortRole.Destroy[%v] error: %v", port.LazyAttr.Name, err)
			}
		}
	}
	klog.Infof("***CreateNeutronBulkPortsAction:RollBack end***")
}
