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
	"github.com/ZTE/Knitter/knitter-agent/domain/ovs"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/ZTE/Knitter/pkg/trans-dsl"
)

type DetachPortFromBrintAction struct {
}

func (this *DetachPortFromBrintAction) Exec(transInfo *transdsl.TransInfo) (err error) {
	klog.Infof("***DetachPortFromBrintAction:Exec begin***")
	defer func() {
		if p := recover(); p != nil {
			RecoverErr(p, &err, "DetachPortFromBrintAction")
		}
		AppendActionName(&err, "DetachPortFromBrintAction")
	}()
	knitterInfo := transInfo.AppInfo.(*KnitterInfo)
	knitterAgtObj := knitteragtobj.GetKnitterAgtObjSingleton()
	portObj, err := knitterAgtObj.PortObjRole.Create(knitterInfo.ports[transInfo.RepeatIdx].Value)
	if err != nil {
		return err
	}
	knitterInfo.portObj = portObj

	bridgeObj := bridgeobj.GetBridgeObjSingleton()
	vethName, err := bridgeObj.BrintRole.GetPortTable(portObj.LazyAttr.ID)
	if err != nil {
		klog.Errorf("DetachPortFromBrintAction: bridgeobj.BrintRole.GetPortTable port[%s] error: %v",
			portObj.LazyAttr.ID, err)
		knitterInfo.vethNameOk = false
	} else {
		knitterInfo.vethNameOk = true
	}

	err = bridgeObj.BrintRole.DetachPort(vethName)
	if err != nil {
		klog.Errorf("DetachPortFromBrintAction:bridgeObj.BrintRole.DetachPort vethName:%s err: %v", vethName, err)
	}
	knitterInfo.vethPair = &ovs.VethPair{}
	knitterInfo.vethPair.VethNameOfBridge = vethName
	klog.Infof("***DetachPortFromBrintAction:Exec end***")
	return nil
}

func (this *DetachPortFromBrintAction) RollBack(transInfo *transdsl.TransInfo) {
	klog.Infof("***DetachPortFromBrintAction:RollBack begin***")
	klog.Infof("***DetachPortFromBrintAction:RollBack end***")
}
