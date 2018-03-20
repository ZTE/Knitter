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
	"github.com/ZTE/Knitter/knitter-agent/domain/object/physical-resource-obj"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/ZTE/Knitter/pkg/trans-dsl"
)

type DelPodAction struct {
}

func (this *DelPodAction) Exec(transInfo *transdsl.TransInfo) (err error) {
	klog.Infof("***DelPodAction:Exec begin***")
	defer func() {
		if p := recover(); p != nil {
			RecoverErr(p, &err, "DelPodAction")
		}
		AppendActionName(&err, "DelPodAction")
	}()
	dbObj := dbobj.GetDbObjSingleton()
	knitterAgtInfo := transInfo.AppInfo.(*KnitterInfo)
	knitterObj := knitterAgtInfo.KnitterObj
	if isDetachTransError(knitterAgtInfo) {
		return nil
	}

	dbObj.PodRole.Delete(knitterObj.CniParam)

	(&physicalresourceobj.NouthInterfaceObj{ContainerID: knitterObj.CniParam.ContainerID}).DeleteContainer()
	klog.Infof("***DelPodAction:Exec end***")
	return nil
}

func (this *DelPodAction) RollBack(transInfo *transdsl.TransInfo) {
	klog.Infof("***DelPodAction:RollBack begin***")
	klog.Infof("***DelPodAction:RollBack end***")
}
