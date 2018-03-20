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

package portobj

import (
	"github.com/ZTE/Knitter/knitter-agent/domain/cni"
	"github.com/ZTE/Knitter/knitter-agent/domain/monitor"
	"github.com/ZTE/Knitter/knitter-agent/domain/role/port-role"
	"github.com/ZTE/Knitter/knitter-agent/infra"
	"github.com/ZTE/Knitter/pkg/iaas-accessor"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/rackspace/gophercloud/openstack/networking/v2/ports"
)

type PortEagerAttr struct {
	NetworkName  string
	NetworkPlane string
	PortName     string
	VnicType     string
	Accelerate   string
	PodName      string
	PodNs        string
	FixIP        string
	IPGroupName  string
	Metadata     interface{}
}

type PortLazyAttr struct {
	//NetworkID      string
	VlanID     string
	Vni        int
	ID         string
	Name       string
	TenantID   string
	MacAddress string
	FixedIps   []ports.IP
	GatewayIP  string
	Cidr       string
	NetAttr    NetworkAttrs
	Pflink     *PfLinkResponse
	BusInfos   []string
	BondInfo   cni.BondInfo
	OrgDriver  string
}

type NetworkAttrs struct {
	Name        string                         `json:"name"`
	ID          string                         `json:"network_id"`
	GateWay     string                         `json:"gateway"`
	Cidr        string                         `json:"cidr"`
	CreateTime  string                         `json:"create_time"`
	Status      string                         `json:"state"`
	Public      bool                           `json:"public"`
	Owner       string                         `json:"owner"`
	Description string                         `json:"description"`
	SubnetID    string                         `json:"subnet_id"`
	Provider    iaasaccessor.NetworkExtenAttrs `json:"provider"`
}

type PfLinkResponse struct {
	PfName   string `json:"pf_name"`
	CurrName string `json:"pf_current_name"`
	PciID    string `json:"pci_id"`
	MacAddr  string `json:"mac_addr"`
}

type PortObj struct {
	BuildPortRole portrole.PortBuilderRole

	EagerAttr PortEagerAttr
	LazyAttr  PortLazyAttr
}

type LogicPortObj struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	MacAddress   string `json:"mac_address"`
	IP           string `json:"ip"`
	NetworkID    string `json:"network_id"`
	SubNetID     string `json:"sub_net_id"`
	Cidr         string `json:"cidr"`
	GatewayIP    string `json:"gateway_ip"`
	Accelerate   string `json:"accelerate"`
	NetworkPlane string `json:"network_plane"`
	NetworkName  string `json:"network_name"`
	PodName      string `json:"pod_name"`
	PodNs        string `json:"pod_ns"`
	TenantID     string `json:"tenant_id"`
}

func CreatePortObj(podNs, podName string, port *monitor.Port) (*PortObj, error) {
	portObj := &PortObj{}
	err := portObj.BuildPortRole.Transform(port)
	if err != nil {
		return nil, err
	}
	//metadata
	metadataObj := port.EagerAttr.Metadata
	if metadataObj != nil {
		portObj.EagerAttr.Metadata = metadataObj
	} else {
		portObj.EagerAttr.Metadata = make(map[string]string)
	}
	portObj.EagerAttr.NetworkName = portObj.BuildPortRole.NetworkName
	portObj.EagerAttr.NetworkPlane = portObj.BuildPortRole.NetworkPlane
	portObj.EagerAttr.PortName = portObj.BuildPortRole.PortName
	portObj.EagerAttr.VnicType = portObj.BuildPortRole.VnicType
	portObj.EagerAttr.Accelerate = portObj.BuildPortRole.Accelerate
	portObj.EagerAttr.FixIP = portObj.BuildPortRole.FixIP
	portObj.EagerAttr.IPGroupName = portObj.BuildPortRole.IPGroupName
	portObj.EagerAttr.PodNs = podNs
	portObj.EagerAttr.PodName = podName
	portObj.LazyAttr.ID = port.LazyAttr.ID
	portObj.LazyAttr.Name = port.LazyAttr.Name
	portObj.LazyAttr.TenantID = port.LazyAttr.TenantID
	portObj.LazyAttr.MacAddress = port.LazyAttr.MacAddress
	portObj.LazyAttr.FixedIps = port.LazyAttr.FixedIps
	portObj.LazyAttr.GatewayIP = port.LazyAttr.GatewayIP
	portObj.LazyAttr.Cidr = port.LazyAttr.Cidr

	/*	ports.IP
		ID         string     `json:"id"`
		Name       string     `json:"name"`
		TenantID   string     `json:"tenant_id"`
		MacAddress string     `json:"mac_address"`
		FixedIps   []ports.IP `json:"fixed_ips"`
		GatewayIP  string     `json:"gateway_ip"`
		Cidr       string     `json:"cidr"`*/
	klog.Debugf("CreatePortObj: portObj is [%v], portLazy attr is [%v], porteagerAttr is [%v]", port, port.LazyAttr, port.EagerAttr)
	return portObj, nil
}

func CreateLogicPortsFromTransInfoPortObjs(ports []*PortObj) []*LogicPortObj {
	agt := cni.GetGlobalContext()
	logicPorts := make([]*LogicPortObj, 0)
	for _, port := range ports {
		if port.EagerAttr.Accelerate == "true" &&
			infra.IsCTNetPlane(port.EagerAttr.NetworkPlane) &&
			agt.RunMode != "overlay" {
			continue
		}
		logicPort := LogicPortObj{
			ID:           port.LazyAttr.ID,
			Name:         port.EagerAttr.PortName,
			MacAddress:   port.LazyAttr.MacAddress,
			IP:           port.LazyAttr.FixedIps[0].IPAddress,
			SubNetID:     port.LazyAttr.FixedIps[0].SubnetID,
			Cidr:         port.LazyAttr.Cidr,
			GatewayIP:    port.LazyAttr.GatewayIP,
			Accelerate:   port.EagerAttr.Accelerate,
			NetworkPlane: port.EagerAttr.NetworkPlane,
			NetworkName:  port.EagerAttr.NetworkName,
			PodName:      port.EagerAttr.PodName,
			PodNs:        port.EagerAttr.PodNs,
			TenantID:     port.LazyAttr.TenantID,
			NetworkID:    port.LazyAttr.NetAttr.ID,
		}
		logicPorts = append(logicPorts, &logicPort)
	}
	return logicPorts
}
