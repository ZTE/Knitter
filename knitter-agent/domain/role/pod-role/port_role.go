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

package podrole

import (
	"encoding/json"
	"fmt"
	"github.com/ZTE/Knitter/knitter-agent/domain/bind"
	"github.com/ZTE/Knitter/knitter-agent/domain/cni"
	"github.com/ZTE/Knitter/knitter-agent/domain/manager"
	"github.com/ZTE/Knitter/knitter-agent/domain/object/port-obj"
	"github.com/ZTE/Knitter/knitter-agent/err-obj"
	"github.com/ZTE/Knitter/pkg/db-accessor"
	"github.com/ZTE/Knitter/pkg/iaas-accessor"
	"github.com/ZTE/Knitter/pkg/klog"
)

type PortRole struct {
	podDataRole *PodDataRole
}

func (this *PortRole) Init(podDataRole *PodDataRole) {
	this.podDataRole = podDataRole
}

func (this *PortRole) Attach(cniObj *cni.CniParam, port *manager.Port, vethNameOfPod string) (*bind.Dpdknic, error) {
	err := bind.AttachVethToPod(cniObj.Netns, port, vethNameOfPod)
	klog.Infof("PortRole:AttachPort:move std's port to pod")
	if err != nil {
		klog.Errorf("PortRole:AttachPort:bind.AttachVethToPod error: %v", err)
		return nil, errobj.ErrAttachVethToPodFailed
	}
	nic, err := bind.BuildNormalNic(port, "")
	if err != nil {
		klog.Errorf("PortRole:AttachPort:adapter.BuildNormalNic error: %v", err)
		return nil, errobj.ErrBuildNormalNicFailed
	}

	return nic, nil
}

func (this *PortRole) StoreToDB(db dbaccessor.DbAccessor, mport *manager.Port,
	portObj *portobj.PortObj, businfo string) error {
	pod := bind.NewPod(this.podDataRole.PodName, this.podDataRole.PodID,
		this.podDataRole.PodNs, this.podDataRole.PodType)
	etcdPort := this.makePaasPort(mport, portObj, businfo)
	err := storeSaveInterface(db, this.podDataRole.PodNs, pod, mport, etcdPort)
	if err != nil {
		return err
	}
	return nil
}

func (this *PortRole) makePaasPort(iaasPort *manager.Port, portObj *portobj.PortObj, businfo string) iaasaccessor.Interface {
	var porttype = ""
	if portObj.EagerAttr.NetworkPlane == "eio" {
		porttype = "dpdk"
	} else {
		if this.podDataRole.PodType == "it" {
			porttype = "nodpdk"
		} else if this.podDataRole.PodType == "ct" {
			if iaasPort.Name == "eth0" {
				porttype = "nodpdk"
			} else {
				porttype = "dpdk"
			}
		} else {
			porttype = "nodpdk"
		}
	}

	vmID := cni.GetGlobalContext().VMID

	paasPort := iaasaccessor.Interface{
		Name:         iaasPort.Name,
		Status:       "ready",
		Id:           iaasPort.ID,
		Ip:           iaasPort.FixedIPs[0].Address,
		MacAddress:   iaasPort.MACAddress,
		NetworkId:    iaasPort.NetworkID,
		SubnetId:     iaasPort.FixedIPs[0].SubnetID,
		DeviceId:     this.podDataRole.PodID,
		VmId:         vmID,
		OwnerType:    "pod",
		PortType:     porttype,
		BusInfo:      businfo,
		NetPlane:     portObj.EagerAttr.NetworkPlane,
		NetPlaneName: portObj.EagerAttr.NetworkName,
		TenantID:     iaasPort.TenantID,
		NicType:      portObj.EagerAttr.VnicType,
		PodName:      this.podDataRole.PodName,
		PodNs:        this.podDataRole.PodNs,
		Accelerate:   "false",
	}
	return paasPort
}

