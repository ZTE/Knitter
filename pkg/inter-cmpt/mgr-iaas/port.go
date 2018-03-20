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

package mgriaas

import (
	"github.com/ZTE/Knitter/pkg/inter-cmpt/agt-mgr"
)

type MgrPortReq struct {
	agtmgr.AgtPortReq
	TenantId  string `json:"tenant_id"`
	NetworkId string `json:"network_id"`
	SubnetId  string `json:"subnet_id"`
	IPGroupId string `json:"ipgroup_id"`
}

type MgrBulkPortsReq struct {
	TranId string
	Ports  []*MgrPortReq `json:"ports"`
}
