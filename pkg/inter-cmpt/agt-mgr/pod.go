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

type AgentPodReq struct {
	PodName  string     `json:"pod_name"`
	PodNs    string     `json:"pod_ns"`
	TenantId string     `json:"tenant_id"`
	Ports    []PortInfo `json:"ports"`
}
type PortInfo struct {
	PortId       string `json:"port_id"`
	NetworkName  string `json:"network_name"`
	NetworkPlane string `json:"network_plane"`
	FixIP        string `json:"fix_ip"`
}
