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
	"fmt"
	"github.com/ZTE/Knitter/knitter-agent/domain/cni"
	"github.com/ZTE/Knitter/knitter-agent/domain/object/knitter-mgr-obj"
	"github.com/ZTE/Knitter/knitter-agent/domain/object/port-obj"
	"github.com/ZTE/Knitter/knitter-agent/domain/ovs"
	"github.com/ZTE/Knitter/knitter-agent/err-obj"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/ZTE/Knitter/pkg/trans-dsl"
)

type GetNetworkAttrsAction struct {
}

func (this *GetNetworkAttrsAction) Exec(transInfo *transdsl.TransInfo) (err error) {
	klog.Infof("***GetNetworkAttrsAction:Exec begin***")
	defer func() {
		if p := recover(); p != nil {
			RecoverErr(p, &err, "GetNetworkAttrsAction")
		}
		AppendActionName(&err, "GetNetworkAttrsAction")
	}()

	var networkNames []string
	var needProvider = false

	knitterInfo := transInfo.AppInfo.(*KnitterInfo)
	for _, portObj := range knitterInfo.podObj.PortObjs {
		networkNames = append(networkNames, portObj.EagerAttr.NetworkName)
	}

	if knitterInfo.KnitterObj.CniParam.HostType == "virtual_machine" {
		needProvider = false
	} else if knitterInfo.KnitterObj.CniParam.HostType == "bare_metal" {
		needProvider = true
	}
	networkAttrs, err := GetNetworksAttrs(knitterInfo.KnitterObj.CniParam.PodNs, networkNames, needProvider)
	if err != nil {
		klog.Errorf("Get network attrs failed")
		return err
	}
	for i, portObj := range knitterInfo.podObj.PortObjs {
		var flag bool = false
		for _, netAttr := range networkAttrs {
			if portObj.EagerAttr.NetworkName == netAttr.Name {
				knitterInfo.podObj.PortObjs[i].LazyAttr.NetAttr = *netAttr
				knitterInfo.podObj.PortObjs[i].LazyAttr.VlanID = netAttr.Provider.SegmentationID
				flag = true
				break
			}
		}
		if !flag {
			return errobj.ErrInvalidNetworkAttrs
		}
	}
	if needProvider == true {
		err = CheckPortsVnicType(knitterInfo.podObj.PortObjs)
		if err != nil {
			return err
		}
	}
	klog.Infof("***GetNetworkAttrsAction:Exec end***")
	return nil
}

func GetNetworksAttrs(tenantID string, networkNames []string, needProvider bool) ([]*portobj.NetworkAttrs, error) {
	knitterMgrObj := knittermgrobj.GetKnitterMgrObjSingleton()
	networksAttrs, err := knitterMgrObj.NetworkAttrsRole.Get(tenantID, networkNames, needProvider)
	if err != nil {
		klog.Errorf("GetNetworksAttrs:knitterMgrObj.NetworkIdRole.Handle err: %v", err)
		return nil, err
	}

	return networksAttrs, nil
}

func CheckPortsVnicType(portsObj []*portobj.PortObj) error {
	agtCtx := cni.GetGlobalContext()

	for _, portObj := range portsObj {
		VnicType, errSetVnicType := checkVnicType(agtCtx, portObj.LazyAttr.NetAttr.Provider.PhysicalNetwork, portObj.EagerAttr.VnicType)
		if errSetVnicType != nil {
			klog.Errorf("checkVnicType:%v error:%v", portObj.EagerAttr.VnicType,
				errSetVnicType)
			return fmt.Errorf("checkVnicType:%v error:%v", portObj.EagerAttr.VnicType,
				errSetVnicType)
		}
		if portObj.EagerAttr.VnicType != VnicType {
			klog.Infof("Actual VnicType:%v, BP VnicType:%v", VnicType, portObj.EagerAttr.VnicType)
			portObj.EagerAttr.VnicType = VnicType
		}
		klog.Infof("checkVnicType:%v successful", portObj.EagerAttr.VnicType)
	}
	return nil
}

func checkVnicType(agtObj *cni.AgentContext, phyNet, PortType string) (string, error) {
	driver, err := ovs.GetNwMechDriver(phyNet, PortType, agtObj.PaasNwConfPath)
	if err != nil {
		klog.Errorf("ovs.GetNwMechDriver: error! %v", err)
		return "", fmt.Errorf("%v:checkVnicType error! Get driver failed", err)
	}
	if driver == "ovs" {
		return "normal", nil
	} else if driver == "sriov" {
		return "direct", nil
	} else if driver == "physical" {
		return "physical", nil
	}
	klog.Errorf("checkVnicType:paasnw_driver.json do not config ovs and sirov or get phynetwork from iaas error.")
	return "", errors.New("checkVnicType error! please check configuration or iaas interface")
}

func (this *GetNetworkAttrsAction) RollBack(transInfo *transdsl.TransInfo) {
	// todo for bare metal
	klog.Infof("***GetVlanExtAttrAction:RollBack begin***")
	klog.Infof("***GetVlanExtAttrAction:RollBack end***")
}
