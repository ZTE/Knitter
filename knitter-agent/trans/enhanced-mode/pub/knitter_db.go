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

package modelsext

import (
	"encoding/json"

	"github.com/ZTE/Knitter/pkg/klog"

	"github.com/ZTE/Knitter/pkg/db-accessor"
	"github.com/ZTE/Knitter/pkg/iaas-accessor"

	"fmt"
	"github.com/ZTE/Knitter/knitter-agent/domain/bind"
	"github.com/ZTE/Knitter/knitter-agent/domain/cni"
	"github.com/ZTE/Knitter/knitter-agent/domain/manager"
)

// duplicated code: /home/m11/code/paasnw/src/github.com/ZTE/Knitter/knitter-agent/domain/object/pod-obj/pod_obj_repo.go
func storeSavePod(db dbaccessor.DbAccessor, tenantID string, pod *bind.Pod) error {
	var keyPod = dbaccessor.GetKeyOfPodSelf(tenantID, pod.K8sns, pod.Name)
	klog.Infof("storeSavePod:keyPod:", keyPod)
	bPod, _ := json.Marshal(pod)
	err := db.SaveLeaf(keyPod, string(bPod))
	if err != nil {
		klog.Errorf("storeSavePod:SaveLeaf keyPod error! -%v", err)
		return fmt.Errorf("%v:storeSavePod: SaveLeaf keyPod error", err)
	}
	return nil
}

func storeSavePodVMID(db dbaccessor.DbAccessor, tenantID string, pod *bind.Pod, vmid string) error {
	var keyPodVmid = dbaccessor.GetKeyOfVmidForPod(tenantID, pod.K8sns, pod.Name)
	klog.Infof("storeSavePodVmId:keyPodVmid:", keyPodVmid)
	err := db.SaveLeaf(keyPodVmid, vmid)
	if err != nil {
		klog.Errorf("storePodPort2Etcd: SaveLeaf keyPodVmid error! -%v", err)
		return fmt.Errorf("%v:storePodPort2Etcd: SaveLeaf keyPodVmid error", err)
	}
	return nil
}

func storeSavePodToCluster(agtObj *cni.AgentContext, tenantID string, pod *bind.Pod) error {
	var key = dbaccessor.GetKeyOfPodForNode(agtObj.ClusterID, agtObj.HostIP, pod.K8sns, pod.Name)
	var value = dbaccessor.GetKeyOfPodSelf(tenantID, pod.K8sns, pod.Name)
	klog.Infof("storeSavePodToCluster:key[%v]value[%v]", key, value)
	err := agtObj.DB.SaveLeaf(key, value)
	if err != nil {
		klog.Errorf("storeSavePodToCluster: db.SaveLeaf error! -%v", err)
		return fmt.Errorf("%v:storeSavePodToCluster: db.SaveLeaf error", err)
	}
	return nil
}

func storeSavePodToClusterToEtcd(agtObj *cni.AgentContext, tenantID string, pod *bind.Pod) error {
	var key = dbaccessor.GetKeyOfPodForNode(agtObj.ClusterID, agtObj.HostIP, pod.K8sns, pod.Name)
	var value = dbaccessor.GetKeyOfPodSelf(tenantID, pod.K8sns, pod.Name)
	klog.Infof("storeSavePodToCluster:key[%v]value[%v]", key, value)
	err := agtObj.RemoteDB.SaveLeaf(key, value)
	if err != nil {
		klog.Errorf("storeSavePodToCluster: db.SaveLeaf error! -%v", err)
		return fmt.Errorf("%v:storeSavePodToCluster: db.SaveLeaf error", err)
	}
	return nil
}

