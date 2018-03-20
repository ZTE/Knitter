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
	"github.com/ZTE/Knitter/knitter-agent/domain/object/knitter-agent-obj"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/ZTE/Knitter/pkg/trans-dsl"
)

type SavePodAction struct {
}

func (this *SavePodAction) Exec(transInfo *transdsl.TransInfo) (err error) {
	klog.Infof("***SavePodAction:Exec begin***")
	defer func() {
		if p := recover(); p != nil {
			RecoverErr(p, &err, "SavePodAction")
		}
		AppendActionName(&err, "SavePodAction")
	}()

	dbObj := dbobj.GetDbObjSingleton()
	err = dbObj.PodRole.Insert(transInfo.AppInfo.(*KnitterInfo).podObj)
	if err != nil {
		klog.Errorf("SavePodAction:Exec:dbObj.PodRole.Insert err: %v", err)
	}

	knitterAgtObj := knitteragtobj.GetKnitterAgtObjSingleton()
	err = knitterAgtObj.NetConfigFileRole.Create(transInfo.AppInfo.(*KnitterInfo).podObj,
		transInfo.AppInfo.(*KnitterInfo).Nics)
	if err != nil {
		klog.Errorf("SavePodAction:Exec:knitterAgtObj.CreateNetConfigFile failed err: %v", err)
	}

	klog.Infof("***SavePodAction:Exec end***")
	return nil
}

func (this *SavePodAction) RollBack(transInfo *transdsl.TransInfo) {
	// done
	klog.Infof("***SavePodAction:RollBack begin***")
	dbObj := dbobj.GetDbObjSingleton()
	dbObj.PodRole.Delete(transInfo.AppInfo.(*KnitterInfo).KnitterObj.CniParam)

	klog.Infof("***SavePodAction:RollBack end***")
}
