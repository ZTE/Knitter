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

package iaasaccessor

import (
	"github.com/ZTE/Knitter/pkg/inter-cmpt/mgr-iaas"
	"github.com/rackspace/gophercloud/openstack/networking/v2/subnets"
)

const FLAT_DEFAULT_ID string = "0"

type Interface struct {
	Name         string   `json:"name"`
	Status       string   `json:"status"`
	Id           string   `json:"port_id"`
	Ip           string   `json:"ip"`
	MacAddress   string   `json:"mac_address"`
	NetworkId    string   `json:"network_id"`
	SubnetId     string   `json:"subnet_id"`
	DeviceId     string   `json:"device_id"`
	VmId         string   `json:"vm_id"`
	OwnerType    string   `json:"owner_type"`
	PortType     string   `json:"port_type"`
	BusInfo      string   `json:"bus_info"`
	NetPlane     string   `json:"net_plane_type"`
	NetPlaneName string   `json:"net_plane_name"`
	TenantID     string   `json:"tenant_id"`
	NicType      string   `json:"nic_type"`
	PodName      string   `json:"pod_name"`
	PodNs        string   `json:"pod_ns"`
	Accelerate   string   `json:"accelerate"`
	BusInfos     []string `json:"bus_infos"`
	BondMode     string   `json:"bond_mode"`
	FixIP        string   `json:"ip_addr"`
	OrgDriver    string   `json:"org_driver"`
	IPGroupID    string   `json:"ipgroup_id"`
}

type Network struct {
	Name string `json:"name"`
	Id   string `json:"network_id"`
}

type VlanExtAttr struct {
	NetworkType     string `json:"network_type"`
	SegmentID       string `json:"segment_id"`
	PhysicalNetwork string `json:"physical_network"`
}

type NetworkExtenAttrs struct {
	Id   string `json:"id"`
	Name string `json:"name"`

	// Specifies the nature of the physical network mapped to this network
	// resource. Examples are flat, vlan, vxlan, or gre.
	NetworkType string `json:"provider:network_type"`

	// Identifies the physical network on top of which this network object is
	// being implemented. The OpenStack Networking API does not expose any facility
	// for retrieving the list of available physical networks. As an example, in
	// the Open vSwitch plug-in this is a symbolic name which is then mapped to
	// specific bridges on each compute host through the Open vSwitch plug-in
	// configuration file.
	PhysicalNetwork string `json:"provider:physical_network"`

	// Identifies an isolated segment on the physical network; the nature of the
	// segment depends on the segmentation model defined by network_type. For
	// instance, if network_type is vlan, then this is a vlan identifier;
	// otherwise, if network_type is gre, then this will be a gre key.
	SegmentationID string `json:"provider:segmentation_id"`

	VlanTransparent bool `json:"vlan_transparent"`
}

type Router struct {
	Name     string `json:"name"`
	Id       string `json:"router_id"`
	ExtNetId string `json:"external_id"`
}

// description of what this is.
type Subnet struct {
	Id              string                   `mapstructure:"id" json:"id"`
	NetworkId       string                   `mapstructure:"network_id" json:"network_id"`
	Name            string                   `mapstructure:"name" json:"name"`
	Cidr            string                   `mapstructure:"cidr" json:"cidr"`
	GatewayIp       string                   `mapstructure:"gateway_ip" json:"gateway_ip"`
	TenantId        string                   `mapstructure:"tenant_id" json:"tenant_id"`
	AllocationPools []subnets.AllocationPool `mapstructure:"allocation_pools" json:"allocation_pools"`
}

type IaaS interface {
	GetTenantUUID(cfg string) (string, error)
	GetType() string
	Auth() error

	CreatePort(networkId, subnetId, portName, ip, mac, vnicType string) (*Interface, error)
	CreateBulkPorts(req *mgriaas.MgrBulkPortsReq) ([]*Interface, error)
	GetPort(id string) (*Interface, error)
	DeletePort(id string) error
	ListPorts(networkID string) ([]*Interface, error)

	CreateNetwork(name string) (*Network, error)
	CreateProviderNetwork(name, nwType, phyNet, sId string, vlanTransparent bool) (*Network, error)
	DeleteNetwork(id string) error
	GetNetworkID(networkName string) (string, error)
	GetNetwork(id string) (*Network, error)
	GetNetworkExtenAttrs(id string) (*NetworkExtenAttrs, error)

	CreateSubnet(id, cidr, gw string, alloctionPools []subnets.AllocationPool) (*Subnet, error)
	DeleteSubnet(id string) error
	GetSubnetID(networkId string) (string, error)
	GetSubnet(id string) (*Subnet, error)

	CreateRouter(name, extNetId string) (string, error)
	UpdateRouter(id, name, extNetID string) error
	GetRouter(id string) (*Router, error)
	DeleteRouter(id string) error

	AttachPortToVM(vmId, portId string) (*Interface, error)
	DetachPortFromVM(vmId, portId string) error

	AttachNetToRouter(routerId, subNetId string) (string, error)
	DetachNetFromRouter(routerId, netId string) (string, error)

	GetAttachReq() int
	SetAttachReq(req int)
}