func storeSaveInterfaceToEtcd(agtObj *cni.AgentContext, tenantID string, pod *bind.Pod,
	iaasPort *manager.Port, paasPort iaasaccessor.Interface) error {
	bytePort, _ := json.Marshal(paasPort)
	interfaceID := iaasPort.ID + paasPort.PodNs + paasPort.PodName
	keyPort := dbaccessor.GetKeyOfInterfaceSelf(tenantID, interfaceID)
	klog.Infof("storeSaveInterfaceToEtcd:GetKeyOfInterfaceSelf:", keyPort)
	err := agtObj.RemoteDB.SaveLeaf(keyPort, string(bytePort))
	if err != nil {
		klog.Errorf("storeSaveInterfaceToEtcd: SaveLeaf keyPort error! -%v", err)
		return fmt.Errorf("%v:storeSaveInterfaceToEtcd: SaveLeaf keyPort error", err)
	}

	keyPortInPod := dbaccessor.GetKeyOfInterfaceInPod(tenantID,
		iaasPort.ID, pod.K8sns, pod.Name)
	klog.Infof("storeSaveInterfaceToEtcd:GetKeyOfInterfaceInPod:", keyPortInPod)
	err = agtObj.RemoteDB.SaveLeaf(keyPortInPod, keyPort)
	if err != nil {
		klog.Errorf("storeSaveInterfaceToEtcd: SaveLeaf keyPortInPod error! -%v", err)
		return fmt.Errorf("%v:storeSaveInterfaceToEtcd: SaveLeaf keyPortInPod error", err)
	}

	keyPortInNetwork := dbaccessor.GetKeyOfInterfaceInNetwork(tenantID,
		iaasPort.NetworkID, interfaceID)
	klog.Infof("storeSaveInterfaceToEtcd:GetKeyOfInterfaceInNetwork:", keyPortInNetwork)
	err = agtObj.RemoteDB.SaveLeaf(keyPortInNetwork, keyPort)
	if err != nil {
		klog.Errorf("storeSaveInterfaceToEtcd: SaveLeaf keyPortInNetwork error! -%v", err)
		return fmt.Errorf("%v:storeSaveInterfaceToEtcd: SaveLeaf keyPortInNetwork error", err)
	}

	keyPaasPort := dbaccessor.GetKeyOfPaasInterfaceForNode(agtObj.ClusterID, agtObj.HostIP, interfaceID)
	err = agtObj.RemoteDB.SaveLeaf(keyPaasPort, keyPort)
	if err != nil {
		klog.Errorf("storeSaveInterfaceToEtcd: SaveLeaf keyPaasPort error! -%v", err)
		return fmt.Errorf("%v:storeSaveInterfaceToEtcd: SaveLeaf keyPaasPort error", err)
	}

	if paasPort.NetPlane == "eio" {
		eioPortKey := dbaccessor.GetKeyOfIaasEioInterfaceForNode(agtObj.ClusterID, agtObj.HostIP, paasPort.Id)
		err = agtObj.RemoteDB.SaveLeaf(eioPortKey, keyPort)
		if err != nil {
			klog.Errorf("storeSaveInterfaceToEtcd: SaveLeaf keyPaasPort error! -%v", err)
			return fmt.Errorf("%v:storeSaveInterfaceToEtcd: SaveLeaf eioPortKey error", err)
		}
	}

	keyPodSelf := dbaccessor.GetKeyOfPodSelf(tenantID, pod.K8sns, pod.Name)
	klog.Infof("storeSaveInterfaceToEtcd:keyPodSelf:", keyPodSelf)
	keyPodInPort := dbaccessor.GetKeyOfPodInInterface(tenantID, interfaceID)
	klog.Infof("storeSaveInterfaceToEtcd:keyPodInPort:", keyPodInPort)
	err = agtObj.RemoteDB.SaveLeaf(keyPodInPort, keyPodSelf)
	if err != nil {
		klog.Errorf("storeSaveInterfaceToEtcd: SaveLeaf keyPodInPort error! -%v", err)
		return fmt.Errorf("%v:storeSaveInterfaceToEtcd: SaveLeaf keyPodInPort error", err)
	}
	return nil
}

