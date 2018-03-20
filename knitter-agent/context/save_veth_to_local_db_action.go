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
	"github.com/ZTE/Knitter/knitter-agent/domain/const-value"
	"github.com/ZTE/Knitter/knitter-agent/domain/object/physical-resource-obj"
	"github.com/ZTE/Knitter/knitter-agent/domain/role/physical-resource-role"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/ZTE/Knitter/pkg/trans-dsl"
)

type SaveVethToLocalDBAction struct {
}

func (this *SaveVethToLocalDBAction) Exec(transInfo *transdsl.TransInfo) (err error) {
	klog.Infof("***SaveVethToLocalDBAction:Exec begin***")
	defer func() {
		if p := recover(); p != nil {
			RecoverErr(p, &err, "SaveVethToLocalDBAction")
		}
		AppendActionName(&err, "SaveVethToLocalDBAction")
	}()
	vethInfo := transInfo.AppInfo.(*KnitterInfo).vethPair
	vethRole := &physicalresourcerole.VethRole{
		ContainerID:     transInfo.AppInfo.(*KnitterInfo).KnitterObj.CniParam.ContainerID,
		BridgeName:      transInfo.AppInfo.(*KnitterInfo).ovsBr,
		NameByBridge:    vethInfo.VethNameOfBridge,
		NameByContainer: vethInfo.VethNameOfPod,
		MacByBridge:     vethInfo.VethMacOfBridge,
		MacByContainer:  vethInfo.VethMacOfPod,
	}
	vethRole.SaveResourceToLocalDB()
	phyNouth := &physicalresourceobj.NouthInterfaceObj{
		ContainerID: transInfo.AppInfo.(*KnitterInfo).KnitterObj.CniParam.ContainerID,
		Driver:      constvalue.VethType,
		InterfaceID: vethRole.NameByBridge,
	}
	phyNouth.SaveInterface()
	klog.Infof("***SaveVethToLocalDBAction:Exec end***")
	return nil
}

func (this *SaveVethToLocalDBAction) RollBack(transInfo *transdsl.TransInfo) {
	klog.Infof("***SavePortAction:RollBack begin***")
	vethInfo := transInfo.AppInfo.(*KnitterInfo).vethPair
	vethRole := &physicalresourcerole.VethRole{
		ContainerID:     transInfo.AppInfo.(*KnitterInfo).KnitterObj.CniParam.ContainerID,
		BridgeName:      transInfo.AppInfo.(*KnitterInfo).ovsBr,
		NameByBridge:    vethInfo.VethNameOfBridge,
		NameByContainer: vethInfo.VethNameOfPod,
		MacByBridge:     vethInfo.VethMacOfBridge,
		MacByContainer:  vethInfo.VethMacOfPod,
	}
	vethRole.DeleteResourceFromLocalDB()
	phyNouth := &physicalresourceobj.NouthInterfaceObj{
		ContainerID: transInfo.AppInfo.(*KnitterInfo).KnitterObj.CniParam.ContainerID,
		Driver:      constvalue.VethType,
		InterfaceID: vethRole.NameByBridge,
	}
	phyNouth.DeleteInterface()
	klog.Infof("***SavePortAction:RollBack end***")
}
