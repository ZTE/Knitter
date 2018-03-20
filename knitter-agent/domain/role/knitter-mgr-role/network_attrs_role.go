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
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ZTE/Knitter/knitter-agent/domain/cni"
	"github.com/ZTE/Knitter/knitter-agent/domain/object/port-obj"
	"github.com/ZTE/Knitter/knitter-agent/err-obj"
	"github.com/ZTE/Knitter/knitter-agent/infra"
	"github.com/ZTE/Knitter/pkg/klog"
)

type NetworkNameReq struct {
	NetworkNames []string `json:"network_names"`
}

type NetworkAttrsRole struct {
}

func (this *NetworkAttrsRole) Get(tenantID string, networkNames []string, needProvider bool) ([]*portobj.NetworkAttrs, error) {
	agtCtx := cni.GetGlobalContext()
	var url = ""
	if !needProvider {
		url = agtCtx.Mc.GetNetworksURL(tenantID)
	} else if needProvider {
		url = agtCtx.Mc.GetNetworksURL(tenantID) + "?provider=true"
	}
	var networkNameReq = &NetworkNameReq{}
	for _, networkName := range networkNames {
		networkNameReq.NetworkNames = append(networkNameReq.NetworkNames, networkName)
	}
	reqJSON, err := json.Marshal(networkNameReq)
	if err != nil {
		klog.Errorf("Marshall http request body: [%v] error: -%v", reqJSON, err)
		return nil, fmt.Errorf("%v:Marshall http request body error", err)
	}
	klog.Infof("Http post url: [%v] body: [%v]", url, string(reqJSON))
	stateCode, networksAttrsByte, err := agtCtx.Mc.PostBytes(url, reqJSON)
	if err != nil {
		klog.Errorf("Get network attrs error: %v", err)
		return nil, err
	}
	if stateCode != 200 {
		klog.Errorf("Get network attrs stateCode: %v", stateCode)
		return nil, errors.New(errobj.GetErrMsg(networksAttrsByte))
	}
	klog.Infof("NetworkAttrsRole:Get networkExtAttrByte: %v", string(networksAttrsByte))
	var networksAttrs = make([]*portobj.NetworkAttrs, 0)
	err = json.Unmarshal(networksAttrsByte, &networksAttrs)
	if err != nil {
		klog.Errorf("NetworkAttrsRole:Get:json.Unmarshal err: %v, networkAttrsJson: %s",
			err, string(networksAttrsByte))
		return nil, infra.ErrJSONUnmarshalFailed
	}
	klog.Infof("NetworkAttrsRole:Get networkAttrs: %v", networksAttrs)
	return networksAttrs, nil
}
