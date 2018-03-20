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

package physicalresourceobj

import (
	"errors"
	"github.com/ZTE/Knitter/knitter-agent/domain/cni"
	"github.com/ZTE/Knitter/knitter-agent/domain/const-value"
	"github.com/ZTE/Knitter/knitter-agent/domain/role/physical-resource-role"
	"github.com/ZTE/Knitter/pkg/db-accessor"
	"github.com/ZTE/Knitter/pkg/klog"
	"strings"
)

type PhysicalResourceObj struct {
	VethRole physicalresourcerole.VethRole
	VnicRole physicalresourcerole.VnicRole
}

type NouthInterfaceObj struct {
	ContainerID string `json:"container_id"`
	Driver      string `json:"driver"`
	InterfaceID string `json:"interface_id"`
}

func (this *NouthInterfaceObj) SaveInterface() error {
	agtCtx := cni.GetGlobalContext()
	keyNouthInterface := dbaccessor.GetKeyOfNouthInterface(this.ContainerID, this.Driver, this.InterfaceID)
	keyInterfaceSelf := dbaccessor.GetKeyOfInterfaceSelfForRole(this.Driver, this.InterfaceID)
	err := agtCtx.DB.SaveLeaf(keyNouthInterface, keyInterfaceSelf)
	if err != nil {
		klog.Errorf("Save[%v] for nouth interface err: %v", this.InterfaceID, err)
		return err
	}
	return nil
}

func (this *NouthInterfaceObj) ReadDriversFromContainer() ([]string, error) {
	agtCtx := cni.GetGlobalContext()
	driverList := make([]string, 0)
	keyNouthContainer := dbaccessor.GetKeyOfNouthContainer(this.ContainerID)
	nodes, err := agtCtx.DB.ReadDir(keyNouthContainer)
	if err != nil {
		klog.Errorf("Read[%v] for driver list err: %v", this.ContainerID, err)
		return driverList, err
	}
	klog.Infof("------------ nodes: %v", nodes)
	for _, node := range nodes {
		driver := strings.TrimPrefix(node.Key, keyNouthContainer+"/")
		driverList = append(driverList, driver)
	}
	return driverList, nil
}

func (this *NouthInterfaceObj) ReadInterfacesFromDriver() ([]string, error) {
	agtCtx := cni.GetGlobalContext()
	interfaceList := make([]string, 0)
	keyNouthInterfaceList := dbaccessor.GetKeyOfNouthInterfaceList(this.ContainerID, this.Driver)
	nodes, err := agtCtx.DB.ReadDir(keyNouthInterfaceList)
	if err != nil {
		klog.Errorf("Read[%v, %v] for interface list err: %v", this.ContainerID, this.Driver, err)
		return interfaceList, err
	}
	for _, node := range nodes {
		interfaceID := strings.TrimPrefix(node.Key, keyNouthInterfaceList+"/")
		interfaceList = append(interfaceList, interfaceID)
	}
	return interfaceList, nil
}

func (this *NouthInterfaceObj) DeleteInterface() error {
	agtCtx := cni.GetGlobalContext()
	keyNouthInterface := dbaccessor.GetKeyOfNouthInterface(this.ContainerID, this.Driver, this.InterfaceID)
	err := agtCtx.DB.DeleteLeaf(keyNouthInterface)
	if err != nil {
		klog.Warningf("Delete[%v] for nouth interface err: %v", this.InterfaceID, err)
	}
	return nil
}

func (this *NouthInterfaceObj) DeleteContainer() error {
	agtCtx := cni.GetGlobalContext()
	keyNouthInterface := dbaccessor.GetKeyOfNouthContainer(this.ContainerID)
	err := agtCtx.DB.DeleteDir(keyNouthInterface)
	if err != nil {
		klog.Warningf("Delete[%v] for nouth container err: %v", this.ContainerID, err)
	}
	return nil
}

func (this *NouthInterfaceObj) CleanDriver() error {
	agtCtx := cni.GetGlobalContext()
	InterList, _ := this.ReadInterfacesFromDriver()
	if len(InterList) == 0 {
		keyNouthDriver := dbaccessor.GetKeyOfNouthInterfaceList(this.ContainerID, this.Driver)
		agtCtx.DB.DeleteDir(keyNouthDriver)
	}
	return nil
}

func (this *NouthInterfaceObj) CleanContainer() error {
	agtCtx := cni.GetGlobalContext()
	DriverList, _ := this.ReadDriversFromContainer()
	if len(DriverList) == 0 {
		keyNouthContainer := dbaccessor.GetKeyOfNouthContainer(this.ContainerID)
		agtCtx.DB.DeleteDir(keyNouthContainer)
	}
	return nil
}

type SouthInterfaceObj struct {
	NetworkID     string   `json:"network_id"`
	ChanType      string   `json:"chan_type"`
	InterfaceID   string   `json:"interface_id"`
	IsPublic      bool     `json:"is_public"`
	VlanID        string   `json:"vlan_id"`
	ContainerList []string `json:"container_list"`
}

func (this *SouthInterfaceObj) SaveInterface() error {
	agtCtx := cni.GetGlobalContext()
	keyNouthInterface := dbaccessor.GetKeyOfSouthInterface(this.NetworkID, this.ChanType, this.InterfaceID)
	var keyInterfaceSelf string
	if this.ChanType == constvalue.Br0Vnic || this.ChanType == constvalue.C0Vnic {
		keyInterfaceSelf = dbaccessor.GetKeyOfInterfaceSelfForRole(constvalue.VnicType, this.InterfaceID)
	} else if this.ChanType == constvalue.C0Vf {
		keyInterfaceSelf = dbaccessor.GetKeyOfInterfaceSelfForRole(constvalue.VfType, this.InterfaceID)
	} else {
		klog.Warningf("South interface type[%v] error", this.ChanType)
		return errors.New("south interface type error")
	}
	err := agtCtx.DB.SaveLeaf(keyNouthInterface, keyInterfaceSelf)
	if err != nil {
		klog.Warningf("Save[%v] for south interface err: %v", this.InterfaceID, err)
	}
	return nil
}

func (this *SouthInterfaceObj) DeleteInterface() error {
	agtCtx := cni.GetGlobalContext()
	keyNouthInterface := dbaccessor.GetKeyOfSouthInterface(this.NetworkID, this.ChanType, this.InterfaceID)
	err := agtCtx.DB.DeleteLeaf(keyNouthInterface)
	if err != nil {
		klog.Warningf("Delete[%v] for south interface err: %v", this.InterfaceID, err)
	}
	return nil
}
