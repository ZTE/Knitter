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
	"github.com/ZTE/Knitter/knitter-agent/domain/object/knitter-agent-obj"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/ZTE/Knitter/pkg/trans-dsl"
)

type ConnectToBrdepAction struct {
}

func (this *ConnectToBrdepAction) Exec(transInfo *transdsl.TransInfo) (err error) {
	klog.Infof("***ConnectToBrdepAction:Exec begin***")
	defer func() {
		if p := recover(); p != nil {
			RecoverErr(p, &err, "ConnectToBrdepAction")
		}
		AppendActionName(&err, "ConnectToBrdepAction")
	}()

	knitterAgtObj := knitteragtobj.GetKnitterAgtObjSingleton()
	err = knitterAgtObj.DeployPodRole.ConnectToBrdep(transInfo.AppInfo.(*KnitterInfo).KnitterObj.CniParam.ContainerID)
	if err != nil {
		klog.Errorf("ConnectToBrdepAction:Exec:knitterAgtObj.DeployPodRole.ConnectToBrdep err: %v", err)
	}
	klog.Infof("***ConnectToBrdepAction:Exec end***")
	return err
}

func (this *ConnectToBrdepAction) RollBack(transInfo *transdsl.TransInfo) {
	// done
	klog.Infof("***ConnectToBrdepAction:RollBack begin***")
	klog.Infof("***ConnectToBrdepAction:RollBack end***")
}
