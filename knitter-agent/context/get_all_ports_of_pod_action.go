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
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/ZTE/Knitter/pkg/trans-dsl"
)

type GetAllPortsOfPodAction struct {
}

func (this *GetAllPortsOfPodAction) Exec(transInfo *transdsl.TransInfo) (err error) {
	klog.Infof("***GetAllPortsOfPodAction:Exec begin***")
	defer func() {
		if p := recover(); p != nil {
			RecoverErr(p, &err, "GetAllPortsOfPodAction")
		}
		AppendActionName(&err, "GetAllPortsOfPodAction")
	}()
	knitterObj := transInfo.AppInfo.(*KnitterInfo).KnitterObj

	dbObj := dbobj.GetDbObjSingleton()
	ports, err := dbObj.PodRole.GetAllPorts(knitterObj.CniParam.TenantID, knitterObj.CniParam.PodNs,
		knitterObj.CniParam.PodName, knitterObj.CniParam.DB)
	if err != nil {
		klog.Errorf("GetAllPortLinksAction:knitterObj.CniParam.DB.ReadDir err: %v", err)
		return err
	}

	//knitterObj.PodProtectionRole.AddDetachingTag()
	klog.Infof("current pod fist time delete trans!")

	transInfo.AppInfo.(*KnitterInfo).ports = ports
	transInfo.Times = len(ports)
	klog.Infof("***GetAllPortLinksAction:Exec end***")
	return nil
}

func (this *GetAllPortsOfPodAction) RollBack(transInfo *transdsl.TransInfo) {
	klog.Infof("***GetAllPortLinksAction:RollBack begin***")
	klog.Infof("***GetAllPortLinksAction:RollBack end***")
}
