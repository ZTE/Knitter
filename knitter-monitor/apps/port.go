package apps

import (
	"github.com/rackspace/gophercloud/openstack/networking/v2/ports"

	"github.com/ZTE/Knitter/knitter-monitor/services"
)

type PortForAgent struct {
	LazyAttr  PortLazyAttrForAgent  `json:"lazy_attr"`
	EagerAttr PortEagerAttrForAgent `json:"eager_attr"`
}

type PortEagerAttrForAgent struct {
	NetworkName  string      `json:"network_name"`
	NetworkPlane string      `json:"network_plane"`
	PortName     string      `json:"port_name"`
	VnicType     string      `json:"vnic_type"`
	Accelerate   string      `json:"accelerate"`
	PodName      string      `json:"pod_name"`
	PodNs        string      `json:"pod_ns"`
	FixIP        string      `json:"fix_ip"`
	IPGroupName  string      `json:"ip_group_name"`
	Metadata     interface{} `json:"metadata"`
	Combinable   string      `json:"combinable"`
	Roles        []string    `json:"roles"`
}

type PortLazyAttrForAgent struct {
	//NetworkID      string
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	TenantID   string     `json:"tenant_id"`
	MacAddress string     `json:"mac_address"`
	FixedIps   []ports.IP `json:"fixed_ips"`
	GatewayIP  string     `json:"gateway_ip"`
	Cidr       string     `json:"cidr"`
}

func newPortForAgent(port *services.Port) *PortForAgent {
	p := &PortForAgent{}
	p.EagerAttr.NetworkName = port.EagerAttr.NetworkName
	p.EagerAttr.NetworkPlane = port.EagerAttr.NetworkPlane
	p.EagerAttr.PortName = port.EagerAttr.PortName
	p.EagerAttr.VnicType = port.EagerAttr.VnicType
	p.EagerAttr.Accelerate = port.EagerAttr.Accelerate
	p.EagerAttr.PodName = port.EagerAttr.PodName
	p.EagerAttr.PodNs = port.EagerAttr.PodNs
	p.EagerAttr.FixIP = port.EagerAttr.FixIP
	p.EagerAttr.IPGroupName = port.EagerAttr.IPGroupName
	p.EagerAttr.Metadata = port.EagerAttr.Metadata
	p.EagerAttr.Combinable = port.EagerAttr.Combinable
	p.EagerAttr.Roles = port.EagerAttr.Roles

	p.LazyAttr.ID = port.LazyAttr.ID
	p.LazyAttr.Name = port.LazyAttr.Name
	p.LazyAttr.TenantID = port.LazyAttr.TenantID
	p.LazyAttr.MacAddress = port.LazyAttr.MacAddress
	p.LazyAttr.FixedIps = port.LazyAttr.FixedIps
	p.LazyAttr.GatewayIP = port.LazyAttr.GatewayIP
	p.LazyAttr.Cidr = port.LazyAttr.Cidr
	return p

}
