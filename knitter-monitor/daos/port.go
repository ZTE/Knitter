package daos

import "github.com/rackspace/gophercloud/openstack/networking/v2/ports"

type PortDao struct {
}

type PortForDB struct {
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

	ID         string     `json:"id"`
	LazyName   string     `json:"lazy_name"`
	TenantID   string     `json:"tenant_id"`
	MacAddress string     `json:"mac_address"`
	FixedIps   []ports.IP `json:"fixed_ips"`
	GatewayIP  string     `json:"gateway_ip"`
	Cidr       string     `json:"cidr"`
}
