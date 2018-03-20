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
	"fmt"
	"github.com/ZTE/Knitter/knitter-agent/domain/const-value"
	"github.com/ZTE/Knitter/knitter-agent/domain/object/knitter-agent-obj"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/ZTE/Knitter/pkg/trans-dsl"
	"runtime/debug"
	"strings"
)

type DetachTriggerAction struct {
}

func (this *DetachTriggerAction) Exec(transInfo *transdsl.TransInfo) (err error) {
	klog.Infof("***DetachTriggerAction:Exec begin***")
	defer func() {
		if p := recover(); p != nil {
			RecoverErr(p, &err, "DetachTriggerAction")
		}
		AppendActionName(&err, "DetachTriggerAction")
	}()

	knitterAgtObj := knitteragtobj.GetKnitterAgtObjSingleton()
	for i := 0; i < constvalue.MaxNwSendDelPodTimes; i++ {
		err = knitterAgtObj.DetachTriggerRole.SendPostReqToSelf(transInfo.AppInfo.(*KnitterInfo).KnitterObj.Args)
		if err == nil {
			klog.Infof("***DetachTriggerAction:Exec end***")
			return nil
		}
	}
	klog.Errorf("DetachTriggerAction:Exec:knitterAgtObj.DetachTriggerRole.SendPostReqToSelf err: %v", err)
	return err

}

func (this *DetachTriggerAction) RollBack(transInfo *transdsl.TransInfo) {
	// done
	klog.Infof("***DetachTriggerAction:RollBack begin***")
	klog.Infof("***DetachTriggerAction:RollBack end***")
}

func RecoverErr(p interface{}, err *error, actionName string) {
	*err = fmt.Errorf("%v:panic:%v", actionName, p)

	klog.Infof("@@@%v panic recover start!@@@", actionName)
	klog.Error("Stack:", string(debug.Stack()))
	klog.Infof("@@@%v panic recover end!@@@", actionName)
}

func AppendActionName(err *error, actionName string) {
	if *err == nil {
		return
	}
	if strings.Contains((*err).Error(), actionName) {
		return
	}
	*err = fmt.Errorf("%v:%v", actionName, *err)
}
