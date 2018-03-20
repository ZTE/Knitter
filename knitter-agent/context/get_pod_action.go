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
	"errors"
	"github.com/ZTE/Knitter/knitter-agent/domain/object/cluster-mgr-obj"
	"github.com/ZTE/Knitter/knitter-agent/domain/object/knitter-monitor-obj"
	"github.com/ZTE/Knitter/knitter-agent/domain/object/pod-obj"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/ZTE/Knitter/pkg/trans-dsl"
)

type GetPodAction struct {
}

func (this *GetPodAction) Exec(transInfo *transdsl.TransInfo) (err error) {
	klog.Infof("***GetPodAction:Exec begin***")
	defer func() {
		if p := recover(); p != nil {
			RecoverErr(p, &err, "GetPodAction")
		}
	}()

	knitterObj := transInfo.AppInfo.(*KnitterInfo).KnitterObj
	podNs := knitterObj.CniParam.PodNs
	podName := knitterObj.CniParam.PodName
	clusterMgrObj := clustermgrobj.GetClusterMgrObjSingleton()
	mtr := knittermonitorobj.GetKnitterMgrObjSingleton()
	monitorPod, err := mtr.GetPodRole.GetPod(podNs, podName)

	if !transInfo.AppInfo.(*KnitterInfo).IsAttachOrDetachFlag && err != nil {
		return nil
	}
	if err != nil {
		klog.Errorf("mtr.GetPodRole.GetPod(podNs:[%v], podName: [%v])err, error is [%v]", podNs, podName, err)
		return err

	}
	if !transInfo.AppInfo.(*KnitterInfo).IsAttachOrDetachFlag && !monitorPod.IsSuccessful {
		klog.Infof("getPodAction: is Detach")
		return nil
	}

	if !monitorPod.IsSuccessful {
		klog.Errorf("mtr.GetPodRole.GetPod(podNs:[%v], podName: [%v])err, error is [%v]", podNs, podName, monitorPod.ErrorMsg)
		err = errors.New("mtr.GetPodRole.GetPod " + monitorPod.ErrorMsg)
		return err
	}
	//todo to delete but podjson is using
	_, podJSON, err := clusterMgrObj.ClusterMgrRole.GetPod(podNs, podName)
	//if trans is detach, need to judge return code 200
	if !transInfo.AppInfo.(*KnitterInfo).IsAttachOrDetachFlag && err != nil {
		klog.Warning("getPodAction: is detach")
		return nil
	}
	if err != nil {
		klog.Errorf("GetPod info error: %v", err)
		return errors.New("GetPodAction:GetPod error")
	}

	podObj, err := podobj.CreatePodObj(knitterObj.CniParam, monitorPod)
	if err != nil {
		return err
	}

	transInfo.AppInfo.(*KnitterInfo).podObj = podObj
	klog.Infof("GetPodAction: podObj is [%+v]", podObj)
	klog.Infof("GetPodAction: podObj is [%+v]", podObj.PortObjs)
	transInfo.AppInfo.(*KnitterInfo).podJSON = podJSON
	transInfo.Times = len(podObj.PortObjs)
	klog.Infof("***GetPodAction:Exec end***")
	return nil
}

func (this *GetPodAction) RollBack(transInfo *transdsl.TransInfo) {
	// done
	klog.Infof("***GetPodAction:RollBack begin***")
	klog.Infof("***GetPodAction:RollBack end***")
}
