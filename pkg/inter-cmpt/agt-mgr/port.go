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

package agtmgr

type AgtPortReq struct {
	TenantID    string `json:"tenant_id"`
	NetworkName string `json:"network_name"`
	PortName    string `json:"port_name"`
	VnicType    string `json:"vnic_type"` // only used by physical port create-attach, logical port create ignore it
	NodeID      string `json:"node_id"`   // node id which send request
	PodNs       string `json:"pod_ns"`
	PodName     string `json:"pod_name"`
	FixIP       string `json:"ip_addr"`
	ClusterID   string `json:"cluster_id"`
	IPGroupName string `json:"ip_group_name"`
}

type AgtBulkPortsReq struct {
	Ports []AgtPortReq `json:"ports"`
}

type AttachPortReq struct {
	TenantID    string `json:"tenant_id"`
	NetworkName string `json:"network_name"`
	PortName    string `json:"port_name"`
	VnicType    string `json:"vnic_type"`
	FixIP       string `json:"ip_addr"`
	NodeID      string `json:"node_id"` // node id which send request
	ClusterID   string `json:"cluster_id"`
}
