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
	"github.com/ZTE/Knitter/knitter-agent/domain/object/physical-resource-obj"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/ZTE/Knitter/pkg/trans-dsl"
)

type CleanPhysicalResourceRecordAction struct {
}

func (this *CleanPhysicalResourceRecordAction) Exec(transInfo *transdsl.TransInfo) (err error) {
	klog.Info("***CleanPhysicalResourceRecordAction:Exec begin***")
	defer func() {
		if p := recover(); p != nil {
			RecoverErr(p, &err, "CleanPhysicalResourceRecordAction")
		}
		AppendActionName(&err, "CleanPhysicalResourceRecordAction")
	}()
	knitterAgtInfo := transInfo.AppInfo.(*KnitterInfo)
	containerID := knitterAgtInfo.KnitterObj.CniParam.ContainerID
	PhyNouth := &physicalresourceobj.NouthInterfaceObj{
		ContainerID: containerID,
	}
	drivers, _ := PhyNouth.ReadDriversFromContainer()
	if len(drivers) == 0 {
		klog.Warningf("Read drivers from container[%v] is 0", PhyNouth.ContainerID)
	}
	for _, driver := range drivers {
		PhyNouth.Driver = driver
		//inters, _ := PhyNouth.ReadInterfacesFromDriver()
		//for _, inter := range inters {
		//	PhyNouth.DeleteInterface()
		//	physicalresourceobj.DeleteInterfaceInfoRecord(driver, inter)
		//}
		PhyNouth.CleanDriver()
	}
	PhyNouth.CleanContainer()
	klog.Info("***CleanPhysicalResourceRecordAction:Exec end***")
	return nil
}

func (this *CleanPhysicalResourceRecordAction) RollBack(transInfo *transdsl.TransInfo) {
	klog.Infof("***CleanPhysicalResourceRecordAction:RollBack begin***")
	klog.Infof("***CleanPhysicalResourceRecordAction:RollBack end***")
}
