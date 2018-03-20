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
	"github.com/ZTE/Knitter/knitter-agent/domain/object/bridge-obj"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/ZTE/Knitter/pkg/trans-dsl"
)

type AttachPortToBrIntAction struct {
}

func (this *AttachPortToBrIntAction) Exec(transInfo *transdsl.TransInfo) (err error) {
	klog.Infof("***AttachPortToBrIntAction:Exec begin***")
	defer func() {
		if p := recover(); p != nil {
			RecoverErr(p, &err, "AttachPortToBrIntAction")
		}
		AppendActionName(&err, "AttachPortToBrIntAction")
	}()
	bridgeObj := bridgeobj.GetBridgeObjSingleton()
	knitterInfo := transInfo.AppInfo.(*KnitterInfo)
	err = bridgeObj.BrintRole.InsertPortTable(knitterInfo.mgrPort.ID, knitterInfo.vethPair.VethNameOfBridge)
	if err != nil {
		klog.Errorf("AttachPortToBrIntAction:Exec bridgeObj.BrintRole.InsertPortTable err: %v", err)
	}

	portObj := knitterInfo.podObj.PortObjs[transInfo.RepeatIdx]
	err = bridgeObj.BrintRole.AttachPort(knitterInfo.vethPair.VethNameOfBridge, portObj.LazyAttr.VlanID)
	if err != nil {
		klog.Errorf("AttachPortToBrIntAction:Exec:bridgeObj.BrintRole.AttachPort err: %v", err)
		errDel := bridgeObj.BrintRole.DelPortTable(knitterInfo.mgrPort.ID)
		if errDel != nil {
			klog.Errorf("AttachPortToBrIntAction:Exec brideObj.BrintRole.DelPortTable err: %v", errDel)
		}
		return err
	}
	knitterInfo.ovsBr = constvalue.OvsBrint
	klog.Infof("***AttachPortToBrIntAction:Exec end***")
	return nil
}

func (this *AttachPortToBrIntAction) RollBack(transInfo *transdsl.TransInfo) {
	klog.Infof("***AttachPortToBrIntAction:RollBack begin***")
	bridgeObj := bridgeobj.GetBridgeObjSingleton()
	knitterInfo := transInfo.AppInfo.(*KnitterInfo)
	err := bridgeObj.BrintRole.DetachPort(knitterInfo.vethPair.VethNameOfBridge)
	if err != nil {
		klog.Errorf("AttachPortToBrIntAction:RollBack:bridgeObj.BrintRole.DetachPort err: %v", err)
	}

	err = bridgeObj.BrintRole.DelPortTable(knitterInfo.mgrPort.ID)
	if err != nil {
		klog.Errorf("AttachPortToBrIntAction:RollBack brideObj.BrintRole.DelPortTable err: %v", err)
	}
	klog.Infof("***AttachPortToBrIntAction:RollBack end***")
}
