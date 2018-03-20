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

package mgragt

import "github.com/rackspace/gophercloud/openstack/networking/v2/ports"

type CreatePortInfo struct {
	Name       string     `json:"name"`
	NetworkID  string     `json:"network_id"`
	MacAddress string     `json:"mac_address"`
	FixedIps   []ports.IP `json:"fixed_ips"`
	GatewayIP  string     `json:"gateway_ip"`
	Cidr       string     `json:"cidr"`
	PortID     string     `json:"id"`
}

type CreatePortResp struct {
	Port CreatePortInfo `json:"port"`
}

type CreatePortsResp struct {
	Ports []CreatePortInfo `json:"ports"`
}
