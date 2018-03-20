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

package podobj

import (
	"encoding/json"
	"fmt"
	"github.com/ZTE/Knitter/knitter-agent/domain/bind"
	"github.com/ZTE/Knitter/knitter-agent/domain/cni"
	"github.com/ZTE/Knitter/pkg/db-accessor"
	"github.com/ZTE/Knitter/pkg/klog"
)

type PodObjRepo struct {
}

func (this *PodObjRepo) Add(podObj *PodObj) error {
	agtCtx := cni.GetGlobalContext()
	err := storePod(agtCtx, podObj)
	if err != nil {
		return err
	}

	err = storePodVMID(agtCtx, podObj)
	if err != nil {
		return err
	}

	return storePodToCluster(agtCtx, podObj)
}

func (this *PodObjRepo) Update(key string, value *PodObj) {

}

func (this *PodObjRepo) Get(key string) *PodObj {
	return nil
}

func (this *PodObjRepo) Remove(cniParam *cni.CniParam) {
	agtCtx := cni.GetGlobalContext()
	keyPod := dbaccessor.GetKeyOfPod(
		cniParam.TenantID, cniParam.PodNs, cniParam.PodName)
	err := agtCtx.DB.DeleteDir(keyPod)
	if err != nil {
		klog.Errorf("PodObjRepo:Remove agtCtx.DB.DeleteDir err: %v", err)
	}
	err = agtCtx.RemoteDB.DeleteDir(keyPod)
	if err != nil {
		klog.Errorf("PodObjRepo:Remove agtCtx.RemoteDB.DeleteDir err: %v", err)
	}

	keyCluster := dbaccessor.GetKeyOfPodForNode(agtCtx.ClusterID, agtCtx.HostIP,
		cniParam.PodNs, cniParam.PodName)
	err = agtCtx.DB.DeleteLeaf(keyCluster)
	if err != nil {
		klog.Errorf("PodObjRepo:Remove agtCtx.DB.DeleteLeaf err: %v", err)
	}
	err = agtCtx.RemoteDB.DeleteLeaf(keyCluster)
	if err != nil {
		klog.Errorf("PodObjRepo:Remove agtCtx.RemoteDB.DeleteLeaf err: %v", err)
	}
}

var podObjRepo *PodObjRepo

func GetPodObjRepoSingleton() *PodObjRepo {
	if podObjRepo == nil {
		podObjRepo = &PodObjRepo{}
	}

	return podObjRepo
}

func storePod(agtCtx *cni.AgentContext, podObj *PodObj) error {
	var keyPod = dbaccessor.GetKeyOfPodSelf(podObj.TenantID, podObj.PodNs, podObj.PodName)
	klog.Infof("storePod:keyPod:", keyPod)
	pod := bind.NewPod(podObj.PodName, podObj.PodID, podObj.PodNs, podObj.PodType)
	bPod, _ := json.Marshal(pod)
	err := agtCtx.DB.SaveLeaf(keyPod, string(bPod))
	if err != nil {
		klog.Errorf("storePod:DB.SaveLeaf keyPod error! -%v", err)
		return fmt.Errorf("%v:storePod: DB.SaveLeaf keyPod error", err)
	}
	err = agtCtx.RemoteDB.SaveLeaf(keyPod, string(bPod))
	if err != nil {
		klog.Errorf("storePod:RemoteDB.SaveLeaf keyPod error! -%v", err)
		return fmt.Errorf("%v:storePod: RemoteDB.SaveLeaf keyPod error", err)
	}
	return nil
}

func storePodVMID(agtCtx *cni.AgentContext, podObj *PodObj) error {
	var keyPodVmid = dbaccessor.GetKeyOfVmidForPod(podObj.TenantID,
		podObj.PodNs, podObj.PodName)
	klog.Infof("storePodVmId:keyPodVmid:", keyPodVmid)
	err := agtCtx.DB.SaveLeaf(keyPodVmid, podObj.VMID)
	if err != nil {
		klog.Errorf("storePodVmId: DB.SaveLeaf keyPodVmid error! %v", err)
		return fmt.Errorf("%v:storePodVmId: DB.SaveLeaf keyPodVmid error", err)
	}

	err = agtCtx.RemoteDB.SaveLeaf(keyPodVmid, podObj.VMID)
	if err != nil {
		klog.Errorf("storePodVmId: RemoteDB.SaveLeaf keyPodVmid error! %v", err)
		return fmt.Errorf("%v:storePodVmId: RemoteDB.SaveLeaf keyPodVmid error", err)
	}
	return nil
}

func storePodToCluster(agtCtx *cni.AgentContext, podObj *PodObj) error {
	key := dbaccessor.GetKeyOfPodForNode(
		agtCtx.ClusterID, agtCtx.HostIP, podObj.PodNs, podObj.PodName)
	value := dbaccessor.GetKeyOfPodSelf(
		podObj.TenantID, podObj.PodNs, podObj.PodName)
	klog.Infof("storePodToCluster:key[%v]value[%v]", key, value)
	err := agtCtx.DB.SaveLeaf(key, value)
	if err != nil {
		klog.Errorf("storePodToCluster: DB.SaveLeaf error! %v", err)
		return fmt.Errorf("%v:storePodToCluster: DB.SaveLeaf error", err)
	}

	err = agtCtx.RemoteDB.SaveLeaf(key, value)
	if err != nil {
		klog.Errorf("storePodToCluster: RemoteDB.SaveLeaf error! %v", err)
		return fmt.Errorf("%v:storePodToCluster: RemoteDB.SaveLeaf error", err)
	}
	return nil
}
