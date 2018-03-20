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
	"github.com/ZTE/Knitter/knitter-agent/domain/cni"
	"github.com/ZTE/Knitter/knitter-agent/domain/const-value"
	"github.com/ZTE/Knitter/pkg/db-accessor"
	"github.com/ZTE/Knitter/pkg/klog"
)

type VnicRole struct {
	TenantID     string   `json:"tenant_id"`
	NetworkID    string   `json:"network_id"`
	NetworkPlane string   `json:"network_plane"`
	NetworkName  string   `json:"network_name"`
	PortID       string   `json:"port_id"`
	MacAddress   string   `json:"mac_address"`
	NicType      string   `json:"nic_type"`
	Accelerate   string   `json:"accelerate"`
	BusInfos     []string `json:"bus_infos"`
	MTU          string   `json:"mtu"`
	OrgDriver    string   `json:"org_driver"`
	ContainerID  string   `json:"container_id"`
	PortName     string   `json:"port_name"`
}

func (this *VnicRole) SaveResourceToLocalDB() error {
	agtCtx := cni.GetGlobalContext()
	keyPort := dbaccessor.GetKeyOfInterfaceSelfForRole(constvalue.VnicType, this.PortID)
	value, _ := json.Marshal(this)
	err := agtCtx.DB.SaveLeaf(keyPort, string(value))
	if err != nil {
		klog.Warningf("SaveLeaf[%v] for vnic self err: %v", this.PortID, err)
	}
	return nil
}

func (this *VnicRole) ReadResourceFromLocalDB() error {
	agtCtx := cni.GetGlobalContext()
	keyPort := dbaccessor.GetKeyOfInterfaceSelfForRole(constvalue.VnicType, this.PortID)
	portInfo, err := agtCtx.DB.ReadLeaf(keyPort)
	json.Unmarshal([]byte(portInfo), this)
	if err != nil {
		klog.Errorf("ReadLeaf[%v] for vnic self err: %v", this.PortID, err)
		return err
	}
	return nil
}

func (this *VnicRole) DeleteResourceFromLocalDB() error {
	agtCtx := cni.GetGlobalContext()
	keyContainerPort := dbaccessor.GetKeyOfInterfaceSelfForRole(constvalue.VnicType, this.PortID)
	err := agtCtx.DB.DeleteLeaf(keyContainerPort)
	if err != nil {
		klog.Warningf("DeleteLeaf[%v] for vnic err: %v", this.PortID, err)
	}
	return nil
}
