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

package monitor

import "github.com/rackspace/gophercloud/openstack/networking/v2/ports"

type Port struct {
	LazyAttr  PortLazyAttr  `json:"lazy_attr"`
	EagerAttr PortEagerAttr `json:"eager_attr"`
}

type PortEagerAttr struct {
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
}

type PortLazyAttr struct {
	//NetworkID      string
	ports.IP
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	TenantID   string     `json:"tenant_id"`
	MacAddress string     `json:"mac_address"`
	FixedIps   []ports.IP `json:"fixed_ips"`
	GatewayIP  string     `json:"gateway_ip"`
	Cidr       string     `json:"cidr"`
}
