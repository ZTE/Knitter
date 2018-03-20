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
	"github.com/ZTE/Knitter/knitter-agent/domain/object/knitter-mgr-obj"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/ZTE/Knitter/pkg/trans-dsl"
)

type GetVniAction struct {
}

func (this *GetVniAction) Exec(transInfo *transdsl.TransInfo) (err error) {
	klog.Infof("***GetVniAction:Exec begin***")
	defer func() {
		if p := recover(); p != nil {
			RecoverErr(p, &err, "GetVniAction")
		}
		AppendActionName(&err, "GetVniAction")
	}()
	knitterMgrObj := knittermgrobj.GetKnitterMgrObjSingleton()
	portObj := transInfo.AppInfo.(*KnitterInfo).podObj.PortObjs[transInfo.RepeatIdx]
	vni, err := knitterMgrObj.VniRole.Get(portObj.LazyAttr.NetAttr.ID)
	if err != nil {
		klog.Errorf("GetVniAction:knitterMgrObj.VniRole.Handle")
		return err
	}
	portObj.LazyAttr.Vni = vni
	klog.Infof("***GetVniAction:Exec end***")
	return nil
}

func (this *GetVniAction) RollBack(transInfo *transdsl.TransInfo) {
	// done
	klog.Infof("***GetVniAction:RollBack begin***")
	klog.Infof("***GetVniAction:RollBack end***")
}
