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
	"github.com/ZTE/Knitter/knitter-agent/domain/manager"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/ZTE/Knitter/pkg/trans-dsl"
)

type GeneralModeGetMgrPortAction struct {
}

func (this *GeneralModeGetMgrPortAction) Exec(transInfo *transdsl.TransInfo) (err error) {
	klog.Infof("***GeneralModeGetMgrPortAction:Exec begin***")
	defer func() {
		if p := recover(); p != nil {
			RecoverErr(p, &err, "GeneralModeGetMgrPortAction")
		}
		AppendActionName(&err, "GeneralModeGetMgrPortAction")
	}()
	agtCtx := cni.GetGlobalContext()

	knitterInfo := transInfo.AppInfo.(*KnitterInfo)
	portObj := knitterInfo.podObj.PortObjs[transInfo.RepeatIdx]

	transInfo.AppInfo.(*KnitterInfo).mgrPort = &manager.Port{
		ID:          portObj.LazyAttr.ID,
		NetworkID:   portObj.LazyAttr.NetAttr.ID,
		NetworkName: portObj.LazyAttr.NetAttr.Name,
		Name:        portObj.LazyAttr.Name,
		MACAddress:  portObj.LazyAttr.MacAddress,
		FixedIPs:    []manager.IP{{SubnetID: portObj.LazyAttr.FixedIps[0].SubnetID, Address: portObj.LazyAttr.FixedIps[0].IPAddress}},
		TenantID:    portObj.LazyAttr.TenantID,
		CIDR:        portObj.LazyAttr.Cidr,
		GatewayIP:   portObj.LazyAttr.GatewayIP,
		MTU:         agtCtx.Mtu,
	}
	klog.Infof("***GeneralModeGetMgrPortAction:Exec end***")
	return nil
}

func (this *GeneralModeGetMgrPortAction) RollBack(transInfo *transdsl.TransInfo) {
	klog.Infof("***GeneralModeGetMgrPortAction:RollBack begin***")
	klog.Infof("***GeneralModeGetMgrPortAction:RollBack end***")
}
