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

type IsNewPod struct {
}

func (this *IsNewPod) Ok(transInfo *transdsl.TransInfo) bool {
	knitterInfo := transInfo.AppInfo.(*KnitterInfo)
	dbObj := dbobj.GetDbObjSingleton()
	if dbObj.PodRole.IsRestarted(knitterInfo.KnitterObj.CniParam) {
		klog.Infof("***IsNewPod: false***")
		return false
	}
	klog.Infof("***IsNewPod: true***")
	return true
}

func (this *IsNewPod) RollBack(transInfo *transdsl.TransInfo) {
	// done
	klog.Infof("***IsNewPod:RollBack begin***")
	klog.Infof("***IsNewPod:RollBack end***")
}