func storeSaveInterface(db dbaccessor.DbAccessor, tenantID string, pod *bind.Pod,
	iaasPort *manager.Port, paasPort iaasaccessor.Interface) error {
	bytePort, _ := json.Marshal(paasPort)
	interfaceID := iaasPort.ID + paasPort.PodNs + paasPort.PodName
	keyPortSelf := dbaccessor.GetKeyOfInterfaceSelf(tenantID, interfaceID)
	klog.Infof("storeSaveInterface:keyPort:", keyPortSelf)
	err := db.SaveLeaf(keyPortSelf, string(bytePort))
	if err != nil {
		klog.Errorf("storeSaveInterface: SaveLeaf keyPort error! -%v", err)
		return fmt.Errorf("%v:storeSaveInterface: SaveLeaf keyPort error", err)
	}

	keyPortInPod := dbaccessor.GetKeyOfInterfaceInPod(tenantID,
		iaasPort.ID, pod.K8sns, pod.Name)
	klog.Infof("storeSaveInterface:keyPortInPod:", keyPortInPod)
	err = db.SaveLeaf(keyPortInPod, keyPortSelf)
	if err != nil {
		klog.Errorf("storeSaveInterface: SaveLeaf keyPortInPod error! -%v", err)
		return fmt.Errorf("%v:storeSaveInterface: SaveLeaf keyPortInPod error", err)
	}

	keyPortInNetwork := dbaccessor.GetKeyOfInterfaceInNetwork(tenantID,
		iaasPort.NetworkID, interfaceID)
	klog.Infof("storeSaveInterface:keyPortInNetwork:", keyPortInNetwork)
	err = db.SaveLeaf(keyPortInNetwork, keyPortSelf)
	if err != nil {
		klog.Errorf("storeSaveInterface: SaveLeaf keyPortInNetwork error! -%v", err)
		return fmt.Errorf("%v:storeSaveInterface: SaveLeaf keyPortInNetwork error", err)
	}

	agtObj := cni.GetGlobalContext()
	keyPaasPort := dbaccessor.GetKeyOfPaasInterfaceForNode(agtObj.ClusterID, agtObj.HostIP, interfaceID)
	err = db.SaveLeaf(keyPaasPort, keyPortSelf)
	if err != nil {
		klog.Errorf("storeSaveInterface: SaveLeaf keyPaasPort error! -%v", err)
		return fmt.Errorf("%v:storeSaveInterface: SaveLeaf keyPaasPort error", err)
	}

	keyPodSelf := dbaccessor.GetKeyOfPodSelf(tenantID, pod.K8sns, pod.Name)
	klog.Infof("storeSaveInterface:keyPodSelf:", keyPodSelf)
	keyPodInPort := dbaccessor.GetKeyOfPodInInterface(tenantID, interfaceID)
	klog.Infof("storeSaveInterface:keyPodInPort:", keyPodInPort)
	err = db.SaveLeaf(keyPodInPort, keyPodSelf)
	if err != nil {
		klog.Errorf("storeSaveInterface: SaveLeaf keyPodInPort error! -%v", err)
		return fmt.Errorf("%v:storeSaveInterface: SaveLeaf keyPodInPort error", err)
	}

	return nil
}

var DeleteFromDB = func(db dbaccessor.DbAccessor, portObj *portobj.PortObj) error {
	interfaceID := portObj.LazyAttr.ID + portObj.EagerAttr.PodNs + portObj.EagerAttr.PodName
	keyPortInNetwork := dbaccessor.GetKeyOfInterfaceInNetwork(portObj.LazyAttr.TenantID,
		portObj.LazyAttr.NetAttr.ID, interfaceID)
	err := db.DeleteLeaf(keyPortInNetwork)
	if err != nil {
		klog.Errorf("PortRole.DeleteFromDB: etcd.DeleteLeaf error! -%v", err)
	}
	urlinterfacesport := dbaccessor.GetKeyOfInterface(portObj.LazyAttr.TenantID, interfaceID)
	err = db.DeleteDir(urlinterfacesport)
	if err != nil {
		klog.Errorf("PortRole.DeleteFromDB: etcd.DeleteDir error! -%v", err)
	}

	agtObj := cni.GetGlobalContext()
	urlPaasInterfaceForNode := dbaccessor.GetKeyOfPaasInterfaceForNode(agtObj.ClusterID, agtObj.HostIP, interfaceID)
	err = agtObj.DB.DeleteLeaf(urlPaasInterfaceForNode)
	if err != nil {
		klog.Errorf("PortRole.DeleteFromDB: agtObj.DB.DeleteLeaf(urlPaasInterfaceForNode) error! -%v", err)
	}

	keyPortInPod := dbaccessor.GetKeyOfInterfaceInPod(portObj.LazyAttr.TenantID,
		portObj.LazyAttr.ID, portObj.EagerAttr.PodNs, portObj.EagerAttr.PodName)
	klog.Infof("PortRole:keyPortInPod:", keyPortInPod)
	err = agtObj.DB.DeleteLeaf(keyPortInPod)
	if err != nil {
		klog.Errorf("PortRole: DeleteLeaf keyPortInPod error! -%v", err)
	}
	return nil
}
