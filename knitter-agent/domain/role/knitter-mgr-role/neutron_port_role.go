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

package knittermgrrole

import (
	"github.com/ZTE/Knitter/knitter-agent/domain/cni"
	"github.com/ZTE/Knitter/knitter-agent/domain/manager"
	"github.com/ZTE/Knitter/knitter-agent/domain/object/port-obj"
	"github.com/ZTE/Knitter/pkg/inter-cmpt/agt-mgr"
	"github.com/ZTE/Knitter/pkg/klog"
)

type NeutronPortRole struct {
}

func (this NeutronPortRole) Create(tenantID string, podName string, eagerAttr *portobj.PortEagerAttr,
	south bool) (*manager.Port, error) {
	agtCtx := cni.GetGlobalContext()
	var req = manager.CreatePortReq{
		AgtPortReq: agtmgr.AgtPortReq{
			TenantID:    tenantID,
			NetworkName: eagerAttr.NetworkName,
			PortName:    eagerAttr.PortName,
			NodeID:      agtCtx.VMID,
			PodNs:       eagerAttr.PodNs,
			PodName:     podName,
			FixIP:       "",
			IPGroupName: eagerAttr.IPGroupName,
			ClusterID:   cni.GetGlobalContext().ClusterUUID,
		}}
	portByte, err := agtCtx.Mc.CreateNeutronPort(podName, req, tenantID)
	if err != nil {
		klog.Errorf("NeutronPortRole:Create:agtCtx.Mc.CreateNeutronPort for CreatePortReq[%v] failed, error! -%v", req, err)
		return nil, err
	}
	mport := &manager.Port{}
	err = mport.Extract(portByte, eagerAttr.NetworkPlane, tenantID)
	if err != nil {
		klog.Errorf("NeutronPortRole:Create:mport.Extract error! -%v", err)
		return nil, err
	}
	mport.MTU = agtCtx.Mtu
	mport.NetworkName = eagerAttr.NetworkName
	return mport, nil
}

func (this NeutronPortRole) Destroy(mc *manager.ManagerClient, tenantID, portID string) error {
	return mc.DeleteNeutronPort(portID, tenantID)
}
