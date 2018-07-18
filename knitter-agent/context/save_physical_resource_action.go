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
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/ZTE/Knitter/pkg/trans-dsl"
)

type SavePhysicalResourceAction struct {
}

func (this *SavePhysicalResourceAction) Exec(transInfo *transdsl.TransInfo) (err error) {
	klog.Info("***SavePhysicalResourceAction:Exec begin***")
	knitterInfo := transInfo.AppInfo.(*KnitterInfo)
	southIf := knitterInfo.southIfs[transInfo.RepeatIdx]

	err = southIf.vnicRole.SaveResourceToLocalDB()
	if err != nil {
		klog.Errorf("SavePhysicalResourceAction.Exec: SaveResourceToLocalDB FAIL, error: %v", err)
		return err
	}

	phsNouthMgr := &physicalresourceobj.NouthInterfaceObj{
		ContainerID: knitterInfo.KnitterObj.CniParam.ContainerID,
		Driver:      constvalue.VnicType,
		InterfaceID: southIf.port.PortID}
	err = phsNouthMgr.SaveInterface()
	if err != nil {
		klog.Errorf("SavePhysicalResourceAction.Exec: phsNouthMgr.SaveInterface FAIL, error: %v", err)
		delErr := southIf.vnicRole.DeleteResourceFromLocalDB()
		if delErr != nil {
			klog.Errorf("SavePhysicalResourceAction.Exec: DeleteResourceFromLocalDB FAIL, error: %v", delErr)
		}
		return err
	}

	klog.Info("***SavePhysicalResourceAction:Exec end***")
	return nil
}

func (this *SavePhysicalResourceAction) RollBack(transInfo *transdsl.TransInfo) {
	klog.Info("***SavePhysicalResourceAction:RollBack begin***")
	knitterInfo := transInfo.AppInfo.(*KnitterInfo)
	southIf := knitterInfo.southIfs[transInfo.RepeatIdx]

	phsNorthMgr := &physicalresourceobj.NouthInterfaceObj{
		ContainerID: knitterInfo.KnitterObj.CniParam.ContainerID,
		Driver:      southIf.vnicRole.NicType,
		InterfaceID: southIf.port.PortID}
	err := phsNorthMgr.DeleteInterface()
	if err != nil {
		klog.Infof("SavePhysicalResourceAction:RollBack: DeleteInterface[phsNorthMgr: %+v] FAIL, error: %v",
			phsNorthMgr, err)
	}

	err = southIf.vnicRole.DeleteResourceFromLocalDB()
	if err != nil {
		klog.Infof("SavePhysicalResourceAction:RollBack: DeleteResourceFromLocalDB[vnicRole: %+v] FAIL, error: %v",
			southIf.vnicRole, err)
	}
	klog.Info("***SavePhysicalResourceAction:RollBack end***")
}
