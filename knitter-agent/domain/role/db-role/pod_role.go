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

package dbrole

import (
	"encoding/json"
	"github.com/ZTE/Knitter/knitter-agent/domain/cni"
	"github.com/ZTE/Knitter/knitter-agent/domain/object/pod-obj"
	"github.com/ZTE/Knitter/knitter-agent/domain/object/port-obj"
	"github.com/ZTE/Knitter/knitter-agent/infra"
	"github.com/ZTE/Knitter/pkg/db-accessor"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/coreos/etcd/client"
)

type PodRole struct {
}

func (this PodRole) GetAllPorts(tenantID, podNs, podName string, db dbaccessor.DbAccessor) ([]*client.Node, error) {
	key := dbaccessor.GetKeyOfInterfaceGroupInPod(tenantID, podNs, podName)
	return db.ReadDir(key)
}

func (this PodRole) GetAllPortJSONList(tenantID, podNs, podName string, db dbaccessor.DbAccessor) ([]string, error) {
	key := dbaccessor.GetKeyOfInterfaceGroupInPod(tenantID, podNs, podName)
	nodes, err := db.ReadDir(key)
	if err != nil {
		klog.Errorf("GetAllPortObjs: db.ReadDir(%s) FAILED, errror: %v", key, err)
		return nil, err
	}

	portJSONList := make([]string, 0)
	for _, node := range nodes {
		portJSON, err := db.ReadLeaf(node.Value)
		if err != nil {
			klog.Errorf("GetAllPortObjs: db.ReadLeaf err: %v", err)
			return nil, err
		}
		portJSONList = append(portJSONList, portJSON)
	}

	klog.Infof("GetAllPortObjs: get all portJsonList: %v SUCC", portJSONList)
	return portJSONList, nil
}

func (this PodRole) IsRestarted(cniParam *cni.CniParam) bool {
	agtCtx := cni.GetGlobalContext()
	keyPortsInPod := dbaccessor.GetKeyOfLogicPortsInPod(cniParam.TenantID,
		cniParam.PodNs, cniParam.PodName)
	ports, err := agtCtx.RemoteDB.ReadDir(keyPortsInPod)
	if err != nil || len(ports) == 0 {
		klog.Infof("IsRestarted: the pod [podns:%v;podname:%v] is first start!",
			cniParam.PodNs, cniParam.PodName)
		return false
	}
	return true
}

func (this PodRole) Insert(podObj *podobj.PodObj) error {
	podObjRepo := podobj.GetPodObjRepoSingleton()
	err := podObjRepo.Add(podObj)
	if err != nil {
		klog.Errorf("PodRole:Insert:podObjRepo.Add err: %v", err)
	}
	return err
}

func (this PodRole) Delete(cniParam *cni.CniParam) {
	podObjRepo := podobj.GetPodObjRepoSingleton()
	podObjRepo.Remove(cniParam)
}

func (this PodRole) IsMigration(cniParam *cni.CniParam) bool {
	agtCtx := cni.GetGlobalContext()
	keyOfVMID := dbaccessor.GetKeyOfVmidForPod(cniParam.TenantID,
		cniParam.PodNs, cniParam.PodName)
	vmIDInETCD, _ := agtCtx.DB.ReadLeaf(keyOfVMID)
	if vmIDInETCD == agtCtx.VMID {
		return false
	}
	klog.Infof("pod in etcd[%v] is migration to another host[%v]", vmIDInETCD, agtCtx.VMID)
	return true
}

func (this PodRole) SaveLogicPortInfoForPod(tanantID, podNs, podName string, ports []*portobj.LogicPortObj) error {
	agtCtx := cni.GetGlobalContext()
	for _, port := range ports {
		if port.Accelerate == "true" &&
			infra.IsCTNetPlane(port.NetworkPlane) {
			continue
		}
		bytePort, _ := json.Marshal(port)
		key := dbaccessor.GetKeyOfLogicPort(tanantID, podNs, podName, port.ID)
		err := agtCtx.RemoteDB.SaveLeaf(key, string(bytePort))
		if err != nil {
			klog.Errorf("SaveLogicPort[%v] SaveLeaf error: %v", port.ID, err)
			return err
		}
	}
	return nil
}

func (this PodRole) DeleteLogicPodDir(tanantID, podNs, podName string) error {
	agtCtx := cni.GetGlobalContext()
	key := dbaccessor.GetKeyOfLogicPod(tanantID, podNs, podName)
	err := agtCtx.RemoteDB.DeleteDir(key)
	if err != nil {
		klog.Errorf("DeleteDir[%v] error: %v", key, err)
		return err
	}
	return nil
}
