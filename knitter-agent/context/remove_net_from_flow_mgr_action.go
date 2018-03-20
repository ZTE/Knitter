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

type RemoveNetFromFlowMgrAction struct {
}

func (this *RemoveNetFromFlowMgrAction) Exec(transInfo *transdsl.TransInfo) (err error) {
	klog.Infof("***RemoveNetFromFlowMgrAction:Exec begin***")
	defer func() {
		if p := recover(); p != nil {
			RecoverErr(p, &err, "RemoveNetFromFlowMgrAction")
		}
		AppendActionName(&err, "RemoveNetFromFlowMgrAction")
	}()
	knitterInfo := transInfo.AppInfo.(*KnitterInfo)
	bridgeObj := bridgeobj.GetBridgeObjSingleton()
	networkID := knitterInfo.portObj.LazyAttr.NetAttr.ID
	err = bridgeObj.BrtunRole.RemoveNetwork(networkID)
	if err != nil {
		klog.Errorf("RemoveNetFromFlowMgrAction:bridgeObj.BrtunRole.RemoveNetwork err: %v", err)
	}
	value, err := bridgeObj.BrintRole.GetTenantNetworkTable(networkID)
	if err != nil {
		klog.Errorf("RemoveNetFromFlowMgrAction:bridgeObj.BrintRole.GetTenantNetworkTable err: %v", err)
	} else {
		knitterAgtObj := knitteragtobj.GetKnitterAgtObjSingleton()
		knitterAgtObj.VlanIDAllocatorRole.Free(value.VlanID)
	}

	err = bridgeObj.BrintRole.DelTenantNetworkTable(networkID)
	if err != nil {
		klog.Errorf("RemoveNetFromFlowMgrAction:bridgeObj.BrintRole.DeleteTenantNetworkTable err: %v", err)
	}

	knitterInfo.Chan <- 1
	knitterInfo.ChanFlag = false

	klog.Infof("***RemoveNetFromFlowMgrAction:Exec end***")
	return nil
}

func (this *RemoveNetFromFlowMgrAction) RollBack(transInfo *transdsl.TransInfo) {
	klog.Infof("***RemoveNetFromFlowMgrAction:RollBack begin***")
	klog.Infof("***RemoveNetFromFlowMgrAction:RollBack end***")
}
