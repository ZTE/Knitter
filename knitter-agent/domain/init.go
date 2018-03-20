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

package domain

import (
	"github.com/ZTE/Knitter/knitter-agent/domain/bind"
	"github.com/ZTE/Knitter/knitter-agent/domain/const-value"
	"github.com/ZTE/Knitter/knitter-agent/domain/object/bridge-obj"
	"github.com/ZTE/Knitter/knitter-agent/domain/object/cluster-mgr-obj"
	"github.com/ZTE/Knitter/knitter-agent/domain/object/db-obj"
	"github.com/ZTE/Knitter/knitter-agent/domain/object/knitter-agent-obj"
	"github.com/ZTE/Knitter/knitter-agent/domain/object/knitter-mgr-obj"
	"github.com/ZTE/Knitter/knitter-agent/domain/object/pod-obj"
	"github.com/ZTE/Knitter/knitter-agent/domain/ovs"
	"github.com/ZTE/Knitter/knitter-agent/domain/role/bridge-role/brint-sub-role"
	"github.com/ZTE/Knitter/knitter-agent/domain/role/knitter-agent-role"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/antonholmquist/jason"
	"strconv"
)

func Init(cfg *jason.Object) error {
	clustermgrobj.GetClusterMgrObjSingleton()
	podobj.GetPodObjRepoSingleton()
	dbobj.GetDbObjSingleton()
	knitteragtobj.GetKnitterAgtObjSingleton()
	knittermgrobj.GetKnitterMgrObjSingleton()
	bridgeObj := bridgeobj.GetBridgeObjSingleton()
	err := ovs.WaitOvsUsable()
	if err != nil {
		klog.Errorf("domain.Init: WaitOvsUsable error: %v", err)
		return err
	}

	if bridgeObj.BrtunRole.SyncSwitch(cfg) {
		klog.Errorf("domain.Init: bridgeObj.BrtunRole.SyncSwitch is true, start sync")
		go bridgeObj.BrtunRole.StartSync()
	}

	bind.DestroyResidualBrintIntfcs()
	tenantNetworkMap := brintsubrole.GetTenantNetworkTableSingleton().GetAll()
	for key, value := range tenantNetworkMap {
		klog.Infof("domain.Init: add network:[id: %s] value:[%v] to flow manager", key, value)
		bridgeObj.BrtunRole.AddNetwork(key, value.Vni, value.VlanID)
	}

	knitteragentrole.InitVlanIDSlice(getVlanIDSlice(tenantNetworkMap))

	go brintsubrole.RetryInitDefaultGw(brintsubrole.GetDefaultNetName(cfg))

	go bind.DelNousedOvsBrInterfacesLoop(constvalue.OvsBrint)

	return nil
}

func getVlanIDSlice(tenantNetworkMap map[string]brintsubrole.TenantNetworkValue) []int {
	buff := make([]int, len(tenantNetworkMap))
	if len(tenantNetworkMap) > 0 {
		index := 0
		for _, value := range tenantNetworkMap {
			id, err := strconv.Atoi(value.VlanID)
			if err != nil {
				klog.Errorf("domain.Init:strconv.Atoi err! value.VlanId: %v, err: %v", value.VlanID, err)
				continue
			}
			buff[index] = id
			index++
		}
	}
	return buff
}
