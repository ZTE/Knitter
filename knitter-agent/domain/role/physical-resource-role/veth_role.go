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

package physicalresourcerole

import (
	"encoding/json"
	"errors"
	"github.com/ZTE/Knitter/knitter-agent/domain/cni"
	"github.com/ZTE/Knitter/knitter-agent/domain/const-value"
	"github.com/ZTE/Knitter/knitter-agent/domain/ovs"
	"github.com/ZTE/Knitter/pkg/db-accessor"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/vishvananda/netlink"
)

type VethRole struct {
	ContainerID     string `json:"container_id"`
	NameByContainer string `json:"name_by_container"`
	NameByBridge    string `json:"name_by_bridge"`
	MacByContainer  string `json:"mac_by_container"`
	MacByBridge     string `json:"mac_by_bridge"`
	BridgeName      string `json:"bridge_name"`
}

func (this *VethRole) SaveResourceToLocalDB() error {
	agtCtx := cni.GetGlobalContext()
	keyPort := dbaccessor.GetKeyOfInterfaceSelfForRole(constvalue.VethType, this.NameByBridge)
	value, _ := json.Marshal(this)
	err := agtCtx.DB.SaveLeaf(keyPort, string(value))
	if err != nil {
		klog.Warningf("SaveLeaf[%v] for veth self err: %v", this.NameByBridge, err)
	}
	return nil
}

func (this *VethRole) ReadResourceFromLocalDB() error {
	agtCtx := cni.GetGlobalContext()
	keyPort := dbaccessor.GetKeyOfInterfaceSelfForRole(constvalue.VethType, this.NameByBridge)
	portInfo, err := agtCtx.DB.ReadLeaf(keyPort)
	json.Unmarshal([]byte(portInfo), this)
	if err != nil {
		klog.Errorf("ReadLeaf[%v] for veth self err: %v", this.NameByBridge, err)
		return err
	}
	return nil
}

func (this *VethRole) DeleteResourceFromLocalDB() error {
	agtCtx := cni.GetGlobalContext()
	keyContainerPort := dbaccessor.GetKeyOfInterfaceSelfForRole(constvalue.VethType, this.NameByBridge)
	err := agtCtx.DB.DeleteLeaf(keyContainerPort)
	if err != nil {
		klog.Warningf("DeleteLeaf[%v] for veth err: %v", this.NameByBridge, err)
	}
	return nil
}

func (this *VethRole) Delete() error {
	link, err := ovs.GetLinkByName(this.NameByBridge)
	if err == nil {
		err := netlink.LinkDel(link)
		if err != nil {
			klog.Errorf("netlink.LinkDel(%v) error: %v", link, err)
			return err
		}
		return nil
	}
	klog.Errorf("createVethPair:netlink.LinkByName(%s) error: %v", this.NameByBridge, err)
	return errors.New("vethName not found")
	//err := bridge.DestroyVethPair(this.NameByBridge)
	//if err != nil {
	//	klog.Errorf("DestroyVethPair[%v] error! -%v", this.NameByBridge, err)
	//}
	//return nil
}