func storeSaveInterface(agtObj *cni.AgentContext, tenantID string, pod *bind.Pod,
	iaasPort *manager.Port, paasPort iaasaccessor.Interface) error {
	bytePort, _ := json.Marshal(paasPort)
	interfaceID := iaasPort.ID + paasPort.PodNs + paasPort.PodName
	keyPort := dbaccessor.GetKeyOfInterfaceSelf(tenantID, interfaceID)
	klog.Infof("storeSaveInterface:GetKeyOfInterfaceSelf:", keyPort)
	err := agtObj.DB.SaveLeaf(keyPort, string(bytePort))
	if err != nil {
		klog.Errorf("storeSaveInterface: SaveLeaf keyPort error! -%v", err)
		return fmt.Errorf("%v:storeSaveInterface: SaveLeaf keyPort error", err)
	}

	keyPortInPod := dbaccessor.GetKeyOfInterfaceInPod(tenantID,
		iaasPort.ID, pod.K8sns, pod.Name)
	klog.Infof("storeSaveInterface:GetKeyOfInterfaceInPod:", keyPortInPod)
	err = agtObj.DB.SaveLeaf(keyPortInPod, keyPort)
	if err != nil {
		klog.Errorf("storeSaveInterface: SaveLeaf keyPortInPod error! -%v", err)
		return fmt.Errorf("%v:storeSaveInterface: SaveLeaf keyPortInPod error", err)
	}

	keyPortInNetwork := dbaccessor.GetKeyOfInterfaceInNetwork(tenantID,
		iaasPort.NetworkID, interfaceID)
	klog.Infof("storeSaveInterface:GetKeyOfInterfaceInNetwork:", keyPortInNetwork)
	err = agtObj.DB.SaveLeaf(keyPortInNetwork, keyPort)
	if err != nil {
		klog.Errorf("storeSaveInterface: SaveLeaf keyPortInNetwork error! -%v", err)
		return fmt.Errorf("%v:storeSaveInterface: SaveLeaf keyPortInNetwork error", err)
	}

	keyPaasPort := dbaccessor.GetKeyOfPaasInterfaceForNode(agtObj.ClusterID, agtObj.HostIP, interfaceID)
	err = agtObj.DB.SaveLeaf(keyPaasPort, keyPort)
	if err != nil {
		klog.Errorf("storeSaveInterface: SaveLeaf keyPaasPort error! -%v", err)
		return fmt.Errorf("%v:storeSaveInterface: SaveLeaf keyPaasPort error", err)
	}

	if paasPort.NetPlane == "eio" {
		eioPortKey := dbaccessor.GetKeyOfIaasEioInterfaceForNode(agtObj.ClusterID, agtObj.HostIP, paasPort.Id)
		err = agtObj.DB.SaveLeaf(eioPortKey, keyPort)
		if err != nil {
			klog.Errorf("storeSaveInterface: SaveLeaf keyPaasPort error! -%v", err)
			return fmt.Errorf("%v:storeSaveInterface: SaveLeaf eioPortKey error", err)
		}
	}

	keyPodSelf := dbaccessor.GetKeyOfPodSelf(tenantID, pod.K8sns, pod.Name)
	klog.Infof("storeSaveInterface:keyPodSelf:", keyPodSelf)
	keyPodInPort := dbaccessor.GetKeyOfPodInInterface(tenantID, interfaceID)
	klog.Infof("storeSaveInterface:keyPodInPort:", keyPodInPort)
	err = agtObj.DB.SaveLeaf(keyPodInPort, keyPodSelf)
	if err != nil {
		klog.Errorf("storeSaveInterface: SaveLeaf keyPodInPort error! -%v", err)
		return fmt.Errorf("%v:storeSaveInterface: SaveLeaf keyPodInPort error", err)
	}
	return nil
}

func DeletePodFromLocalDB(agtObj *cni.AgentContext, cniObj *cni.CniParam) error {
	keyPod := dbaccessor.GetKeyOfPod(cniObj.TenantID, cniObj.PodNs, cniObj.PodName)
	err := agtObj.DB.DeleteDir(keyPod)
	if err != nil {
		klog.Errorf("DeletePodFromLocalDB: DB.DeleteDir(podKey) error! -%v", err)
	}
	keyCluster := dbaccessor.GetKeyOfPodForNode(agtObj.ClusterID, agtObj.HostIP, cniObj.PodNs, cniObj.PodName)
	err = agtObj.DB.DeleteLeaf(keyCluster)
	if err != nil {
		klog.Errorf("DeletePodFromLocalDB: DB.DeleteLeaf(clusterKey) error! -%v", err)
	}
	return nil
}

func DeletePodFromEtcd(agtObj *cni.AgentContext, cniObj *cni.CniParam) error {
	keyPod := dbaccessor.GetKeyOfPod(cniObj.TenantID, cniObj.PodNs, cniObj.PodName)
	err := agtObj.RemoteDB.DeleteDir(keyPod)
	if err != nil {
		klog.Errorf("DeletePodFromEtcd: etcd.DeleteDir(podKey) error! -%v", err)
	}
	keyCluster := dbaccessor.GetKeyOfPodForNode(agtObj.ClusterID, agtObj.HostIP, cniObj.PodNs, cniObj.PodName)
	err = agtObj.RemoteDB.DeleteLeaf(keyCluster)
	if err != nil {
		klog.Errorf("DeletePodFromEtcd: etcd.DeleteLeaf(clusterKey) error! -%v", err)
	}
	return nil
}

