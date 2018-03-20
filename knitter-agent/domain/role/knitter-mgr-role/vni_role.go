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

package knittermgrrole

import (
	"github.com/ZTE/Knitter/knitter-agent/domain/cni"
	"github.com/ZTE/Knitter/knitter-agent/domain/const-value"
	"github.com/ZTE/Knitter/knitter-agent/err-obj"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/antonholmquist/jason"
	"strconv"
)

const DefaultTenant string = constvalue.PaaSTenantAdminDefaultUUID

type VniRole struct {
}

func (this *VniRole) Get(networkID string) (int, error) {
	agtCtx := cni.GetGlobalContext()
	url := agtCtx.Mc.GetSegmentIDURLByID(DefaultTenant, networkID)
	stateCode, vxlanBytes, err := agtCtx.Mc.Get(url)
	if err != nil || stateCode != 200 {
		klog.Errorf("VniRole:Get vxlan info error! %v ", err)
		return -1, errobj.ErrGetVxlanIDFailed
	}
	vxlanJSON, _ := jason.NewObjectFromBytes(vxlanBytes)
	vniStr, err := vxlanJSON.GetString("vni")
	if err != nil {
		klog.Errorf("Get vxlan info error! -%v ", err)
		return -1, errobj.ErrJasonGetStringFailed
	}
	vni, _ := strconv.Atoi(vniStr)
	return vni, nil
}