var DeletePortFromEtcd = func(agtObj *cni.AgentContext, port iaasaccessor.Interface) error {
	interfaceID := port.Id + port.PodNs + port.PodName
	keyOfInterfaceInNetwork := dbaccessor.GetKeyOfInterfaceInNetwork(port.TenantID, port.NetworkId, interfaceID)
	err3 := agtObj.RemoteDB.DeleteLeaf(keyOfInterfaceInNetwork)
	if err3 != nil {
		klog.Errorf("deletePortFromEtcd: agtObj.RemoteDB.DeleteLeaf(urlInterfaceInNetwork) error! -%v", err3)
	}
	keyOfInterface := dbaccessor.GetKeyOfInterface(port.TenantID, interfaceID)
	err5 := agtObj.RemoteDB.DeleteDir(keyOfInterface)
	if err5 != nil {
		klog.Errorf("deletePortFromEtcd:  agtObj.RemoteDB.DeleteDir(portSelf) error! -%v", err5)
	}
	urlPaasInterfaceForNode := dbaccessor.GetKeyOfPaasInterfaceForNode(agtObj.ClusterID, agtObj.HostIP, interfaceID)
	err6 := agtObj.RemoteDB.DeleteLeaf(urlPaasInterfaceForNode)
	if err6 != nil {
		klog.Errorf("deletePortFromEtcd: agtObj.RemoteDB.DeleteLeaf(urlPaasInterfaceForNode) error! -%v", err6)
	}

	if port.NetPlane == "eio" {
		urlIaasEioInterfaceForNode := dbaccessor.GetKeyOfIaasEioInterfaceForNode(agtObj.ClusterID, agtObj.HostIP, port.Id)
		err7 := agtObj.RemoteDB.DeleteLeaf(urlIaasEioInterfaceForNode)
		if err7 != nil {
			klog.Errorf("deletePortFromEtcd: agtObj.RemoteDB.DeleteLeaf(urlIaasEioInterfaceForNode) error! -%v", err7)
		}
	}
	keyPortInPod := dbaccessor.GetKeyOfInterfaceInPod(port.TenantID,
		port.Id, port.PodNs, port.PodName)
	klog.Infof("deletePortFromEtcd:GetKeyOfInterfaceInPod:", keyPortInPod)
	err8 := agtObj.RemoteDB.DeleteLeaf(keyPortInPod)
	if err8 != nil {
		klog.Errorf("deletePortFromEtcd: agtObj.RemoteDB.DeleteLeaf keyPortInPod error! -%v", err8)
	}
	return nil
}

var DeletePortFromLocalDB = func(agtObj *cni.AgentContext, port iaasaccessor.Interface) error {
	interfaceID := port.Id + port.PodNs + port.PodName
	keyOfInterfaceInNetwork := dbaccessor.GetKeyOfInterfaceInNetwork(port.TenantID, port.NetworkId, interfaceID)
	err3 := agtObj.DB.DeleteLeaf(keyOfInterfaceInNetwork)
	if err3 != nil {
		klog.Errorf("DeletePortFromLocalDB: agtObj.DB.DeleteLeaf(urlInterfaceInNetwork) error! -%v", err3)
	}
	keyOfInterface := dbaccessor.GetKeyOfInterface(port.TenantID, interfaceID)
	err5 := agtObj.DB.DeleteDir(keyOfInterface)
	if err5 != nil {
		klog.Errorf("DeletePortFromLocalDB:  agtObj.DB.DeleteDir(portSelf) error! -%v", err5)
	}
	urlPaasInterfaceForNode := dbaccessor.GetKeyOfPaasInterfaceForNode(agtObj.ClusterID, agtObj.HostIP, interfaceID)
	err6 := agtObj.DB.DeleteLeaf(urlPaasInterfaceForNode)
	if err6 != nil {
		klog.Errorf("DeletePortFromLocalDB: agtObj.DB.DeleteLeaf(urlPaasInterfaceForNode) error! -%v", err6)
	}

	if port.NetPlane == "eio" {
		urlIaasEioInterfaceForNode := dbaccessor.GetKeyOfIaasEioInterfaceForNode(agtObj.ClusterID, agtObj.HostIP, port.Id)
		err7 := agtObj.DB.DeleteLeaf(urlIaasEioInterfaceForNode)
		if err7 != nil {
			klog.Errorf("DeletePortFromLocalDB: agtObj.DB.DeleteLeaf(urlIaasEioInterfaceForNode) error! -%v", err7)
		}
	}
	keyPortInPod := dbaccessor.GetKeyOfInterfaceInPod(port.TenantID,
		port.Id, port.PodNs, port.PodName)
	klog.Infof("DeletePortFromLocalDB:GetKeyOfInterfaceInPod:", keyPortInPod)
	err8 := agtObj.DB.DeleteLeaf(keyPortInPod)
	if err8 != nil {
		klog.Errorf("DeletePortFromLocalDB: DeleteLeaf keyPortInPod error! -%v", err8)
	}
	return nil
}
